import { TestBed } from '@angular/core/testing';
import { ThemeService } from './theme.service';

describe('ThemeService', () => {
  let service: ThemeService;

  beforeEach(() => {
    document.documentElement.removeAttribute('data-mode');
    TestBed.configureTestingModule({});
    service = TestBed.inject(ThemeService);
  });

  afterEach(() => {
    document.documentElement.removeAttribute('data-mode');
  });

  it('defaults to dark "booth" mode with no data-mode attribute', () => {
    TestBed.tick();
    expect(service.mode()).toBe('dark');
    expect(document.documentElement.hasAttribute('data-mode')).toBe(false);
  });

  it('setMode("light") writes data-mode="light" to the app root', () => {
    service.setMode('light');
    TestBed.tick();
    expect(document.documentElement.getAttribute('data-mode')).toBe('light');
  });

  it('toggleMode flips between dark and light and clears the attribute on dark', () => {
    service.toggleMode();
    TestBed.tick();
    expect(service.mode()).toBe('light');
    expect(document.documentElement.getAttribute('data-mode')).toBe('light');

    service.toggleMode();
    TestBed.tick();
    expect(service.mode()).toBe('dark');
    expect(document.documentElement.hasAttribute('data-mode')).toBe(false);
  });
});
