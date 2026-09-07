import * as THREE from 'three'
import { EffectComposer } from 'three/addons/postprocessing/EffectComposer.js'
import { RenderPass } from 'three/addons/postprocessing/RenderPass.js'
import { ShaderPass } from 'three/addons/postprocessing/ShaderPass.js'
import { UnrealBloomPass } from 'three/addons/postprocessing/UnrealBloomPass.js'
import { OutputPass } from 'three/addons/postprocessing/OutputPass.js'

// createLoadingArtwork owns the decorative GPU scene until the surface unmounts.
// Startup state remains in the caller; DOM geometry only locates the light.
export function createLoadingArtwork(
  canvas: HTMLCanvasElement,
  shell: HTMLElement,
): () => void {
  const emblem = shell.querySelector<HTMLElement>('.swl-emblem')!
  const reduced = matchMedia('(prefers-reduced-motion: reduce)')
  const renderer = new THREE.WebGLRenderer({
    canvas,
    alpha: true,
    antialias: true,
    depth: false,
    powerPreference: 'low-power',
  })
  renderer.setClearColor(0x000000, 1)
  renderer.setPixelRatio(1)
  renderer.toneMapping = THREE.ACESFilmicToneMapping
  const scene = new THREE.Scene()
  const camera = new THREE.Camera()

  // Bloom receives linear HDR energy; the unblurred image remains at native density.
  const target = new THREE.WebGLRenderTarget(1, 1, {
    type: THREE.HalfFloatType,
    depthBuffer: false,
    samples: Math.min(4, renderer.capabilities.maxSamples),
  })
  const composer = new EffectComposer(renderer, target)
  const renderPass = new RenderPass(scene, camera)
  const bloom = new UnrealBloomPass(new THREE.Vector2(1, 1), 0.18, 0.08, 1.4)
  const output = new OutputPass()

  // Sky drift shares the scene clock and stops with reduced motion or a hidden page.
  const probe = document.createElement('canvas').getContext('2d')!
  probe.fillStyle =
    getComputedStyle(shell).getPropertyValue('--scene-background')
  probe.fillRect(0, 0, 1, 1)
  const background = probe.getImageData(0, 0, 1, 1).data
  const composite = new ShaderPass({
    uniforms: {
      tDiffuse: { value: null },
      background: {
        value: new THREE.Vector3(
          background[0],
          background[1],
          background[2],
        ).divideScalar(255),
      },
      aspect: { value: 1 },
      time: { value: 12 },
      focus: { value: new THREE.Vector2() },
    },
    vertexShader:
      'varying vec2 vUv;void main(){vUv=uv;gl_Position=vec4(position.xy,0.0,1.0);}',
    fragmentShader: `
    uniform sampler2D tDiffuse;
    uniform vec3 background;
    uniform float aspect,time;
    uniform vec2 focus;
    varying vec2 vUv;
    float hash(vec2 p){
      vec3 q=fract(vec3(p.xyx)*0.1031);
      q+=dot(q,q.yzx+33.33);
      return fract((q.x+q.y)*q.z);
    }
    float noise(vec2 p){
      vec2 cell=floor(p),f=fract(p);
      f=f*f*(3.0-2.0*f);
      return mix(mix(hash(cell),hash(cell+vec2(1,0)),f.x),mix(hash(cell+vec2(0,1)),hash(cell+1.0),f.x),f.y);
    }
    float dustNoise(vec2 p){
      return noise(p)*0.50+noise(p*2.13+4.7)*0.27+noise(p*4.67+9.2)*0.15+noise(p*9.73+13.8)*0.08;
    }
    void main(){
      vec3 light=texture2D(tDiffuse,vUv).rgb;
      vec2 p=(vUv-0.5)*vec2(aspect,1.0);
      vec2 drift=vec2(time*0.0015,-time*0.0007);
      vec2 warp=vec2(noise(p*3.0+7.0+drift),noise(p*3.0+19.0-drift))-0.5;
      float dust=dustNoise(p*vec2(5.0,8.0)+warp*2.7+drift);
      float band=p.y+p.x*0.58+0.05*sin(p.x*3.0);
      float envelope=exp(-band*band/0.045)*smoothstep(0.22,0.70,dust);
      vec2 curl=vec2(warp.y,-warp.x);
      float wisps=dustNoise(p*vec2(20.0,32.0)+curl*6.0+drift*1.7);
      float behind=dustNoise(p*vec2(11.0,17.0)-curl*3.0+8.0-drift);
      float hollow=1.0-exp(-dot(p-focus,p-focus)/0.055)*0.75;
      float density=envelope*mix(wisps,behind,0.3)*hollow;

      // Approximate extinction and scattered light through two drifting dust layers.
      float transmission=exp(-density*1.8);
      float scatter=1.0-transmission;
      float backlight=0.65+0.35*exp(-dot(p-focus,p-focus)/0.6);
      float filaments=pow(max(0.0,1.0-abs(wisps-0.53)*8.0),4.0);
      float edgeLight=filaments*envelope*transmission*hollow;
      float vignette=0.82+0.18*exp(-dot(p,p)/0.5);
      vec3 sky=background*0.9*vignette*(0.96+transmission*0.04);
      sky+=vec3(0.105,0.05,0.145)*scatter*backlight;
      sky+=vec3(0.10,0.055,0.15)*edgeLight*0.22;
      sky+=(hash(gl_FragCoord.xy)-0.5)*0.003;
      gl_FragColor=vec4(sky+(1.0-sky)*light,1.0);
    }`,
  })
  composer.addPass(renderPass)
  composer.addPass(bloom)
  composer.addPass(output)
  composer.addPass(composite)

  const uniforms = {
    time: composite.uniforms.time,
    aspect: { value: 1 },
    height: { value: 800 },
    pointer: { value: new THREE.Vector2() },
    pointerStrength: { value: 0 },
    origin: { value: new THREE.Vector2() },
    focus: { value: new THREE.Vector2() },
    quietArea: { value: new THREE.Vector4() },
    coreRadius: { value: 0.12 },
  }
  const sparkCount = 40
  const starCount = 1000
  const count = sparkCount + starCount
  const data = new Float32Array(count * 4)
  let seed = 988381
  function random() {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0
    return seed / 4294967296
  }
  for (let i = 0; i < count; i++) {
    const kind = i < sparkCount ? 0 : 2
    data.set([random(), random(), random(), kind + random() * 0.99], i * 4)
  }
  const geometry = new THREE.InstancedBufferGeometry()
  geometry.setAttribute(
    'position',
    new THREE.Float32BufferAttribute(
      [-1, -1, 0, 1, -1, 0, -1, 1, 0, -1, 1, 0, 1, -1, 0, 1, 1, 0],
      3,
    ),
  )
  geometry.setAttribute('seed', new THREE.InstancedBufferAttribute(data, 4))
  geometry.instanceCount = count
  // Both moving sparks and continuous ribbons follow the same bow-shaped lanes.
  const laneCount = 24
  const flowPath = `
  vec2 currentPath(float progress,vec4 s){
    float index=floor(s.y*${laneCount}.0);
    float variation=fract(sin(index*31.17+4.0)*1743.13);
    float lane=(index+0.2+variation*0.6)/${laneCount / 2}.0-1.0;
    float x=(progress*2.0-1.0)*(aspect+0.3);
    float dx=x-origin.x;
    float spread=coreRadius*2.6+0.14+abs(lane)*0.03;
    float bow=exp(-dx*dx/(spread*spread));
    float edge=clamp(abs(dx)/aspect,0.0,1.0);
    float width=coreRadius*(0.34+1.65*edge*edge*edge)+0.005;
    float bandWidth=mix(width,coreRadius*0.55,bow);
    float y=origin.y+lane*bandWidth;
    y+=sign(lane)*(coreRadius*1.58+0.01)*bow;
    y+=lane*sin(dx*1.7+lane*3.0)*0.018*(1.0-bow);
    y+=sin(dx*2.0+s.y*3.0+time*0.06)*0.003;
    if(lane<0.0)y=origin.y+(y-origin.y)*0.82;
    return vec2(x,y);
  }
`
  const material = new THREE.ShaderMaterial({
    uniforms,
    transparent: true,
    depthWrite: false,
    depthTest: false,
    blending: THREE.AdditiveBlending,
    vertexShader: `
    attribute vec4 seed;
    uniform float time,aspect,height,coreRadius;
    uniform vec2 origin,focus;
    uniform vec4 quietArea;
    varying vec2 vUv;
    varying vec3 color;
    varying float opacity,heat,star;
    ${flowPath}
    vec2 path(float progress,vec4 s){
      float kind=floor(s.w);
      if(kind>1.5){
        return vec2((s.x*2.0-1.0)*aspect,(s.y*2.0-1.0));
      }
      bool pulse=fract(s.w)<0.25;
      if(!pulse)return currentPath(progress,s);
      {
        float angle=s.y*6.283185;
        float r=progress*2.7;
        vec2 ripple=vec2(cos(angle)*1.8,sin(angle)*0.72)*r;
        ripple.y+=sin(r*3.0+time*0.12)*0.025;
        return focus+ripple;
      }

    }
    void main(){
      vUv=position.xy;
      float kind=floor(seed.w);
      star=step(1.5,kind);
      bool pulse=fract(seed.w)<0.25;
      float speed=pulse?0.075:0.032;
      float progress=fract(seed.x+time*speed*(0.45+seed.z*0.85));
      float travel=progress+0.045*sin((progress-0.5)*6.283185);
      vec2 p=path(travel,seed);
      vec2 next=path(travel+0.001,seed);
      vec2 tangent=kind>1.5?vec2(1.0,0.0):normalize(next-p);
      vec2 normal=vec2(-tangent.y,tangent.x);
      heat=pow(seed.z,7.0);
      float length=(1.5+heat*8.0)*2.0/height;
      float width=(0.45+heat*0.15)*2.0/height;
      float fade=sin(progress*3.14159);
      opacity=(0.045+heat*0.5)*fade;
      if(pulse&&kind<0.5){opacity*=pow(1.0-progress,1.5)*0.8;}
      float hue=fract(seed.x*17.0+seed.y*31.0);
      color=mix(vec3(0.46,0.075,0.30),vec3(0.85,0.26,0.09),smoothstep(0.7,0.85,hue));
      color*=0.25+heat*10.0;
      if(kind>1.5){
        heat=smoothstep(0.98,1.0,seed.z);
        width=(0.3+pow(seed.z,8.0)*0.5+heat*2.8)*2.0/height;
        length=width*2.0;
        opacity=0.035+pow(seed.z,8.0)*0.5;
        vec3 starLight=mix(vec3(0.10,0.32,1.0),vec3(1.0,0.40,0.13),step(0.55,seed.x));
        color=mix(vec3(0.28,0.30,0.40),starLight,heat)*(1.0+heat*5.0);
      }
      vec2 distance=abs(p-quietArea.xy)-quietArea.zw;
      opacity*=smoothstep(0.0,0.06,max(distance.x,distance.y));
      p+=tangent*vUv.x*length+normal*vUv.y*width*2.0;
      gl_Position=vec4(p.x/aspect,p.y,0.0,1.0);
    }`,
    fragmentShader: `
    varying vec2 vUv;varying vec3 color;varying float opacity,heat,star;
    void main(){
      if(star>0.5){
        float radius=length(vUv);
        float edge=max(fwidth(radius),0.001);
        float size=0.5/(1.0+heat*3.0);
        float core=1.0-smoothstep(size-edge*0.5,size+edge*0.5,radius);
        vec2 axis=abs(vUv);
        float spikes=(exp(-axis.x*45.0-axis.y*4.0)+exp(-axis.y*45.0-axis.x*4.0))*heat*0.8;
        gl_FragColor=vec4(color,(core+spikes)*opacity);
        return;
      }
      float edge=max(fwidth(vUv.y),0.001);
      float core=1.0-smoothstep(0.5-edge*0.5,0.5+edge*0.5,abs(vUv.y));
      float tail=smoothstep(-1.0,0.2,vUv.x)*(1.0-smoothstep(0.75,1.0,vUv.x));
      float head=smoothstep(0.45,0.7,vUv.x)*(1.0-smoothstep(0.75,1.0,vUv.x));
      vec3 radiance=color*tail+vec3(1.0,0.83,0.74)*head*heat*2.0;
      gl_FragColor=vec4(radiance,core*opacity);
    }`,
  })
  const energy = new THREE.Mesh(geometry, material)
  energy.frustumCulled = false
  scene.add(energy)

  // Thin continuous ribbons establish the flow; traveling highlights supply motion.
  const ribbonVertices: number[] = []
  const segments = 256
  for (let segment = 0; segment < segments; segment++) {
    const start = segment / segments
    const end = (segment + 1) / segments
    ribbonVertices.push(
      start,
      -1,
      0,
      end,
      -1,
      0,
      start,
      1,
      0,
      start,
      1,
      0,
      end,
      -1,
      0,
      end,
      1,
      0,
    )
  }
  const ribbonData = new Float32Array(laneCount * 4)
  for (let lane = 0; lane < laneCount; lane++) {
    ribbonData.set(
      [random(), (lane + 0.5) / laneCount, random(), random()],
      lane * 4,
    )
  }
  const ribbonGeometry = new THREE.InstancedBufferGeometry()
  ribbonGeometry.setAttribute(
    'position',
    new THREE.Float32BufferAttribute(ribbonVertices, 3),
  )
  ribbonGeometry.setAttribute(
    'seed',
    new THREE.InstancedBufferAttribute(ribbonData, 4),
  )
  ribbonGeometry.instanceCount = laneCount
  const ribbonMaterial = new THREE.ShaderMaterial({
    uniforms,
    transparent: true,
    depthWrite: false,
    depthTest: false,
    blending: THREE.AdditiveBlending,
    vertexShader: `
    attribute vec4 seed;
    uniform float time,aspect,height,coreRadius;
    uniform vec2 origin;
    varying vec2 vUv;
    varying vec2 surfacePoint;
    varying vec2 surfaceTangent;
    varying vec3 color;
    varying vec4 strand;
    ${flowPath}
    void main(){
      vUv=position.xy;
      vec2 p=currentPath(position.x,seed);
      vec2 next=currentPath(position.x+0.001,seed);
      surfacePoint=p;
      vec2 tangent=normalize(next-p);
      surfaceTangent=tangent;
      float lane=floor(seed.y*${laneCount}.0);
      float primary=1.0-step(0.5,mod(lane-1.0,6.0));
      float depth=mod(lane,3.0)/2.0;
      p+=vec2(-tangent.y,tangent.x)*position.y*(0.65+depth*0.5+primary*1.6)*4.0/height;
      float hue=clamp(abs((lane+0.5)/${laneCount / 2}.0-1.0)+sin(lane*2.3)*0.12,0.0,1.0);
      float accent=(lane-1.0)/6.0;
      color=mix(vec3(0.05,0.13,0.32),vec3(0.64,0.045,0.36),smoothstep(0.2,0.8,hue));
      if(primary>0.5)color=mix(color,vec3(0.64,0.045,0.36),0.5);
      float phase=primary>0.5?fract(accent*0.211+0.353):seed.x;
      strand=vec4(primary,phase,lane,depth);
      gl_Position=vec4(p.x/aspect,p.y,0.0,1.0);
    }`,
    fragmentShader: `
    varying vec2 vUv;
    varying vec2 surfacePoint;
    varying vec2 surfaceTangent;
    varying vec3 color;
    uniform float time,height,pointerStrength;
    uniform vec2 pointer;
    varying vec4 strand;
    void main(){
      float edge=max(fwidth(vUv.y),0.001);
      float core=1.0-smoothstep(0.125-edge*0.5,0.125+edge*0.5,abs(vUv.y));
      float aura=exp(-vUv.y*vUv.y*18.0)*0.30;
      // Evaluate light per pixel so the traveling crest never exposes mesh segments.
      float angle=(vUv.x*6.283185-time*0.22+strand.y*6.28)*0.5;
      float phase=sin(angle);
      float companion=sin(angle+1.7);
      float crest=exp(-phase*phase*520.0)+exp(-companion*companion*900.0)*0.4;
      float glow=exp(-phase*phase*80.0)+exp(-companion*companion*140.0)*0.4;
      float depth=mix(0.065,0.19,strand.w);
      float light=(depth+strand.x*0.30)*sin(vUv.x*3.14159);
      float amber=mod(floor((strand.z-1.0)/6.0),2.0);
      vec3 accent=mix(vec3(0.85,0.035,0.52),vec3(1.0,0.28,0.055),amber);
      vec3 highlight=mix(vec3(1.0,0.65,0.88),vec3(1.0,0.75,0.48),amber);
      float goldSpan=exp(-phase*phase*16.0)+exp(-companion*companion*30.0)*0.55;
      vec3 gold=vec3(1.0,0.34,0.065);
      vec3 radiance=mix(color,gold,min(goldSpan,1.0)*amber*strand.x*0.85)*light;
      radiance+=(accent*(crest*7.5+glow*1.6)+highlight*pow(crest,4.0))*strand.x*sin(vUv.x*3.14159);
      radiance+=gold*goldSpan*amber*strand.x*0.85*sin(vUv.x*3.14159);
      // Iridescence follows the strand's local angle beneath the pointer light.
      // Geometry and the ambient animation remain independent of pointer input.
      vec2 lightOffset=(surfacePoint-pointer)*height*0.5;
      vec2 tangent=normalize(surfaceTangent);
      float along=dot(lightOffset,tangent);
      float across=dot(lightOffset,vec2(-tangent.y,tangent.x));
      float illumination=exp(-along*along/(140.0*140.0)-across*across/(85.0*85.0))*pointerStrength;
      float hue=clamp(0.5+along/210.0+across/170.0+sin(strand.z*0.7)*0.12,0.0,1.0);
      vec3 iridescence=mix(vec3(0.22,0.10,1.0),vec3(0.95,0.08,0.46),smoothstep(0.08,0.60,hue));
      iridescence=mix(iridescence,vec3(1.0,0.55,0.16),smoothstep(0.62,0.98,hue));
      float reflected=exp(-along*along/(22.0*22.0));
      float energy=max(radiance.r,max(radiance.g,radiance.b));
      radiance=mix(radiance,iridescence*(energy+0.22),illumination*0.72);
      radiance+=iridescence*illumination*(0.14+reflected*0.42);
      radiance*=0.7;
      gl_FragColor=vec4(radiance*(core+aura),1.0);
    }`,
  })
  const ribbons = new THREE.Mesh(ribbonGeometry, ribbonMaterial)
  ribbons.frustumCulled = false
  scene.add(ribbons)

  function center(element: Element, target: THREE.Vector2) {
    const rect = element.getBoundingClientRect()
    const height = canvas.clientHeight
    const bounds = canvas.getBoundingClientRect()
    target.set(
      ((rect.x - bounds.x + rect.width / 2 - canvas.clientWidth / 2) * 2) /
        height,
      1 - ((rect.y - bounds.y + rect.height / 2) * 2) / height,
    )
  }
  function layout() {
    const width = canvas.clientWidth
    const height = canvas.clientHeight
    if (!width || !height) return
    const limit = renderer.capabilities.maxTextureSize
    const scale = Math.min(devicePixelRatio, limit / width, limit / height)
    renderer.setSize(
      Math.round(width * scale),
      Math.round(height * scale),
      false,
    )
    composer.setSize(Math.round(width * scale), Math.round(height * scale))
    uniforms.aspect.value = width / height
    composite.uniforms.aspect.value = width / height
    geometry.instanceCount =
      sparkCount +
      Math.round(starCount * Math.min(1, (width * height) / 1050000))
    uniforms.height.value = height
    center(emblem, uniforms.origin.value)
    composite.uniforms.focus.value
      .copy(uniforms.origin.value)
      .multiplyScalar(0.5)
    const consoleBounds = shell
      .querySelector<HTMLElement>('.swl-console')!
      .getBoundingClientRect()
    const bounds = canvas.getBoundingClientRect()
    uniforms.quietArea.value.set(
      ((consoleBounds.x - bounds.x + consoleBounds.width / 2 - width / 2) * 2) /
        height,
      1 -
        ((consoleBounds.y - bounds.y + consoleBounds.height / 2) * 2) / height,
      consoleBounds.width / height,
      consoleBounds.height / height,
    )
    uniforms.coreRadius.value = emblem.getBoundingClientRect().width / height
    const active = shell.querySelector(
      '.swb-step[data-state="current"] .swb-step-mark, [data-sw-boot-phase-state="current"], .swl-phases .animate-spin',
    )
    center(active || emblem, uniforms.focus.value)
  }
  let request = 0
  let previous = 0
  let elapsed = 12
  const pointerTarget = new THREE.Vector2()
  let pointerInside = false

  function movePointer(event: PointerEvent) {
    if (event.pointerType === 'touch' || reduced.matches) return
    const rect = canvas.getBoundingClientRect()
    pointerTarget.set(
      ((event.clientX - rect.left - rect.width / 2) * 2) / rect.height,
      1 - ((event.clientY - rect.top) * 2) / rect.height,
    )
    pointerInside = true
    resume()
  }

  function leavePointer() {
    pointerInside = false
  }

  function draw(now: number) {
    request = 0
    if (document.hidden) return
    const dt = previous ? Math.min((now - previous) / 1000, 0.05) : 0
    previous = now
    if (!reduced.matches) elapsed += dt
    uniforms.time.value = elapsed
    uniforms.pointer.value.copy(pointerTarget)
    const strength = pointerInside && !reduced.matches ? 1 : 0
    uniforms.pointerStrength.value = reduced.matches
      ? 0
      : THREE.MathUtils.lerp(
          uniforms.pointerStrength.value,
          strength,
          1 - Math.exp(-dt * 4),
        )
    composer.render(dt)
    if (!reduced.matches) request = requestAnimationFrame(draw)
  }
  function resume() {
    if (!request) request = requestAnimationFrame(draw)
  }
  let density = matchMedia(`(resolution: ${devicePixelRatio}dppx)`)
  function densityChanged() {
    density.removeEventListener('change', densityChanged)
    density = matchMedia(`(resolution: ${devicePixelRatio}dppx)`)
    density.addEventListener('change', densityChanged)
    layout()
    resume()
  }
  density.addEventListener('change', densityChanged)
  const observer = new ResizeObserver(() => {
    layout()
    resume()
  })
  observer.observe(canvas)
  const stages = new MutationObserver(() => {
    layout()
    resume()
  })
  stages.observe(shell, {
    subtree: true,
    childList: true,
    attributes: true,
    attributeFilter: ['data-state', 'data-sw-boot-phase-state', 'class'],
  })
  shell.addEventListener('pointermove', movePointer)
  shell.addEventListener('pointerleave', leavePointer)
  function resetClock() {
    previous = 0
    resume()
  }
  document.addEventListener('visibilitychange', resetClock)
  reduced.addEventListener('change', resetClock)
  layout()
  canvas.dataset.renderer = 'three'
  resume()
  return () => {
    document.removeEventListener('visibilitychange', resetClock)
    reduced.removeEventListener('change', resetClock)
    cancelAnimationFrame(request)
    density.removeEventListener('change', densityChanged)
    observer.disconnect()
    stages.disconnect()
    shell.removeEventListener('pointermove', movePointer)
    shell.removeEventListener('pointerleave', leavePointer)
    geometry.dispose()
    material.dispose()
    ribbonGeometry.dispose()
    ribbonMaterial.dispose()
    renderPass.dispose()
    bloom.dispose()
    output.dispose()
    composite.dispose()
    composer.dispose()
    renderer.dispose()
    delete canvas.dataset.renderer
  }
}
