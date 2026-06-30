import { Component, inject } from '@angular/core';
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
}
