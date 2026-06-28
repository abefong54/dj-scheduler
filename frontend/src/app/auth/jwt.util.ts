// Minimal client-side JWT helpers. We only read the `exp` claim to decide
// whether to treat a token as logged-in; the backend remains the source of truth.

/** Returns the `exp` claim (unix seconds) or null if absent/malformed. */
export function decodeExp(token: string): number | null {
  const parts = token.split('.');
  if (parts.length !== 3) return null;
  try {
    const json = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'));
    const payload = JSON.parse(json) as { exp?: unknown };
    return typeof payload.exp === 'number' ? payload.exp : null;
  } catch {
    return null;
  }
}

/** A token is expired if it has no usable exp claim or exp is in the past. */
export function isExpired(token: string, nowMs: number = Date.now()): boolean {
  const exp = decodeExp(token);
  if (exp === null) return true;
  return exp * 1000 <= nowMs;
}
