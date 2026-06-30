import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTranslateService } from '@ngx-translate/core';

import { LoginComponent } from './login.component';
import { AuthService } from '../services/auth.service';

describe('LoginComponent (EL-062 restyle)', () => {
  let signInWithGoogle: ReturnType<typeof vi.fn>;

  async function build(): Promise<ComponentFixture<LoginComponent>> {
    signInWithGoogle = vi.fn();
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideTranslateService(),
        { provide: AuthService, useValue: { signInWithGoogle } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(LoginComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('renders the Console dark console card with the EventLineup wordmark', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const wordmark = el.querySelector('h1');
    expect(wordmark?.textContent).toContain('EventLineup');
    expect(wordmark?.className).toContain('font-display');

    // Dark, full-height field with the violet void gradient.
    const section = el.querySelector('section');
    expect(section?.getAttribute('style')).toContain('var(--void)');
  });

  it('keeps Google OAuth as the working primary action', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const googleBtn = el.querySelector<HTMLButtonElement>('[data-testid="google-signin"]');
    expect(googleBtn).toBeTruthy();
    expect(googleBtn?.disabled).toBe(false);

    googleBtn!.click();
    expect(signInWithGoogle).toHaveBeenCalledTimes(1);
  });

  it('renders the magic-link control as visual-only / disabled (no backend)', async () => {
    const fixture = await build();
    const el = fixture.nativeElement as HTMLElement;

    const email = el.querySelector<HTMLInputElement>('[data-testid="magic-link-email"]');
    const submit = el.querySelector<HTMLButtonElement>('[data-testid="magic-link-submit"]');
    expect(email?.disabled).toBe(true);
    expect(submit?.disabled).toBe(true);
  });
});
