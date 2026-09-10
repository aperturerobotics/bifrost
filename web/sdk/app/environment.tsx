import { atom, type Atom, type StateType } from '@s4wave/web/state/persist.js'
import { createContext, use } from 'react'
import type { Root } from '@s4wave/sdk/root'
import {
  getAppNavigation,
  getAppNavigationGeneration,
  getAppPath,
  normalizeAppPath,
  setAppPath,
  type AppPathNavigation,
} from '@s4wave/web/router/app-path.js'

export interface AppNavigation {
  getAppPath(): string
  getAppNavigation(): AppPathNavigation
  getAppNavigationGeneration(): number
  setAppPath(path: string, params?: Record<string, string>): void
  subscribe(listener: () => void): () => void
}

// AppEnvironment owns browser state for one mounted application. Backend
// persistence is supplied independently by its ResourceClient.
export interface AppEnvironment {
  id: string
  instanceKey: string
  httpPathPrefix: string
  rootAtom: Atom<StateType>
  storage: Storage
  documentStorage: Storage
  navigation: AppNavigation
  events: EventTarget
}

export class MemoryStorage implements Storage {
  private readonly values = new Map<string, string>()
  private readonly listeners = new Set<() => void>()

  snapshot(): Record<string, string> {
    return Object.fromEntries(this.values)
  }

  restore(values: Record<string, string>): void {
    this.values.clear()
    for (const [key, value] of Object.entries(values))
      this.values.set(key, value)
  }

  subscribe(listener: () => void): () => void {
    this.listeners.add(listener)
    return () => {
      this.listeners.delete(listener)
    }
  }

  private changed(): void {
    for (const listener of this.listeners) listener()
  }

  get length() {
    return this.values.size
  }
  clear() {
    if (!this.values.size) return
    this.values.clear()
    this.changed()
  }
  getItem(key: string) {
    return this.values.get(key) ?? null
  }
  key(index: number) {
    return [...this.values.keys()][index] ?? null
  }
  removeItem(key: string) {
    if (this.values.delete(key)) this.changed()
  }
  setItem(key: string, value: string) {
    if (this.values.get(key) === value) return
    this.values.set(key, value)
    this.changed()
  }
}

export function createAppEnvironment(id: string, path = '/'): AppEnvironment {
  if (!id) throw new Error('An app instance requires an ID')
  let current: AppPathNavigation = { path: normalizeAppPath(path), params: {} }
  let generation = 0
  const listeners = new Set<() => void>()
  return {
    id,
    instanceKey: crypto.randomUUID(),
    httpPathPrefix: '',
    rootAtom: atom<StateType>({}),
    storage: new MemoryStorage(),
    documentStorage: new MemoryStorage(),
    events: new EventTarget(),
    navigation: {
      getAppPath: () => current.path,
      getAppNavigation: () => current,
      getAppNavigationGeneration: () => generation,
      setAppPath: (nextPath, params = current.params) => {
        const next = { path: normalizeAppPath(nextPath), params: { ...params } }
        if (JSON.stringify(next) === JSON.stringify(current)) return
        current = next
        generation++
        queueMicrotask(() => {
          for (const listener of listeners) listener()
        })
      },
      subscribe: (listener) => {
        listeners.add(listener)
        return () => {
          listeners.delete(listener)
        }
      },
    },
  }
}

const browserEnvironment: AppEnvironment = {
  id: '',
  instanceKey: '',
  httpPathPrefix: '',
  rootAtom: atom<StateType>({}),
  get storage() {
    return localStorage
  },
  get documentStorage() {
    return sessionStorage
  },
  get events() {
    return window
  },
  navigation: {
    getAppPath,
    getAppNavigation,
    getAppNavigationGeneration,
    setAppPath,
    subscribe: (listener) => {
      window.addEventListener('hashchange', listener)
      window.addEventListener('popstate', listener)
      return () => {
        window.removeEventListener('hashchange', listener)
        window.removeEventListener('popstate', listener)
      }
    },
  },
}

export const AppEnvironmentContext = createContext(browserEnvironment)
export function useAppEnvironment(): AppEnvironment {
  return use(AppEnvironmentContext)
}
export function useAppNavigation(): AppNavigation {
  return useAppEnvironment().navigation
}

const rootEnvironments = new WeakMap<Root, AppEnvironment>()

// bindRootEnvironment scopes non-React operations such as Quickstart to the
// same environment as the frontend. The returned cleanup is generation-safe.
export function bindRootEnvironment(root: Root, environment: AppEnvironment) {
  rootEnvironments.set(root, environment)
  return () => {
    if (rootEnvironments.get(root) === environment)
      rootEnvironments.delete(root)
  }
}
export function getRootEnvironment(root: Root): AppEnvironment {
  return rootEnvironments.get(root) ?? browserEnvironment
}
