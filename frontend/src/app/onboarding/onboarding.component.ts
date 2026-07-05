import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { TranslatePipe } from '@ngx-translate/core';
import { ApiService, Member } from '../services/api.service';
import { SessionStore } from '../services/session.store';

/**
 * Two-step onboarding wizard (EL-106).
 *
 * Step 1 — Register school: the user provides a name (required) and city
 *   (optional) then calls POST /api/schools. On success, session.refresh()
 *   is called so the school signal propagates to AdminShell immediately.
 *
 * Step 2 — Add team: the user can invite colleagues by email + role. Each
 *   invite calls POST /api/schools/current/members. The current member list
 *   is fetched from GET /api/schools/current/members and rendered. The user
 *   can add as many members as they like, then click Done to navigate to
 *   /admin/events.
 */
@Component({
  selector: 'app-onboarding',
  standalone: true,
  imports: [FormsModule, TranslatePipe],
  templateUrl: './onboarding.component.html',
})
export class OnboardingComponent {
  private readonly api = inject(ApiService);
  private readonly session = inject(SessionStore);
  private readonly router = inject(Router);

  /** 1 = register school, 2 = add team. */
  protected readonly step = signal<1 | 2>(1);

  // ── Step 1 fields ──────────────────────────────────────────────────────────
  protected readonly schoolName = signal('');
  protected readonly schoolCity = signal('');
  protected readonly registering = signal(false);
  protected readonly registerError = signal('');

  // ── Step 2 fields ──────────────────────────────────────────────────────────
  protected readonly members = signal<Member[]>([]);
  protected readonly memberEmail = signal('');
  protected readonly memberRole = signal('member');
  protected readonly adding = signal(false);
  protected readonly addError = signal('');

  /** Submit step 1: create the school, refresh session, advance to step 2. */
  registerSchool(): void {
    if (!this.schoolName()) return;
    this.registering.set(true);
    this.registerError.set('');

    const body: { name: string; city?: string } = { name: this.schoolName() };
    if (this.schoolCity()) body.city = this.schoolCity();

    this.api.createSchool(body).subscribe({
      next: () => {
        this.session.refresh();
        this._loadMembers();
        this.step.set(2);
        this.registering.set(false);
      },
      error: () => {
        this.registerError.set('onboarding.error.generic');
        this.registering.set(false);
      },
    });
  }

  /** Add a member row in step 2. */
  addMember(): void {
    if (!this.memberEmail()) return;
    this.adding.set(true);
    this.addError.set('');

    this.api.inviteMember({ email: this.memberEmail(), role: this.memberRole() }).subscribe({
      next: (member: Member) => {
        this.members.update((m) => [...m, member]);
        this.memberEmail.set('');
        this.adding.set(false);
      },
      error: () => {
        this.addError.set('onboarding.error.generic');
        this.adding.set(false);
      },
    });
  }

  /** Navigate to the admin console when onboarding is done. */
  done(): void {
    this.router.navigate(['/admin/events']);
  }

  private _loadMembers(): void {
    this.api.listMembers().subscribe({
      next: (members: Member[]) => this.members.set(members),
    });
  }
}
