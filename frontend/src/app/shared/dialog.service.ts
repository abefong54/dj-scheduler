import { Injectable, signal } from '@angular/core';

export interface DialogConfig {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'default';
}

/**
 * ShareLinkConfig drives the `share` dialog: a read-only URL the user can copy as
 * many times as they like, an optional expiry line, and an explanatory/warning
 * message (in `message`). Used for the DJ portal link (EL-089), where the raw
 * link only exists at generation time and re-generating revokes the previous one.
 */
export interface ShareLinkConfig {
  title: string;
  message: string;
  url: string;
  expiresLabel?: string;
  copyLabel?: string;
  copiedLabel?: string;
  closeLabel?: string;
}

interface DialogState {
  type: 'confirm' | 'alert' | 'share';
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'default';
  url?: string;
  expiresLabel?: string;
  copyLabel?: string;
  copiedLabel?: string;
  closeLabel?: string;
  resolve: (value: boolean) => void;
}

@Injectable({ providedIn: 'root' })
export class DialogService {
  readonly state = signal<DialogState | null>(null);

  confirm(config: DialogConfig): Promise<boolean> {
    return new Promise(resolve => {
      this.state.set({ ...config, type: 'confirm', resolve });
    });
  }

  alert(config: Omit<DialogConfig, 'cancelLabel'>): Promise<void> {
    return new Promise(resolve => {
      this.state.set({
        ...config,
        type: 'alert',
        resolve: () => resolve(),
      });
    });
  }

  /** Opens a dialog showing a copyable link. Resolves when dismissed. */
  shareLink(config: ShareLinkConfig): Promise<void> {
    return new Promise(resolve => {
      this.state.set({
        ...config,
        type: 'share',
        resolve: () => resolve(),
      });
    });
  }

  respond(value: boolean) {
    this.state()?.resolve(value);
    this.state.set(null);
  }
}
