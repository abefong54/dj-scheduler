import { Component, computed, inject, input } from '@angular/core';
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser';

/**
 * The minimal media reference a submission carries. A full `Submission`
 * (services/challenges.service) is structurally assignable to this, so callers
 * can pass the whole submission or just the three media fields.
 */
export interface MediaRef {
  media_kind: string;     // 'embed' today; 'file' is the forward-compat seam
  media_provider: string; // youtube | soundcloud | mixcloud
  media_ref: string;      // extracted id (youtube) or normalized URL (sc/mixcloud)
}

/** Providers we know how to embed. Anything else falls through to the fallback. */
export type EmbedProvider = 'youtube' | 'soundcloud' | 'mixcloud';

/**
 * Build a known-good iframe `src` from a FIXED template per provider. We never
 * inject provider-returned HTML — the backend hands us only `{provider, ref}`
 * (host-allowlisted + extracted), and this function is the only place a URL is
 * constructed. A student cannot smuggle markup through the embed seam.
 *
 * Templates (confirm against provider docs on drift — the approach is stable):
 *   youtube    → https://www.youtube.com/embed/<ref>
 *   soundcloud → https://w.soundcloud.com/player/?url=<encoded ref>
 *   mixcloud   → https://player.mixcloud.com/widget/iframe/?feed=<encoded ref>
 *
 * Returns null for an unknown provider or an empty ref, so the caller renders a
 * fallback instead of a broken frame.
 */
export function buildEmbedUrl(provider: string, ref: string): string | null {
  if (!ref) return null;
  switch (provider) {
    case 'youtube':
      return `https://www.youtube.com/embed/${encodeURIComponent(ref)}`;
    case 'soundcloud':
      return `https://w.soundcloud.com/player/?url=${encodeURIComponent(ref)}`;
    case 'mixcloud':
      return `https://player.mixcloud.com/widget/iframe/?feed=${encodeURIComponent(ref)}`;
    default:
      return null;
  }
}

/**
 * media-embed — the render-by-`media_kind` seam (see the design's "The Embed
 * Seam" / "Forward Compatibility"). Every media consumer (feed, review, profile)
 * only asks "render this submission's media"; the day we own files, add a `file`
 * branch here and nothing in the challenge loop changes.
 *
 *   <app-media-embed [media]="submission" />
 */
@Component({
  selector: 'app-media-embed',
  standalone: true,
  template: `
    @switch (media().media_kind) {
      @case ('embed') {
        @if (safeSrc(); as src) {
          <iframe
            class="w-full rounded border-0"
            [class.aspect-video]="isVideo()"
            [style.height.px]="isVideo() ? null : 166"
            [src]="src"
            loading="lazy"
            referrerpolicy="strict-origin-when-cross-origin"
            allow="autoplay; encrypted-media; picture-in-picture"
            allowfullscreen
            [attr.title]="'media.embed.' + provider()"
            data-testid="media-embed-frame"
          ></iframe>
        } @else {
          <div
            class="flex items-center justify-center rounded border border-dashed border-gray-300 p-6 text-sm text-gray-500"
            data-testid="media-embed-fallback"
          >
            Media unavailable.
          </div>
        }
      }
      <!-- TODO(forward-compat): when media_kind === 'file' lands, add a native
           <audio>/<video> player branch here — the storage backend + uploader are
           the only new pieces; the challenge loop, sign-off, badges, and feed are
           untouched. -->
      @default {
        <div
          class="flex items-center justify-center rounded border border-dashed border-gray-300 p-6 text-sm text-gray-500"
          data-testid="media-embed-fallback"
        >
          Unsupported media.
        </div>
      }
    }
  `,
})
export class MediaEmbedComponent {
  private sanitizer = inject(DomSanitizer);

  /** The submission (or bare media reference) to render. */
  media = input.required<MediaRef>();

  protected readonly provider = computed(() => this.media().media_provider);

  /** Video providers get a 16:9 frame; audio widgets get a short fixed height. */
  protected readonly isVideo = computed(() => this.provider() === 'youtube');

  /** The raw, self-constructed URL (null when we can't build one). */
  protected readonly rawSrc = computed(() =>
    buildEmbedUrl(this.provider(), this.media().media_ref),
  );

  /**
   * The trusted resource URL for the iframe. We call `bypassSecurityTrustResourceUrl`
   * ONLY on the URL we built ourselves from a fixed template above — never on
   * provider-supplied HTML or the student's pasted string.
   */
  protected readonly safeSrc = computed<SafeResourceUrl | null>(() => {
    const raw = this.rawSrc();
    return raw ? this.sanitizer.bypassSecurityTrustResourceUrl(raw) : null;
  });
}
