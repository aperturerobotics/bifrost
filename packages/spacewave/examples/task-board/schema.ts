import { defineSchema } from 'spacewave'
import { z } from 'zod'

/** todo validates stored tasks and the result of completing a task. */
const todo = z.object({
  title: z.string().min(1).max(120),
  done: z.boolean(),
})

/** schema is the shared contract for the server, vanilla client, and React client. */
export const schema = defineSchema({
  id: 'task-board',
  version: 1,
  collections: { todos: todo },
  mutations: {
    completeTodo: {
      input: z.object({ id: z.string().min(1) }),
      output: todo,
    },
  },
})
