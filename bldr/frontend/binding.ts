import type { Binding } from './frontend.pb.js'

/** frontendBindingPath is a stable attachment resolved by the project owner. */
export function frontendBindingPath(binding: Binding): string {
  if (!binding.entrypoint) throw new Error('frontend binding has no entrypoint')
  return (
    '/b/fe/entrypoint/' +
    binding.entrypoint.split('/').map(encodeURIComponent).join('/')
  )
}
