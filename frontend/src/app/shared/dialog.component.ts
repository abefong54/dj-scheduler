import { Component, inject } from '@angular/core';
import { DialogService } from './dialog.service';
import { TranslatePipe } from '@ngx-translate/core';

@Component({
  selector: 'app-dialog',
  standalone: true,
  imports: [TranslatePipe],
  templateUrl: './dialog.component.html',
  styleUrl: './dialog.component.css',
})
export class DialogComponent {
  dialog = inject(DialogService);
}
