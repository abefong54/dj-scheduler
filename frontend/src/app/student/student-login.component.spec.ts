import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';

import { StudentLoginComponent } from './student-login.component';
import { StudentAuthService } from '../services/student-auth.service';

describe('StudentLoginComponent (Track E)', () => {
  let navigate: ReturnType<typeof vi.fn>;
  let setToken: ReturnType<typeof vi.fn>;
  let signInWithGoogle: ReturnType<typeof vi.fn>;
  let isAuthenticated: ReturnType<typeof vi.fn>;

  async function build(
    params: Record<string, string>,
  ): Promise<ComponentFixture<StudentLoginComponent>> {
    navigate = vi.fn();
    setToken = vi.fn();
    signInWithGoogle = vi.fn();
    isAuthenticated = vi.fn(() => false);
    await TestBed.configureTestingModule({
      imports: [StudentLoginComponent],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: convertToParamMap(params) } },
        },
        { provide: Router, useValue: { navigate } },
        { provide: StudentAuthService, useValue: { setToken, signInWithGoogle, isAuthenticated } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(StudentLoginComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('stores the token and forwards to the portal on ?token=', async () => {
    await build({ token: 'jwt-abc' });
    expect(setToken).toHaveBeenCalledWith('jwt-abc');
    expect(navigate).toHaveBeenCalledWith(['/student/challenges']);
  });

  it('shows the roster message on ?error=not_on_roster and does not navigate', async () => {
    const fixture = await build({ error: 'not_on_roster' });
    const err = fixture.nativeElement.querySelector('[data-testid="student-login-error"]');
    expect(err?.textContent).toContain('not on a roster');
    expect(navigate).not.toHaveBeenCalled();
  });

  it('shows the Google sign-in button and triggers OAuth on click', async () => {
    const fixture = await build({});
    const el = fixture.nativeElement as HTMLElement;
    const btn = el.querySelector<HTMLButtonElement>('[data-testid="student-google-signin"]');
    expect(btn).toBeTruthy();
    expect(btn?.className).toContain('btn-primary');
    btn!.click();
    expect(signInWithGoogle).toHaveBeenCalledTimes(1);
  });

  it('skips straight to the portal when already authenticated', async () => {
    const nav = vi.fn();
    await TestBed.configureTestingModule({
      imports: [StudentLoginComponent],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: convertToParamMap({}) } },
        },
        { provide: Router, useValue: { navigate: nav } },
        {
          provide: StudentAuthService,
          useValue: {
            setToken: vi.fn(),
            signInWithGoogle: vi.fn(),
            isAuthenticated: vi.fn(() => true),
          },
        },
      ],
    }).compileComponents();
    TestBed.createComponent(StudentLoginComponent).detectChanges();
    expect(nav).toHaveBeenCalledWith(['/student/challenges']);
  });
});
