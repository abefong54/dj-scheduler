import { Injectable, signal } from '@angular/core';

export interface DialogConfig {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'default';
}

interface DialogState extends DialogConfig {
  type: 'confirm' | 'alert';
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

  respond(value: boolean) {
    this.state()?.resolve(value);
    this.state.set(null);
  }
}
