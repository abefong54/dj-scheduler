import { Component, ElementRef, computed, inject, signal, viewChild } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { toPng } from 'html-to-image';
import { ApiService, PublicSlot } from '../services/api.service';
import { environment } from '../../environments/environment';
import { genreTheme } from './genre-theme';
import { ButtonComponent } from '../shared/button.component';

// Pilot brand for the card lockup. IMDJ is the beachhead school; this becomes
// per-school config when EL-045 (configurable curriculum/branding) lands.
const BRAND = 'IMDJ';

// The card's intrinsic export size — matches the approved 4:5 artifact (EL-053).
const EXPORT_WIDTH = 1080;

const WEEKDAYS = ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'];
const MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];

/**
 * Public per-DJ "I'm playing" share card (EL-049). Renders the Echo Press card
 * (see Design/echo-press.md + student-card.png) from a single slot's public
 * data, with Share-to-LINE. No auth — the slot data is already public via the
 * schedule. The shared link points at the backend OG endpoint (/s/dj/:id) so it
 * unfurls on LINE/social; this Angular page is what a human lands on.
 */
@Component({
  selector: 'app-card',
  standalone: true,
  imports: [TranslatePipe, RouterLink, ButtonComponent],
  templateUrl: './card.component.html',
  styleUrl: './card.component.css',
})
export class CardComponent {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);

  private slotId = this.route.snapshot.paramMap.get('slotId') ?? '';

  // The rendered card element, captured for client-side PNG export (EL-053).
  private cardRef = viewChild<ElementRef<HTMLElement>>('cardEl');

  readonly brand = BRAND;
  data = signal<PublicSlot | null>(null);
  loading = signal(true);
  notFound = signal(false);
  copied = signal(false);
  exporting = signal(false);

  djName = computed(() => this.data()?.slot.dj_name ?? '');
  genre = computed(() => this.data()?.slot.genre ?? '');
  stage = computed(() => this.data()?.slot.stage_name ?? '');
  time = computed(() => this.data()?.slot.start_time ?? '');
  eventName = computed(() => this.data()?.event.name ?? '');
  venue = computed(() => this.data()?.event.venue_name ?? '');
  eventId = computed(() => this.data()?.event.id ?? '');

  // Date inside the visual card stays in English abbreviations to match the
  // approved brand artifact (student-card.png); the surrounding UI is i18n'd.
  dateLabel = computed(() => {
    const iso = this.data()?.slot.slot_date;
    if (!iso) return '';
    const [y, m, d] = iso.split('-').map(Number);
    if (!y || !m || !d) return iso;
    // Build in UTC so the weekday never shifts with the viewer's timezone.
    const dt = new Date(Date.UTC(y, m - 1, d));
    return `${WEEKDAYS[dt.getUTCDay()]} ${d} ${MONTHS[m - 1]}`;
  });

  year = computed(() => this.data()?.slot.slot_date?.slice(0, 4) ?? '');

  // The headline shrinks for longer names so the echo wordmark stays on one line.
  nameSizeCqw = computed(() => {
    const n = this.djName().length;
    if (n <= 4) return 23;
    if (n <= 6) return 18;
    if (n <= 9) return 13;
    return 10;
  });

  // Genre-derived theme (EL-052): maps the slot's genre to a card palette, with
  // the brand violet/lime default applied by the base .card rule. The matching
  // CSS-variable overrides live under `.card.theme-<key>` in card.component.css.
  themeClass = computed(() => {
    const theme = genreTheme(this.genre());
    return theme === 'default' ? '' : `theme-${theme}`;
  });

  // A white vinyl spiral (Archimedean) drawn as a polyline so it exports crisply.
  spiralPoints = computed(() => {
    const pts: string[] = [];
    const steps = 240;
    const turns = 5;
    for (let i = 0; i <= steps; i++) {
      const t = (i / steps) * turns * 2 * Math.PI;
      const r = t * 3;
      pts.push(`${(r * Math.cos(t)).toFixed(1)},${(r * Math.sin(t)).toFixed(1)}`);
    }
    return pts.join(' ');
  });

  constructor() {
    if (!this.slotId) {
      this.notFound.set(true);
      this.loading.set(false);
      return;
    }
    this.api.getPublicSlot(this.slotId).subscribe({
      next: (d) => {
        this.data.set(d);
        this.loading.set(false);
      },
      error: () => {
        this.notFound.set(true);
        this.loading.set(false);
      },
    });
  }

  // Absolute link to the backend OG endpoint. In prod environment.apiUrl is the
  // backend origin; in dev it's empty, so fall back to the page origin (the dev
  // proxy forwards /s to the API).
  private shareLink(): string {
    const base = environment.apiUrl || window.location.origin;
    return `${base}/s/dj/${this.slotId}`;
  }

  shareViaLine() {
    const url = encodeURIComponent(this.shareLink());
    window.open(`https://line.me/R/msg/text/?${url}`, '_blank');
  }

  async copyLink() {
    try {
      await navigator.clipboard.writeText(this.shareLink());
      this.copied.set(true);
      setTimeout(() => this.copied.set(false), 2000);
    } catch {
      // Clipboard unavailable (e.g. insecure context) — silently no-op; the LINE
      // share remains the primary action.
    }
  }

  // Download filename, slugged from the DJ + event so saved cards are findable.
  fileName(): string {
    const slug = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
    return `${slug(this.djName()) || 'set'}-${slug(this.eventName()) || 'soundcheck'}.png`;
  }

  // Export the card to a PNG for Instagram/feed, where you need an image, not a
  // link (EL-053). Renders the live card DOM client-side and scales it up to the
  // intrinsic export width so the output is crisp regardless of on-screen size.
  async saveImage() {
    const node = this.cardRef()?.nativeElement;
    if (!node || this.exporting()) return;
    this.exporting.set(true);
    try {
      // Wait for the brand web fonts so glyphs aren't swapped mid-capture.
      await document.fonts?.ready;
      const pixelRatio = node.offsetWidth ? EXPORT_WIDTH / node.offsetWidth : 2;
      const dataUrl = await toPng(node, { pixelRatio, cacheBust: true });
      const link = document.createElement('a');
      link.href = dataUrl;
      link.download = this.fileName();
      link.click();
    } catch {
      // Export can fail (e.g. cross-origin font embedding) — keep the page intact;
      // Share-to-LINE remains the primary path.
    } finally {
      this.exporting.set(false);
    }
  }
}
