import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import { Resource } from '@aptre/bldr-sdk/resource/resource.js'

import type {
  Dashboard,
  LabelSet,
  Panel,
  QueryRequest,
  QueryResponse,
  TargetInfo,
  TargetMeta,
} from './prom.pb.js'
import {
  PromDashboardResourceServiceClient,
  PromQueryResourceServiceClient,
} from './prom_srpc.pb.js'

// TargetTypeID is the World ObjectType id of a Scrape target.
export const TargetTypeID = 'prom/target'

// DashboardTypeID is the World ObjectType id of a Dashboard.
export const DashboardTypeID = 'prom/dashboard'

// defaultScrapeIntervalMs is the interval of a target that sets none.
const defaultScrapeIntervalMs = 15_000

// PromTargets maps each target's object key to its configuration and status.
export type PromTargets = ReadonlyMap<string, TargetInfo>

// PromQuery is the typed handle of a prom/target object. It answers PromQL
// over every target in the Space.
export class PromQuery extends Resource {
  protected readonly queryService: PromQueryResourceServiceClient

  constructor(resourceRef: ClientResourceRef) {
    super(resourceRef)
    this.queryService = new PromQueryResourceServiceClient(resourceRef.client)
  }

  // query evaluates one PromQL expression.
  public query(
    req: QueryRequest,
    abortSignal?: AbortSignal,
  ): Promise<QueryResponse> {
    return this.queryService.Query(req, abortSignal)
  }

  // labelNames lists the label names of the series matching match.
  public async labelNames(
    match: string[],
    abortSignal?: AbortSignal,
  ): Promise<string[]> {
    const resp = await this.queryService.LabelNames({ match }, abortSignal)
    return resp.values ?? []
  }

  // labelValues lists the values of the label name among the series matching
  // match.
  public async labelValues(
    name: string,
    match: string[],
    abortSignal?: AbortSignal,
  ): Promise<string[]> {
    const resp = await this.queryService.LabelValues(
      { name, match },
      abortSignal,
    )
    return resp.values ?? []
  }

  // watchTargets yields every target in the Space, then the whole set again
  // after each change the plugin reports, at most once per second. A target
  // changes on each scrape commit, so each yield also means query answers may
  // have changed.
  public async *watchTargets(
    abortSignal?: AbortSignal,
  ): AsyncIterable<PromTargets> {
    const targets = new Map<string, TargetInfo>()
    for await (const delta of this.queryService.WatchTargets({}, abortSignal)) {
      for (const key of delta.removed ?? []) {
        targets.delete(key)
      }
      for (const info of delta.changed ?? []) {
        targets.set(info.objectKey ?? '', info)
      }
      yield new Map(targets)
    }
  }
}

// PromDashboard is the typed handle of a prom/dashboard object.
export class PromDashboard extends PromQuery {
  private readonly dashboardService: PromDashboardResourceServiceClient

  constructor(resourceRef: ClientResourceRef) {
    super(resourceRef)
    this.dashboardService = new PromDashboardResourceServiceClient(
      resourceRef.client,
    )
  }

  // watchDashboard yields the Dashboard, then the Dashboard after each commit
  // that changes it.
  public watchDashboard(abortSignal?: AbortSignal): AsyncIterable<Dashboard> {
    return this.dashboardService.WatchDashboard({}, abortSignal)
  }

  // putPanel adds panel when its id is empty, or replaces the panel with its
  // id, and returns the stored id. Replacing a panel another member removed
  // fails.
  public async putPanel(
    panel: Panel,
    abortSignal?: AbortSignal,
  ): Promise<string> {
    const resp = await this.dashboardService.PutPanel({ panel }, abortSignal)
    return resp.id ?? ''
  }

  // removePanel removes the panel id.
  public async removePanel(
    id: string,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    await this.dashboardService.RemovePanel({ id }, abortSignal)
  }

  // movePanel moves the panel id to index.
  public async movePanel(
    id: string,
    index: number,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    await this.dashboardService.MovePanel({ id, index }, abortSignal)
  }

  // createTarget creates a Scrape target at objectKey.
  public async createTarget(
    objectKey: string,
    meta: TargetMeta,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    await this.dashboardService.CreateTarget({ objectKey, meta }, abortSignal)
  }
}

// targetStaleAfterMs returns how long after its last scrape a target's status
// goes stale: two scrape intervals plus one commit window of the scraping
// Device. It matches TargetStatus.StaleAfter in Go.
export function targetStaleAfterMs(info: TargetInfo): number {
  const interval = info.meta?.scrape?.intervalMs || defaultScrapeIntervalMs
  return 2 * interval + (info.status?.commitWindowMs ?? 0)
}

// isTargetStale reports whether the target has a status whose scraping Device
// stopped committing by nowMs.
export function isTargetStale(info: TargetInfo, nowMs: number): boolean {
  const last = Number(info.status?.lastScrapeMs ?? 0n)
  if (!info.status || last === 0) return false
  return nowMs - last > targetStaleAfterMs(info)
}

// labelValue returns the value of the label name in set, or ''.
export function labelValue(set: LabelSet | undefined, name: string): string {
  return set?.labels?.find((l) => l.name === name)?.value ?? ''
}
