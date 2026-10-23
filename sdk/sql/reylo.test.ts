import { describe, expect, test } from 'vitest'

import type { SqlValue } from '../../db/sql/sql.pb.js'
import { reyloClient, type SqlStatementRunner } from './reylo.js'

describe('reyloClient', () => {
  test('runs reads as queries, decoded by column type, and writes as exec', async () => {
    // Record each call and answer reads with one row of every decoded kind.
    const calls: { kind: string; query: string; args?: SqlValue[] }[] = []
    const runner: SqlStatementRunner = {
      exec(query, args) {
        calls.push({ kind: 'exec', query, args })
        return Promise.resolve({ rowsAffected: 1n })
      },
      query(query, args) {
        calls.push({ kind: 'query', query, args })
        const cols = [
          ['id', 'BIGINT'],
          ['n', 'INT'],
          ['price', 'DECIMAL'],
          ['at', 'DATETIME'],
          ['day', 'DATE'],
          ['doc', 'JSON'],
          ['raw', 'BLOB'],
          ['name', 'TEXT'],
          ['email', 'VARCHAR'],
        ]
        const values: SqlValue[] = [
          { value: { case: 'intValue', value: 9007199254740993n } },
          { value: { case: 'intValue', value: 7n } },
          { value: { case: 'strValue', value: '1.50' } },
          { value: { case: 'strValue', value: '2026-10-23T08:24:40.5Z' } },
          { value: { case: 'strValue', value: '2026-10-23T00:00:00Z' } },
          { value: { case: 'strValue', value: '{"a":[1]}' } },
          { value: { case: 'blobValue', value: new Uint8Array([1, 2]) } },
          { value: { case: 'strValue', value: 'ada' } },
          {},
        ]
        return Promise.resolve({
          columns: cols.map(([name, databaseTypeName]) => ({
            name,
            databaseTypeName,
          })),
          rows: [{ values }],
        })
      },
    }
    const client = reyloClient(runner)

    // A read, written the way Reylo passes it, decodes every column.
    const rows = await client.query(
      { sql: '/* q */ (select * from t where id = ? and ok = ?)' },
      [3, true],
    )
    expect(rows).toEqual([
      {
        id: 9007199254740993n,
        n: 7,
        price: '1.50',
        at: '2026-10-23T08:24:40.5',
        day: '2026-10-23',
        doc: { a: [1] },
        raw: new Uint8Array([1, 2]),
        name: 'ada',
        email: null,
      },
    ])

    // A write goes through exec with its parameters bound.
    const result = await client.query(
      'insert into t (a, b, c) values (?, ?, ?)',
      [1.5, null, new Uint8Array([3])],
    )
    expect(Array.isArray(result)).toBe(false)
    expect(calls.map((c) => c.kind)).toEqual(['query', 'exec'])
    expect(calls[0].args).toEqual([
      { value: { case: 'intValue', value: 3n } },
      { value: { case: 'intValue', value: 1n } },
    ])
    expect(calls[1].args).toEqual([
      { value: { case: 'floatValue', value: 1.5 } },
      {},
      { value: { case: 'blobValue', value: new Uint8Array([3]) } },
    ])

    // A parameter that would lose its value is rejected before any call.
    for (const bad of [2 ** 53, Number.NaN, 2n ** 63n, new Date()]) {
      await expect(client.query('select ?', [bad])).rejects.toThrow()
    }
    expect(calls).toHaveLength(2)
  })
})
