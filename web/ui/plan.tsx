import * as React from 'react'
import { LuRoute } from 'react-icons/lu'

import {
  Trace,
  TraceCollapsible,
  TraceCollapsibleContent,
  TraceCollapsibleTrigger,
  TraceDescription,
  TraceIcon,
  TraceLine,
  TraceTitle,
  type TraceStatus,
} from '@s4wave/web/ui/trace.js'
import { statusLabel, stepIcon } from '@s4wave/web/ui/trace-status.js'

/** PlanStep is one step in the plan. id is the caller's identity for the step. */
export type PlanStep = {
  id: string
  title: string
  status: TraceStatus
}

/** planHeaderStatus is running, else error, else pending, else done. */
function planHeaderStatus(steps: PlanStep[]): TraceStatus {
  if (steps.some((step) => step.status === 'running')) {
    return 'running'
  }
  if (steps.some((step) => step.status === 'error')) {
    return 'error'
  }
  if (steps.some((step) => step.status === 'pending')) {
    return 'pending'
  }
  return 'done'
}

/** Plan renders the steps the agent decided on, and how many are done. */
function Plan({
  title,
  description,
  steps,
  className,
  ...props
}: React.ComponentProps<'div'> & {
  title: string
  description?: string
  steps: PlanStep[]
}) {
  const done = steps.filter((step) => step.status === 'done').length

  return (
    <div data-slot="plan" className={className} {...props}>
      <Trace>
        <TraceCollapsible defaultOpen>
          <TraceCollapsibleTrigger>
            <TraceLine status={planHeaderStatus(steps)}>
              <TraceIcon>
                <LuRoute />
              </TraceIcon>
              <TraceTitle>{title}</TraceTitle>
              <TraceDescription>
                {done} of {steps.length} done
              </TraceDescription>
            </TraceLine>
          </TraceCollapsibleTrigger>
          <TraceCollapsibleContent className="flex flex-col gap-3">
            {description ? <p>{description}</p> : null}
            <Trace>
              {steps.map((step) => (
                <TraceLine key={step.id} status={step.status}>
                  <TraceIcon>{stepIcon(step.status)}</TraceIcon>
                  <TraceTitle>{step.title}</TraceTitle>
                  <span className="sr-only">{statusLabel(step.status)}</span>
                </TraceLine>
              ))}
            </Trace>
          </TraceCollapsibleContent>
        </TraceCollapsible>
      </Trace>
    </div>
  )
}

export { Plan }
