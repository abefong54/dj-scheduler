import { Component, computed, input } from '@angular/core';
import { IconComponent, IconName } from './icon.component';

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'success' | 'ghost';
export type ButtonSize = 'md' | 'sm' | 'xs' | 'icon';

const VARIANT_CLASS: Record<ButtonVariant, string> = {
  primary: 'btn-primary',
  secondary: 'btn-secondary',
  danger: 'btn-danger',
  success: 'btn-success',
  ghost: 'btn-ghost',
};

const SIZE_CLASS: Record<ButtonSize, string> = {
  md: 'btn-md',
  sm: 'btn-sm',
  xs: 'btn-xs',
  icon: 'btn-icon',
};

@Component({
  selector: 'app-button',
  standalone: true,
  imports: [IconComponent],
  host: {
    '[class.block]': 'fullWidth()',
    '[class.w-full]': 'fullWidth()',
  },
  template: `
    <button
      [type]="type()"
      [disabled]="disabled()"
      [class]="classes()"
      [attr.data-testid]="testId() || null"
      [attr.aria-label]="accessibleLabel()"
      [attr.title]="iconOnly() ? label() || null : null"
    >
      @if (icon()) {
        <app-icon [name]="icon()!" class="w-4 h-4" />
      }
      @if (!iconOnly() && label()) {
        <span>{{ label() }}</span>
      }
    </button>
  `,
})
export class ButtonComponent {
  variant = input<ButtonVariant>('secondary');
  size = input<ButtonSize>('md');
  icon = input<IconName | null>(null);
  label = input<string>('');
  iconOnly = input<boolean>(false);
  type = input<'button' | 'submit'>('button');
  disabled = input<boolean>(false);
  testId = input<string>('');
  ariaLabel = input<string>('');
  fullWidth = input<boolean>(false);

  classes = computed(() => {
    const base = `btn ${VARIANT_CLASS[this.variant()]} ${SIZE_CLASS[this.size()]}`;
    return this.fullWidth() ? `${base} w-full` : base;
  });

  /** Accessible name: explicit override, else the label when the text is hidden. */
  accessibleLabel = computed<string | null>(
    () => this.ariaLabel() || (this.iconOnly() ? this.label() : '') || null,
  );
}
