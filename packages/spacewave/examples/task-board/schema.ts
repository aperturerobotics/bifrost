import { defineSchema } from 'spacewave'
import { z } from 'zod'

const todo = z.object({ title: z.string().min(1).max(120), done: z.boolean() })
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
