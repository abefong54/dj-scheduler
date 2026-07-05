import { Component, inject } from '@angular/core';
import { ChallengesService } from '../services/challenges.service';

/**
 * Student challenge surfaces (My Challenges + Browse Core). SCAFFOLD shell:
 * injects ChallengesService so the wiring is real; Track G (student surfaces)
 * fills My Challenges, Browse Core, and the post-attempt flow.
 */
@Component({
  selector: 'app-student-challenges',
  standalone: true,
  template: `
    <section style="padding:2rem;">
      <h1>My Challenges</h1>
      <p>SCAFFOLD — not implemented.</p>
    </section>
  `,
})
export class StudentChallengesComponent {
  protected challenges = inject(ChallengesService);
}
