import type { StandardSchemaV1 } from '@standard-schema/spec'
import { createAccess } from './access.js'
import { defineSchema } from './schema.js'

declare const textToNumber: StandardSchemaV1<string, number>
declare const numberToText: StandardSchemaV1<number, string>
const schema = defineSchema({
  id: 'inference',
  version: 1,
  collections: { counts: textToNumber },
  mutations: { summarize: { input: textToNumber, output: numberToText } },
})
const database = createAccess<typeof schema>(async () => null)
const collection = database.collection('counts')
const read: Promise<number | undefined> = collection.get('x')
const result: Promise<string> = database.mutate('summarize', '4')
void read
void result
void collection.put('x', '4')
// @ts-expect-error Records take the validator's input, not its transformed output.
void collection.put('x', 4)
// @ts-expect-error Unknown collection names are rejected by the shared schema alone.
database.collection('unknown')
// @ts-expect-error Mutation input inference does not require server handler imports.
void database.mutate('summarize', 4)
