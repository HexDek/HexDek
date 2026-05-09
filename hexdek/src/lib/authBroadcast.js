// Cross-tab signal that the magic-link tab uses to tell the original
// "check your email" tab that the user has authenticated. The new tab
// posts {type:'authed', email}; the original tab listens, plays a
// console animation, and redirects to /operator.
//
// BroadcastChannel is supported in every evergreen browser. We
// gracefully no-op when it isn't available (e.g. very old Safari) —
// the new tab still completes its own redirect, just without the
// hand-off.

export const AUTH_CHANNEL = 'hexdek-auth'

export const AUTH_EVENT = {
  SUCCEEDED: 'authed',
  FAILED: 'auth-failed',
}

function supported() {
  return typeof window !== 'undefined' && typeof window.BroadcastChannel === 'function'
}

export function broadcastAuth(payload) {
  if (!supported()) return
  let ch
  try {
    ch = new BroadcastChannel(AUTH_CHANNEL)
    ch.postMessage(payload)
  } catch { /* ignore */ }
  // Close on next tick — Chrome silently drops messages when the
  // channel closes synchronously after postMessage.
  setTimeout(() => { try { ch && ch.close() } catch {} }, 50)
}

export function listenAuth(handler) {
  if (!supported()) return () => {}
  const ch = new BroadcastChannel(AUTH_CHANNEL)
  const onMsg = (evt) => { try { handler(evt.data) } catch {} }
  ch.addEventListener('message', onMsg)
  return () => {
    try { ch.removeEventListener('message', onMsg) } catch {}
    try { ch.close() } catch {}
  }
}
