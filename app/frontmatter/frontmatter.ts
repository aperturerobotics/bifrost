// FrontmatterValue is one parsed frontmatter value: an integer, a boolean, a
// bracketed list of strings, or the raw text.
export type FrontmatterValue = number | boolean | string | string[]

// FrontmatterField is one `key: value` line of a frontmatter block, with the
// UTF-16 offsets of its key and value in the Markdown file.
export interface FrontmatterField {
  key: string
  value: FrontmatterValue
  keyStart: number
  valueStart: number
  valueEnd: number
}

// Frontmatter is a Markdown file split into its frontmatter fields and body.
export interface Frontmatter {
  // fields lists the block's lines in source order, or nothing without a block.
  fields: FrontmatterField[]
  body: string
}

// readFrontmatter splits a Markdown file into its flat `---` frontmatter block
// and its body. A file without a block has no fields and all of its text as
// body. The frontmatter content mapper reads files with this same function.
export function readFrontmatter(raw: string): Frontmatter {
  const match = /^---\n([\s\S]*?)\n---\n/.exec(raw)
  if (!match) return { fields: [], body: raw }

  // Read one field from each line with a key before its first colon.
  const fields: FrontmatterField[] = []
  let lineStart = match[0].indexOf('\n') + 1
  for (const line of match[1].split('\n')) {
    const colon = line.indexOf(':')
    const key = line.slice(0, colon).trim()
    if (colon !== -1 && key) {
      const text = line.slice(colon + 1).trim()
      const valueStart = lineStart + line.indexOf(text, colon + 1)
      fields.push({
        key,
        value: parseValue(text),
        keyStart: lineStart + line.indexOf(key),
        valueStart,
        valueEnd: valueStart + text.length,
      })
    }
    lineStart += line.length + 1
  }

  return { fields, body: raw.slice(match[0].length) }
}

// frontmatterRecord collects a block's fields into one record by key.
export function frontmatterRecord(
  fm: Frontmatter,
): Record<string, FrontmatterValue> {
  return Object.fromEntries(fm.fields.map((field) => [field.key, field.value]))
}

// parseValue reads digits as an integer, true and false as booleans, and
// [a, b] as a list of strings, and keeps any other text as written.
function parseValue(text: string): FrontmatterValue {
  if (/^\d+$/.test(text)) return parseInt(text, 10)
  if (text === 'true') return true
  if (text === 'false') return false
  if (text.startsWith('[') && text.endsWith(']')) {
    return text
      .slice(1, -1)
      .split(',')
      .flatMap((part) => (part.trim() ? [part.trim()] : []))
  }
  return text
}
