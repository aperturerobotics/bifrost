import { defineSchema } from 'spacewave'
import type { StandardSchemaV1 } from '@standard-schema/spec'

export interface Todo {
  title: string
  done: boolean
}

function validator<T>(
  accepts: (value: unknown) => value is T,
): StandardSchemaV1<T> {
  return {
    '~standard': {
      version: 1,
      vendor: 'sync-acceptance',
      validate: (value) =>
        accepts(value) ? { value } : { issues: [{ message: 'Invalid value' }] },
    },
  }
}

const todo = validator<Todo>((value): value is Todo => {
  if (value === null || typeof value !== 'object') return false
  return (
    'title' in value &&
    typeof value.title === 'string' &&
    'done' in value &&
    typeof value.done === 'boolean'
  )
})
const number = validator<number>(
  (value): value is number => typeof value === 'number',
)
const empty = validator<null>((value): value is null => value === null)

export const schema = defineSchema({
  id: 'task-board-acceptance',
  version: 1,
  collections: { todos: todo, counts: number, secrets: todo },
  mutations: { increment: { input: empty, output: number } },
})
