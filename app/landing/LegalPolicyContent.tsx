import Markdown, { type MarkdownToJSX } from 'markdown-to-jsx'

import { MarkdownLink } from '@s4wave/app/docs/MarkdownLink.js'

const markdownOptions: MarkdownToJSX.Options = {
  overrides: {
    a: {
      component: MarkdownLink,
      props: { className: 'text-brand hover:text-brand-highlight underline' },
    },
    p: { props: { className: 'mb-3 last:mb-0' } },
  },
}

interface LegalPolicyContentProps {
  sections: readonly { title: string; content: string }[]
}

// LegalPolicyContent renders the canonical policy sections with routable links.
export function LegalPolicyContent({ sections }: LegalPolicyContentProps) {
  return (
    <section className="relative z-10 mx-auto w-full max-w-4xl px-4 pb-14 @lg:px-8 @lg:pb-16">
      <div className="space-y-6">
        {sections.map(({ title, content }) => (
          <section
            key={title}
            className="border-foreground/8 bg-background-card/50 rounded-lg border p-6 backdrop-blur-sm @lg:p-8"
          >
            <h2 className="text-foreground mb-4 text-lg font-semibold">
              {title}
            </h2>
            <div className="text-foreground-alt text-sm leading-relaxed">
              <Markdown options={markdownOptions}>{content}</Markdown>
            </div>
          </section>
        ))}
      </div>
    </section>
  )
}
