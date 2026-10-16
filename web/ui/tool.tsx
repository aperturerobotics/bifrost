import * as React from 'react'
import { LuTerminal } from 'react-icons/lu'

import { CodeBlock, CodeBlockContent } from '@s4wave/web/ui/code-block.js'
import {
  Trace,
  TraceCollapsible,
  TraceCollapsibleContent,
  TraceCollapsibleTrigger,
  TraceIcon,
  TraceLine,
  TraceTitle,
  type TraceStatus,
} from '@s4wave/web/ui/trace.js'
import { statusTitle } from '@s4wave/web/ui/trace-status.js'

/** Tool renders a tool call as a trace line, opening once it settles with text. */
function Tool({
  name,
  status,
  input,
  output,
  className,
  ...props
}: React.ComponentProps<'div'> & {
  name: string
  status: TraceStatus
  input?: string
  output?: string
}) {
  const expandable =
    (status === 'done' || status === 'error') && Boolean(input || output)
  const line = (
    <TraceLine status={status}>
      <TraceIcon>
        <LuTerminal />
      </TraceIcon>
      <TraceTitle>{statusTitle(status, name)}</TraceTitle>
    </TraceLine>
  )

  return (
    <div data-slot="tool" className={className} {...props}>
      <Trace>
        {expandable ? (
          <TraceCollapsible>
            <TraceCollapsibleTrigger>{line}</TraceCollapsibleTrigger>
            <TraceCollapsibleContent className="flex flex-col gap-2">
              {input ? (
                <CodeBlock size="sm">
                  <CodeBlockContent wrap className="max-h-64">
                    {input}
                  </CodeBlockContent>
                </CodeBlock>
              ) : null}
              {output ? (
                <CodeBlock
                  size="sm"
                  variant={status === 'error' ? 'destructive' : 'default'}
                >
                  <CodeBlockContent wrap className="max-h-64">
                    {output}
                  </CodeBlockContent>
                </CodeBlock>
              ) : null}
            </TraceCollapsibleContent>
          </TraceCollapsible>
        ) : (
          line
        )}
      </Trace>
    </div>
  )
}

export { Tool }
