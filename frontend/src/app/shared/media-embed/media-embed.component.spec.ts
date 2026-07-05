import { TestBed } from '@angular/core/testing';

import { MediaEmbedComponent, buildEmbedUrl } from './media-embed.component';

describe('buildEmbedUrl', () => {
  it('builds a fixed-template embed URL per provider', () => {
    expect(buildEmbedUrl('youtube', 'dQw4w9WgXcQ')).toBe(
      'https://www.youtube.com/embed/dQw4w9WgXcQ',
    );
    expect(buildEmbedUrl('soundcloud', 'https://soundcloud.com/a/mix')).toBe(
      'https://w.soundcloud.com/player/?url=' + encodeURIComponent('https://soundcloud.com/a/mix'),
    );
    expect(buildEmbedUrl('mixcloud', 'https://www.mixcloud.com/a/show/')).toBe(
      'https://player.mixcloud.com/widget/iframe/?feed=' +
        encodeURIComponent('https://www.mixcloud.com/a/show/'),
    );
  });

  it('returns null for an unknown provider or an empty ref (caller shows a fallback)', () => {
    expect(buildEmbedUrl('vimeo', 'abc')).toBeNull();
    expect(buildEmbedUrl('youtube', '')).toBeNull();
  });

  it('encodes the ref so it cannot break out of the URL template', () => {
    const url = buildEmbedUrl('youtube', 'a"><script>');
    expect(url).not.toContain('<script>');
    expect(url).toContain(encodeURIComponent('a"><script>'));
  });
});

describe('MediaEmbedComponent', () => {
  function render(media: { media_kind: string; media_provider: string; media_ref: string }) {
    const fixture = TestBed.createComponent(MediaEmbedComponent);
    fixture.componentRef.setInput('media', media);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  it('renders an iframe for an embeddable submission', () => {
    const el = render({ media_kind: 'embed', media_provider: 'youtube', media_ref: 'dQw4w9WgXcQ' });
    const frame = el.querySelector('[data-testid="media-embed-frame"]');
    expect(frame).toBeTruthy();
    expect(frame?.tagName.toLowerCase()).toBe('iframe');
  });

  it('renders a fallback for an unknown provider', () => {
    const el = render({ media_kind: 'embed', media_provider: 'vimeo', media_ref: '123' });
    expect(el.querySelector('[data-testid="media-embed-fallback"]')).toBeTruthy();
    expect(el.querySelector('[data-testid="media-embed-frame"]')).toBeNull();
  });

  it('renders a fallback for a non-embed media kind (forward-compat seam)', () => {
    const el = render({ media_kind: 'file', media_provider: '', media_ref: 'x' });
    expect(el.querySelector('[data-testid="media-embed-fallback"]')).toBeTruthy();
  });
});
