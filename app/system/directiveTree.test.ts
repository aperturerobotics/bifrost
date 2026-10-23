import { describe, expect, it } from 'vitest'

import { buildDirectiveTree, relevantDirectives } from './directiveTree.js'

// response is a session bus where a ConfigSet entry loads a controller.
// The controller adds a shared lookup, which resolves by adding a link.
// A loop directive adds itself.
const response = {
  directiveCount: 6,
  directives: [
    {
      id: 1,
      name: 'ApplyConfigSet',
      ident: 'ApplyConfigSet<controller-keys=fg/node1@1>',
      references: [{}],
    },
    {
      id: 2,
      name: 'LoadControllerWithConfig',
      ident: 'LoadControllerWithConfig<config-id=hydra/node>',
      references: [
        {
          originDirectiveId: 1,
          originDirectiveIdent: 'ApplyConfigSet<controller-keys=fg/node1@1>',
          originControllerId: 'controllerbus/configset',
          configKey: 'fg/node1',
        },
      ],
    },
    {
      id: 3,
      name: 'LookupVolume',
      ident: 'LookupVolume',
      references: [
        {
          originDirectiveId: 2,
          originControllerId: 'hydra/node',
        },
        { originControllerId: 'core/session' },
        { originControllerId: 'devtool/status', weak: true },
      ],
    },
    {
      id: 4,
      name: 'EstablishLinkWithPeer',
      ident: 'EstablishLinkWithPeer<peer=a>',
      references: [
        { originDirectiveId: 3, originControllerId: 'hydra/volume' },
      ],
    },
    {
      id: 5,
      name: 'Loop',
      ident: 'Loop',
      references: [
        { originControllerId: 'test/loop' },
        { originDirectiveId: 5, originControllerId: 'test/loop' },
      ],
    },
    {
      id: 6,
      name: 'Bridged',
      ident: 'Bridged',
      references: [
        {
          originDirectiveIdent: 'LookupRpcService<plugin>',
          originControllerId: 'bus-bridge',
        },
      ],
    },
  ],
}

describe('buildDirectiveTree', () => {
  it('links each directive to its origin directives and controllers', () => {
    const tree = buildDirectiveTree(response)

    expect(tree.unlinked).toEqual([1])
    expect(tree.configKeys).toEqual(['fg/node1'])
    expect(tree.nodes.get(1)?.children).toEqual([
      { id: 2, via: 'controllerbus/configset' },
    ])
    expect(tree.nodes.get(2)?.configKeys).toEqual(['fg/node1'])
    expect(tree.nodes.get(2)?.children).toEqual([{ id: 3, via: 'hydra/node' }])
    expect(tree.nodes.get(3)?.children).toEqual([
      { id: 4, via: 'hydra/volume' },
    ])
    expect(tree.nodes.get(5)?.children).toEqual([{ id: 5, via: 'test/loop' }])
    expect(tree.controllers).toEqual([
      { label: 'core/session', children: [{ id: 3, via: 'core/session' }] },
      { label: 'test/loop', children: [{ id: 5, via: 'test/loop' }] },
    ])
    expect(tree.remote).toEqual([
      {
        label: 'LookupRpcService<plugin>',
        children: [{ id: 6, via: 'bus-bridge' }],
      },
    ])
  })

  it('links a directive held only by weak references through them', () => {
    const tree = buildDirectiveTree({
      directives: [
        {
          id: 1,
          name: 'Observed',
          references: [{ originControllerId: 'devtool/status', weak: true }],
        },
      ],
    })

    expect(tree.unlinked).toEqual([])
    expect(tree.controllers).toEqual([
      { label: 'devtool/status', children: [{ id: 1, via: 'devtool/status' }] },
    ])
  })
})

describe('relevantDirectives', () => {
  it('keeps each match and every directive above it', () => {
    const tree = buildDirectiveTree(response)

    expect(relevantDirectives(tree, '')).toBeNull()
    expect([...relevantDirectives(tree, 'peer=a')!].sort()).toEqual([
      1, 2, 3, 4,
    ])
    expect([...relevantDirectives(tree, 'fg/node1')!].sort()).toEqual([1, 2])
  })
})
