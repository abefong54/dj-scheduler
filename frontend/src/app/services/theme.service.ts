import { DOCUMENT } from '@angular/common';
import { Injectable, effect, inject, signal } from '@angular/core';

/** The two Soundcheck canvases (EL-078 / EL-072 tokens). */
export const THEME_MODES = ['dark', 'light'] as const;
export type ThemeMode = (typeof THEME_MODES)[number];

/** Dark "booth" mode is the app's default surface. */
const DEFAULT_MODE: ThemeMode = 'dark';

/**
 * Soundcheck brand/mode switch (EL-078).
 *
 * `data-brand="soundcheck"` ships pre-set on `<html>` (index.html) so the
 * design system is active from first paint with no flash. This service owns the
 * runtime *mode* toggle: dark "booth" mode is the default (no attribute), and
 * `setMode('light')` writes `data-mode="light"` to switch to the light "admin"
 * canvas. Only the canvas + neutral ink flip; the accent + status hues are
 * shared across both modes (see the token layer in styles.css).
 *
 * Skin only — this changes no data, routing, or behavior.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly document = inject(DOCUMENT);

  /** Current canvas mode. Dark "booth" by default. */
  readonly mode = signal<ThemeMode>(DEFAULT_MODE);

  constructor() {
    // Reflect the mode onto <html>. Dark is the default and carries no
    // attribute (matching index.html), so we only set data-mode for light.
    effect(() => {
      const root = this.document.documentElement;
      if (this.mode() === 'light') {
        root.setAttribute('data-mode', 'light');
      } else {
        root.removeAttribute('data-mode');
      }
    });
  }

  /** Switch the canvas to an explicit mode. */
  setMode(mode: ThemeMode): void {
    this.mode.set(mode);
  }

  /** Flip between dark "booth" and light "admin" canvases. */
  toggleMode(): void {
    this.mode.update((m) => (m === 'dark' ? 'light' : 'dark'));
  }
}
