import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { once } from 'node:events'
import { mkdtemp, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { chromium, webkit } from 'playwright'

test(
  'packaged vanilla and React task boards share durable accepted state',
  { timeout: 90_000 },
  async () => {
    const directory = await mkdtemp(join(tmpdir(), 'sync-example-'))
    const example = process.env.SYNC_EXAMPLE_DIRECTORY
    assert.ok(example)
    const child = spawn(process.execPath, ['server.ts'], {
      cwd: example,
      env: { ...process.env, DATA_DIRECTORY: directory, PORT: '0' },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    const exited = once(child, 'exit')
    let output = ''
    child.stderr.on('data', (data) => {
      output += String(data)
    })
    const ready = Promise.withResolvers<string>()
    child.stdout.on('data', (data) => {
      output += String(data)
      const url = /Task board: (http:\/\/[^\s]+)/.exec(output)?.[1]
      if (url) ready.resolve(url)
    })
    child.once('exit', () => ready.reject(new Error(output)))
    child.once('error', ready.reject)
    try {
      const url = await ready.promise
      for (const type of [chromium, webkit]) {
        const browser = await type.launch({ headless: true })
        const context = await browser.newContext()
        context.setDefaultTimeout(10_000)
        const errors: string[] = []
        try {
          const vanilla = await context.newPage()
          const react = await context.newPage()
          for (const page of [vanilla, react])
            page.on('pageerror', (error) => errors.push(error.message))
          await vanilla.goto(url)
          await react.goto(url + '/react')
          const title = `Accepted in ${type.name()}`
          await vanilla.getByLabel('Add a task').fill(title)
          await vanilla
            .getByRole('button', { name: 'Add task', exact: true })
            .click()
          const row = react.locator('.task').filter({ hasText: title })
          await row
            .getByRole('button', { name: 'Complete', exact: true })
            .click()
          await vanilla
            .locator('.task[data-done="true"]')
            .filter({ hasText: title })
            .waitFor()
          await react.reload()
          await react
            .locator('.task[data-done="true"]')
            .filter({ hasText: title })
            .waitFor()
          assert.deepEqual(errors, [])
        } finally {
          await context.close()
          await browser.close()
        }
      }
    } finally {
      child.kill('SIGTERM')
      const [code, signal] = await exited
      assert.equal(signal, null, output)
      assert.equal(code, 0, output)
      await rm(directory, { recursive: true, force: true })
    }
  },
)
