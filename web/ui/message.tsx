import * as React from 'react'

import { cn } from '@s4wave/web/style/utils.js'

/** MessageFrom is who sent the turn. */
export type MessageFrom = 'user' | 'assistant'

/** Message renders one turn. The person's text sits in a bubble on the right. */
function Message({
  from,
  text,
  actions,
  className,
  ...props
}: React.ComponentProps<'div'> & {
  from: MessageFrom
  text: string
  actions?: React.ReactNode
}) {
  const alignEnd = from === 'user'

  return (
    <div
      data-slot="message"
      className={cn(
        'flex w-full min-w-0 flex-col gap-2 text-sm',
        alignEnd && 'items-end',
        className,
      )}
      {...props}
    >
      {alignEnd ? (
        <div
          data-slot="message-bubble"
          className="bg-muted text-foreground max-w-full rounded-2xl px-3 py-2 wrap-break-word"
        >
          {text}
        </div>
      ) : (
        <div
          data-slot="message-content"
          className="flex w-full min-w-0 flex-col gap-2 wrap-break-word"
        >
          <p className="whitespace-pre-wrap">{text}</p>
        </div>
      )}
      {actions ? (
        <div
          data-slot="message-footer"
          className={cn('flex items-center gap-1', alignEnd && 'justify-end')}
        >
          {actions}
        </div>
      ) : null}
    </div>
  )
}

export { Message }
