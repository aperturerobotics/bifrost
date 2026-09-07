import { LoadingScreen as BaseLoadingScreen } from '@s4wave/web/ui/loading/LoadingScreen.js'

import { PhaseChecklist } from '@s4wave/app/session/setup/PhaseChecklist.js'

import type {
  QuickstartProgressState,
  QuickstartProgressStep,
} from './create.js'

const progressLabels: Array<{ step: QuickstartProgressStep; label: string }> = [
  { step: 'session', label: 'Local session' },
  { step: 'space', label: 'New Space' },
  { step: 'frame', label: 'Open workspace' },
  { step: 'content', label: 'Add starter content' },
]

function getProgressLabels(quickstartId: string) {
  if (quickstartId === 'local') return progressLabels.slice(0, 1)
  return progressLabels
}

function buildQuickstartProgress(
  quickstartId: string,
): QuickstartProgressState {
  const labels = getProgressLabels(quickstartId)
  return {
    step: 'session',
    stepIndex: 1,
    stepCount: labels.length,
    detail: `Setting up ${quickstartId}`,
  }
}

// LoadingScreen presents the quickstart's current setup phase. Phase boundaries
// stay separate from measured download progress.
export function LoadingScreen({
  quickstartId,
  progress,
}: {
  quickstartId: string
  progress?: QuickstartProgressState | null
}) {
  const current = progress ?? buildQuickstartProgress(quickstartId)
  const labels = getProgressLabels(quickstartId)
  const activeIndex = Math.max(
    0,
    labels.findIndex((item) => item.step === current.step),
  )
  const phases = labels.map((item, index) => ({
    label: item.label,
    done: index < activeIndex,
    active: index === activeIndex,
  }))

  return (
    <BaseLoadingScreen
      view={{
        state: 'active',
        title:
          quickstartId === 'local'
            ? 'Opening your workspace'
            : 'Creating your space',
        detail: current.detail,
      }}
    >
      <PhaseChecklist phases={phases} />
    </BaseLoadingScreen>
  )
}
