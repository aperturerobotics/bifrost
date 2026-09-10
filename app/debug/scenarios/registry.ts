import { canvasFiles, imageContext } from './canvas-files.js'
import type { AppScenario } from './ScenarioPage.js'

export const appScenarios: AppScenario[] = [
  { name: 'canvas-files', title: 'Canvas and Files', setup: canvasFiles },
  {
    name: 'image-context',
    title: 'Image and Space context',
    setup: imageContext,
  },
]
