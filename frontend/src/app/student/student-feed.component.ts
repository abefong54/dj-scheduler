import { Component, inject } from '@angular/core';
import { ChallengesService } from '../services/challenges.service';

/**
 * School feed of passed submissions. SCAFFOLD shell: injects ChallengesService so
 * the wiring is real; Track G (student surfaces) fills the feed + media embeds.
 */
@Component({
  selector: 'app-student-feed',
  standalone: true,
  template: `
    <section style="padding:2rem;">
      <h1>Feed</h1>
      <p>SCAFFOLD — not implemented.</p>
    </section>
  `,
})
export class StudentFeedComponent {
  protected challenges = inject(ChallengesService);
}
