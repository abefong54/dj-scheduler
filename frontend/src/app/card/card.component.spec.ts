import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { provideTranslateService } from '@ngx-translate/core';

import { toPng } from 'html-to-image';
import { CardComponent } from './card.component';
import { ApiService, PublicSlot } from '../services/api.service';

// Stub the DOM→PNG library so the export flow is testable without a real canvas.
vi.mock('html-to-image', () => ({ toPng: vi.fn().mockResolvedValue('data:image/png;base64,AAAA') }));

const PUBLIC_SLOT: PublicSlot = {
  slot: {
    id: 'slot-1',
    event_id: 'evt-1',
    stage_id: 'st-1',
    stage_name: 'Main Stage',
    dj_id: 'dj-1',
    dj_name: 'Cody',
    genre: 'House',
    slot_date: '2026-07-12', // a Sunday-relative check: 2026-07-12 is a Sunday
    start_time: '22:00',
    end_time: '23:00',
    notes: '',
    dj_confirmation: null,
  },
  event: {
    id: 'evt-1',
    name: 'Spring Showcase',
    venue_name: 'Revolver',
    start_date: '2026-07-12',
    end_date: '2026-07-12',
    genres: [],
  },
};

describe('CardComponent (EL-049)', () => {
  let slotResponse$: any;
  let slotId = 'slot-1';

  const apiMock = { getPublicSlot: () => slotResponse$ };

  async function build(): Promise<{ fixture: ComponentFixture<CardComponent>; component: CardComponent }> {
    await TestBed.configureTestingModule({
      imports: [CardComponent],
      providers: [
        provideTranslateService(),
        provideRouter([]),
        { provide: ApiService, useValue: apiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => slotId } } } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(CardComponent);
    const component = fixture.componentInstance;
    fixture.detectChanges();
    return { fixture, component };
  }

  beforeEach(() => {
    slotId = 'slot-1';
    slotResponse$ = of(PUBLIC_SLOT);
  });

  it('derives the card fields from the slot + event', async () => {
    const { component } = await build();
    expect(component.djName()).toBe('Cody');
    expect(component.genre()).toBe('House');
    expect(component.stage()).toBe('Main Stage');
    expect(component.time()).toBe('22:00');
    expect(component.eventName()).toBe('Spring Showcase');
    expect(component.venue()).toBe('Revolver');
    expect(component.eventId()).toBe('evt-1');
    expect(component.loading()).toBe(false);
    expect(component.notFound()).toBe(false);
  });

  it('formats the date in brand English abbreviations, timezone-stable', async () => {
    const { component } = await build();
    // 2026-07-12 is a Sunday; must read the same regardless of the host timezone.
    expect(component.dateLabel()).toBe('SUN 12 JUL');
    expect(component.year()).toBe('2026');
  });

  it('shrinks the headline size as the DJ name gets longer', async () => {
    const { component } = await build();
    expect(component.nameSizeCqw()).toBe(23); // "Cody" (4 chars)
    component.data.set({ ...PUBLIC_SLOT, slot: { ...PUBLIC_SLOT.slot, dj_name: 'DJ Anthony' } });
    expect(component.nameSizeCqw()).toBe(10); // 10 chars
  });

  it('maps the genre to a theme palette class (EL-052)', async () => {
    const { component } = await build();
    expect(component.themeClass()).toBe('theme-house');
    component.data.set({ ...PUBLIC_SLOT, slot: { ...PUBLIC_SLOT.slot, genre: 'Hip-Hop' } });
    expect(component.themeClass()).toBe('theme-hiphop');
    // Aliased genre rides the soul palette.
    component.data.set({ ...PUBLIC_SLOT, slot: { ...PUBLIC_SLOT.slot, genre: 'R&B' } });
    expect(component.themeClass()).toBe('theme-soul');
    // Unknown / empty genre → brand default (no theme class).
    component.data.set({ ...PUBLIC_SLOT, slot: { ...PUBLIC_SLOT.slot, genre: 'Reggae' } });
    expect(component.themeClass()).toBe('');
    component.data.set({ ...PUBLIC_SLOT, slot: { ...PUBLIC_SLOT.slot, genre: '' } });
    expect(component.themeClass()).toBe('');
  });

  it('produces a spiral polyline', async () => {
    const { component } = await build();
    const pts = component.spiralPoints().split(' ');
    expect(pts.length).toBe(241); // 240 steps inclusive
    expect(pts[0]).toBe('0.0,0.0');
  });

  it('shares the backend OG link (/s/dj/:id) to LINE, not the SPA route', async () => {
    const { component } = await build();
    const open = vi.spyOn(window, 'open').mockImplementation(() => null);
    component.shareViaLine();
    expect(open).toHaveBeenCalledTimes(1);
    const url = open.mock.calls[0][0] as string;
    expect(url).toContain('line.me/R/msg/text/');
    expect(decodeURIComponent(url)).toContain('/s/dj/slot-1');
    open.mockRestore();
  });

  it('copies the share link to the clipboard and flags copied', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, { clipboard: { writeText } });
    const { component } = await build();
    await component.copyLink();
    expect(writeText).toHaveBeenCalledTimes(1);
    expect(writeText.mock.calls[0][0]).toContain('/s/dj/slot-1');
    expect(component.copied()).toBe(true);
  });

  it('shows not-found when the slot lookup fails', async () => {
    slotResponse$ = throwError(() => ({ status: 404 }));
    const { component } = await build();
    expect(component.notFound()).toBe(true);
    expect(component.loading()).toBe(false);
  });

  it('shows not-found when no slotId is in the route', async () => {
    slotId = '';
    const { component } = await build();
    expect(component.notFound()).toBe(true);
  });

  describe('save image (EL-053)', () => {
    it('builds a slugged PNG filename from the DJ + event', async () => {
      const { component } = await build();
      expect(component.fileName()).toBe('cody-spring-showcase.png');
      component.data.set({
        ...PUBLIC_SLOT,
        slot: { ...PUBLIC_SLOT.slot, dj_name: 'DJ Anthony!' },
        event: { ...PUBLIC_SLOT.event, name: 'Summer / Bash 2026' },
      });
      expect(component.fileName()).toBe('dj-anthony-summer-bash-2026.png');
    });

    it('exports the card DOM to a PNG and triggers a download', async () => {
      const { component } = await build();
      const clicked: HTMLAnchorElement[] = [];
      const realCreate = document.createElement.bind(document);
      vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
        const el = realCreate(tag) as HTMLElement;
        if (tag === 'a') {
          (el as HTMLAnchorElement).click = () => clicked.push(el as HTMLAnchorElement);
        }
        return el as any;
      });

      await component.saveImage();

      expect(toPng).toHaveBeenCalledTimes(1);
      // Scales up to the 1080px export width and busts the cache.
      expect((toPng as any).mock.calls[0][1]).toMatchObject({ cacheBust: true });
      expect(clicked.length).toBe(1);
      expect(clicked[0].download).toBe('cody-spring-showcase.png');
      expect(clicked[0].href).toContain('data:image/png');
      expect(component.exporting()).toBe(false);
      vi.restoreAllMocks();
    });
  });
});
