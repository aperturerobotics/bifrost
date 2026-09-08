---
title: Quickstarts and App Surfaces
section: objects
order: 2
summary: Add a static Quickstart, seed its content, test its initial route, and rebuild its public page.
---

A Quickstart is a first-run or in-session creation path. Space-creating
Quickstarts seed World objects, write Space settings, and open the initial
route. The local-session option creates or reuses a session without creating a
Space; account and pairing options navigate to their existing setup pages.

## Static and dynamic Quickstarts

Static Quickstarts live in `app/quickstart/options.ts`. Normal visible creation
options include Space, Drive, Git, and Canvas. Notebook, Chat, KV, SQL, Docs,
Blog, V86, Device, and Forge are experimental in the current catalog.

Dynamic Quickstarts come from the Quickstart registry. They must supply an ID,
name, description, category, and plugin ID. Dynamic options can add required
plugin IDs and a default Space name. They are app options, not public static
`/quickstart/:id` pages.

## Add a static Quickstart

Use `drive` in `app/quickstart/create.ts` as the example for content built into
the core plugin, or `notebook` for content supplied by another plugin.

1. Add the option to `QUICKSTART_OPTIONS` in `app/quickstart/options.ts`. Give it
   a stable ID, name, description, category, and icon. Add `seoDescription` for
   its public page; metadata tests require 120 to 160 characters. Set
   `experimental: true` while the feature is experimental. A `path` makes the
   option navigation-only, so leave it unset for a Space-creating Quickstart.
2. Add the ID to `QUICKSTART_SEEDS` in `app/quickstart/create.ts` with its default
   Space name. Supply `initialObjectKey` and `initialObjectType` only when the
   first navigation should address an object directly. Otherwise navigation
   opens the Space's configured index. The descriptor table is exhaustive over
   `QuickstartSpaceCreateId`.
3. Add its case to `populateSpace`. Use the existing SDK operation or resource
   that creates the content. Pass the attempt's abort signal through calls and
   release acquired resources through their existing cleanup contract. For
   plugin content, install the plugin and wait for its Quickstart or ObjectType
   registration before invoking it; the Notes and SQL cases show both steps.
4. Write the final index path with `createSpaceSettingsObject`, or use
   `ensureSpacePlugins` when installing required plugins. These helpers preserve
   settings that the new content does not replace. Point the index at an object
   the seed operation actually created. Dynamic Quickstart results can provide
   the index path and plugin IDs through `executeDynamicQuickstart`.
5. Extend the creation and catalog tests below. Open the option from the public
   Quickstart URL when it is release-visible, and from an existing session's
   create-Space screen. Confirm the seeded content opens at the intended route
   and still opens after leaving and returning to the Space.

`QuickstartId` is derived from the option table. Its navigation-only exclusions
are explicit: `account` and `pair` are not creation IDs, and `local` is not a
Space-creation ID. A new navigation-only action needs the same classification
in the creation-ID helpers; setting `path` alone does not change those types.

## Test the creation contract

Add a focused case in `app/quickstart/create.test.ts` using its existing
`buildQuickstartWorld` fixture. Assert the content operation, the saved index
path, and any required plugin IDs. Extend “indexes every quickstart to the
object it creates or seeds” so navigation cannot point at a missing object.
For plugin content, also verify that registration precedes execution.

Update `app/quickstart/options.test.ts` for catalog visibility and
`app/prerender/static-pages.test.ts` for release-page inventory. Run these from
the repository root:

```sh
bun run test:js:alpha -- app/quickstart/create.test.ts app/quickstart/options.test.ts app/prerender/static-pages.test.ts
bun run typecheck
```

These tests check creation and routing contracts. Exercise the running app to
check plugin availability, the rendered initial view, and cancellation during
setup.

## Rebuild public Quickstart pages

`PUBLIC_QUICKSTART_OPTIONS` supplies the release inventory in
`app/prerender/static-pages.ts`. It excludes experimental, hidden, dynamic, and
navigation-only options. A release-visible static option automatically receives
`/quickstart/{id}` metadata and the shared `QuickstartLoading` page; there is no
second per-option route list to edit.

Build the browser assets, hydration bundle, and prerender bundle in this order
from the repository root:

```sh
bun run build:release:web
bun run vite build --config app/prerender/vite.hydrate.config.ts
bun run vite build --config app/prerender/vite.ssr.config.ts
bun run app/prerender/ssr-dist/build.js --dist-dir .bldr-dist/build/js/spacewave-browser/dist
```

The final command writes the public HTML, `static-manifest.ts`, and `sitemap.xml`
under `app/prerender/dist/`. Check that the new route appears in the manifest
and sitemap, then open its generated page through the local release preview.
Keep these generated assets in the release output; publishing them is a separate
release operation.

## Creation flow

The standalone Quickstart route creates or reuses a local session, creates a
Space for non-local creation options, populates content, stages handoff
resources, and redirects to `/u/{session}/so/{space}` plus an initial object
route when one exists.

The in-session create-space route uses the current session or organization,
mounts the new Space, populates it, and navigates to the result. Canceling the
progress UI returns to the dashboard and does not promise to delete an in-flight
Space.

## SpaceSettings

SpaceSettings live in the hidden settings object. Today they carry the Space
index path and plugin IDs. Helpers that update the index path preserve existing
settings. Dynamic Quickstarts can return plugin IDs and index path values, which
are written into SpaceSettings.

## First-run boundary

Seed only the content that exists today. If a Quickstart depends on a plugin,
wait for that plugin registration or report that the surface is unavailable.
Do not advertise a first-run route for a dynamic Quickstart until the app route
accepts it.
