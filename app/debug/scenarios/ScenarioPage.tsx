import { useId, useState } from 'react'
import {
  SpacewaveApp,
  type SpacewaveAppProps,
} from '@s4wave/app/SpacewaveApp.js'
import { getDebugContext } from '@s4wave/sdk/debug/context.js'

export interface AppScenario {
  name: string
  title: string
  setup: NonNullable<SpacewaveAppProps['setup']>
}

const ephemeralStorage = { ephemeral: true } as const

// ScenarioPage keeps the outer route stable while each app owns its World.
export function ScenarioPage({ scenario }: { scenario: AppScenario }) {
  const [compare, setCompare] = useState(false)
  const [resetError, setResetError] = useState('')
  const instanceId = useId()
  const appId = `scenario:${scenario.name}:${instanceId}`
  const reset = async () => {
    try {
      setResetError('')
      await getDebugContext<{ reset(): Promise<void> }>(appId).reset()
    } catch (error) {
      setResetError(error instanceof Error ? error.message : String(error))
    }
  }
  return (
    <div
      className="flex h-full min-h-0 flex-col"
      data-testid="app-scenario"
      data-app-id={appId}
    >
      <div className="flex shrink-0 items-center gap-4 border-b px-3 py-2 text-xs">
        <strong>{scenario.title}</strong>
        <button
          type="button"
          className="rounded border px-2 py-1"
          onClick={() => {
            void reset()
          }}
        >
          Reset scenario
        </button>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={compare}
            onChange={(event) => setCompare(event.target.checked)}
          />
          Compare two apps
        </label>
        {resetError && <span role="alert">{resetError}</span>}
      </div>
      <div className="flex min-h-0 flex-1 divide-x">
        <SpacewaveApp
          key={appId}
          appId={appId}
          suppliedStorage={ephemeralStorage}
          setup={scenario.setup}
        />
        {compare && (
          <SpacewaveApp
            key={`${appId}:comparison`}
            appId={`${appId}:comparison`}
            suppliedStorage={ephemeralStorage}
            setup={scenario.setup}
          />
        )}
      </div>
    </div>
  )
}
