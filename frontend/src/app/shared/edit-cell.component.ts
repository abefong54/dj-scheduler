import { Component, ElementRef, input, output, signal, viewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';

/**
 * Reusable inline-edit cell. Controlled component: it renders projected display
 * content, and on activate swaps to an editor. Enter/blur commit, Esc reverts.
 * It never mutates its own displayed value — it emits (commit) and the parent
 * re-renders with new data on success (so an error path reverts automatically).
 */
@Component({
  selector: 'app-edit-cell',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './edit-cell.component.html',
  styleUrl: './edit-cell.component.css',
})
export class EditCellComponent {
  readonly value = input<string>('');
  readonly endValue = input<string>('');
  readonly type = input<'text' | 'date-range' | 'genres'>('text');
  readonly required = input<boolean>(false);
  readonly label = input<string>('');
  readonly trigger = input<'click' | 'pencil'>('click');
  readonly disabled = input<boolean>(false);

  readonly commit = output<{ value: string; endValue: string }>();

  readonly editing = signal(false);
  readonly draft = signal('');
  readonly draftEnd = signal('');

  private readonly editor = viewChild<ElementRef<HTMLInputElement>>('editor');

  startEdit() {
    if (this.disabled() || this.editing()) return;
    this.draft.set(this.value());
    this.draftEnd.set(this.endValue());
    this.editing.set(true);
    setTimeout(() => {
      const el = this.editor()?.nativeElement;
      el?.focus();
      el?.select?.();
    });
  }

  onKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      this.commitEdit();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      this.cancelEdit();
    }
  }

  /** focusout on the editor container: commit only when focus leaves the cell. */
  onFocusOut(e: FocusEvent) {
    const container = e.currentTarget as HTMLElement;
    const next = e.relatedTarget as Node | null;
    if (next && container.contains(next)) return; // moved between the two date inputs
    this.commitEdit();
  }

  commitEdit() {
    if (!this.editing()) return;
    if (!this.isValid() || this.isUnchanged()) {
      this.editing.set(false);
      return;
    }
    this.editing.set(false);
    this.commit.emit({ value: this.draft(), endValue: this.draftEnd() });
  }

  cancelEdit() {
    this.editing.set(false);
  }

  private isUnchanged(): boolean {
    return this.draft() === this.value() && this.draftEnd() === this.endValue();
  }

  private isValid(): boolean {
    if (this.type() === 'date-range') {
      const start = this.draft().trim();
      if (!start) return false;
      const end = this.draftEnd().trim();
      return end === '' || end >= start;
    }
    return !this.required() || this.draft().trim().length > 0;
  }
}
