// LOADING_SCREEN_CSS styles the static and React startup surfaces before app.css loads.
export const LOADING_SCREEN_CSS = `
.swl-canvas{position:relative;isolation:isolate;display:flex;flex:1 1 0%;flex-direction:column;align-items:center;box-sizing:border-box;width:100%;height:100%;min-height:0;color:#e9edf5}
.swl-canvas *{box-sizing:border-box}
.swl-emblem{display:grid;place-items:center}
.swl-emblem>img{display:block;width:100%;height:100%;object-fit:contain;border-radius:inherit}
.swl-emblem>svg{width:54px;height:54px}
.swl-title{margin:0}
.swl-detail{margin:0;text-align:center;font-size:12px;line-height:18px;overflow-wrap:anywhere;color:#c3b5c5}
.swl-footer a,.swl-action{display:inline-flex;align-items:center;justify-content:center;gap:8px;font:inherit;color:inherit;text-decoration:none;background:transparent;border:0;border-radius:4px;cursor:pointer}
.swl-action svg{height:14px;width:14px}
.swl-footer a:hover,.swl-action:hover{color:#f4f7fc}
.swl-footer a:focus-visible,.swl-action:focus-visible{outline:2px solid var(--app-brand);outline-offset:4px}
.swl-canvas{--scene-background:#141215;--app-background:oklch(0.225 0.002 16.5);--app-card:oklch(0.21 0.003 16.5);--app-brand:oklch(0.82 0.1355 16.5);--color-brand:var(--app-brand);--color-background:var(--app-card);--color-foreground:oklch(1 0 0);--color-foreground-alt:oklch(0.9 0.005 16.5);--focus-height:43%;padding:0;background:var(--app-background);overflow:auto;overflow-x:hidden;font-family:var(--font-display,'Manrope Variable',sans-serif)}
.swl-main{position:static;flex:1;width:100%;height:auto;min-height:0;margin:0;padding:0}
.swl-art{position:absolute;z-index:0;inset:0;width:100%;height:100%;isolation:auto;display:block}
.swl-art::before,.swl-art::after{display:none}
.swl-artwork{position:absolute;display:block;inset:0;width:100%;height:100%;opacity:.86}
.swl-emblem{position:absolute;visibility:visible;z-index:1;left:50%;top:var(--focus-height);width:clamp(88px,13.1vw,128px);height:clamp(88px,13.1vw,128px);border-radius:22%;transform:translate(-50%,-50%);box-shadow:0 8px 24px #0009,0 0 0 1px #efb8e6,0 0 12px #efb09570,0 0 26px #d56bcf55,0 0 46px #9d51b326}
.swl-canvas .swl-console{position:absolute;z-index:2;left:50%;top:calc(var(--focus-height) + clamp(120px,min(14.6vw,26.7vh),156px));transform:translateX(-50%);display:flex;flex-direction:column;align-items:stretch;gap:18px;width:clamp(300px,34vw,380px);padding:0;background:transparent;border:0;box-shadow:none;border-radius:0}
.swl-canvas .swl-head{position:static;width:auto;height:auto;margin:0;overflow:visible;clip-path:none;white-space:normal}.swl-head::before{display:none}.swl-title{font-size:18px;line-height:28px;font-weight:400;letter-spacing:.015em;color:#d7cdd8;text-align:center;text-wrap:pretty}.swl-title-brand{font-weight:560;letter-spacing:-.025em;color:#f5e7f2;text-shadow:0 0 18px #efa6cf24}
.swl-phases{width:100%}.swb-step-label{font-size:10px;line-height:18px}.swb-step-mark{background:transparent}
.swl-phases>.space-y-2{display:flex;flex-direction:column;gap:8px;padding:0;width:max-content;max-width:100%;margin:0 auto}.swl-phases .items-center{display:flex;align-items:center;gap:9px}.swl-phases .size-4,.swl-phases .h-4{flex:none;width:16px;height:16px}.swl-phases .size-3{width:12px;height:12px}.swl-phases .rounded-full{border-radius:50%}.swl-phases .border{border:1px solid #ffffff33}.swl-phases .bg-brand{background:var(--app-brand);color:var(--app-background);justify-content:center}.swl-phases .text-brand{color:var(--app-brand)}.swl-phases .text-xs{font-size:12px;line-height:18px;color:oklch(0.9 0.005 16.5)}.swl-phases .animate-spin{animation:swb-spin 1.4s linear infinite}
.swl-transfer{display:flex;justify-content:space-between;font-size:11px;color:#ffffff9c;margin-top:-2px}.swl-transfer-track{height:3px;overflow:hidden;border-radius:3px;background:#ffffff12;margin-top:-8px}.swl-transfer-track>div{height:100%;background:var(--app-brand)}
.swl-canvas .swl-footer{display:flex;justify-content:center;gap:18px;font-size:11px;color:oklch(0.78 0.008 16.5);margin-top:2px}.swl-action{min-height:32px;padding:4px 8px}.swl-action:focus-visible{outline-color:var(--app-brand)}.swl-action--primary{background:transparent;border:0;color:var(--app-brand)}.swl-error{overflow-wrap:anywhere;font-size:12px;line-height:18px;color:#f1acb2;margin:0;text-align:center}
@media(max-width:600px){.swl-canvas{--focus-height:40%}.swl-emblem{width:88px;height:88px;border-radius:20px}.swl-canvas .swl-console{width:min(320px,calc(100% - 32px));padding:0;top:calc(var(--focus-height) + 140px)}.swl-title{font-size:16px;line-height:24px}}
@media(max-height:520px){.swl-canvas{--focus-height:34%}.swl-emblem{width:68px;height:68px;border-radius:16px}.swl-canvas .swl-console{top:calc(var(--focus-height) + 100px);gap:14px;padding:0}.swl-phases>.space-y-2{gap:5px}}
@media(prefers-reduced-motion:reduce){.swl-phases .animate-spin{animation:none}}
[data-sw-boot-phase-state=pending] .swb-dot{border:1px solid #ffffff40;background:transparent!important}
[data-sw-boot-phase-state=current] .swb-dot{width:14px;height:14px;border:2px solid #efb3c340;border-top-color:var(--app-brand);background:transparent!important;animation:swb-spin .7s linear infinite}
[data-sw-boot-phase-state=complete] .swb-dot{background:var(--app-brand)}
@media(prefers-reduced-motion:reduce){[data-sw-boot-phase-state=current] .swb-dot{animation:none}}
.swl-boot-error-title,[data-sw-boot-state=error] .swl-boot-title{display:none}
[data-sw-boot-state=error] .swl-boot-error-title{display:block}
`.trim()
