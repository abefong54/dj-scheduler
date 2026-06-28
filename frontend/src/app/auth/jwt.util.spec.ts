import { describe, it, expect } from 'vitest';
import { decodeExp, isExpired } from './jwt.util';

function b64url(obj: object): string {
  return btoa(JSON.stringify(obj)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
function makeToken(payload: object): string {
  return `${b64url({ alg: 'HS256' })}.${b64url(payload)}.sig`;
}

describe('jwt.util', () => {
  it('decodes the exp claim', () => {
    expect(decodeExp(makeToken({ exp: 1893456000 }))).toBe(1893456000);
  });

  it('returns null for a malformed token', () => {
    expect(decodeExp('not-a-jwt')).toBeNull();
  });

  it('returns null when exp is missing', () => {
    expect(decodeExp(makeToken({ sub: 'x' }))).toBeNull();
  });

  it('treats a past exp as expired', () => {
    expect(isExpired(makeToken({ exp: 1000 }), 2_000_000)).toBe(true);
  });

  it('treats a future exp as not expired', () => {
    expect(isExpired(makeToken({ exp: 9_999_999_999 }), Date.now())).toBe(false);
  });

  it('treats a malformed token as expired', () => {
    expect(isExpired('garbage')).toBe(true);
  });
});
