import { Component, effect, inject, signal } from '@angular/core';
import { DialogService } from './dialog.service';
import { TranslatePipe } from '@ngx-translate/core';
import { ButtonComponent } from './button.component';

@Component({
  selector: 'app-dialog',
  standalone: true,
  imports: [TranslatePipe, ButtonComponent],
  templateUrl: './dialog.component.html',
  styleUrl: './dialog.component.css',
})
export class DialogComponent {
  dialog = inject(DialogService);

  /** True once the share-dialog URL has been copied (drives the "Copied!" label). */
  protected readonly copied = signal(false);

  constructor() {
    // Reset the copied state whenever the dialog opens or closes, so a reused
    // dialog never shows a stale "Copied!".
    effect(() => {
      this.dialog.state();
      this.copied.set(false);
    });
  }

  protected async copyLink() {
    const url = this.dialog.state()?.url;
    if (!url) return;
    await navigator.clipboard.writeText(url);
    this.copied.set(true);
  }

  protected selectAll(event: Event) {
    (event.target as HTMLInputElement).select();
  }
}
