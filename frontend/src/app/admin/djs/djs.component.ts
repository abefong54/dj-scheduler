import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, DJ } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';
import { LanguageService } from '../../services/language.service';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [FormsModule, TranslatePipe, DataTableComponent, ColumnDefDirective],
  templateUrl: './djs.component.html',
  styleUrl: './djs.component.css',
})
export class DjsComponent implements OnDestroy {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private langService = inject(LanguageService);

  djs = signal<DJ[]>([]);
  newDJ = signal({ name: '' });
  newDJGenres = signal('');

  private subscriptions: Subscription[] = [];

  constructor() {
    this.loadDJs();
  }

  private loadDJs() {
    this.subscriptions.push(
      this.api.getDJs().subscribe(djs => this.djs.set(djs))
    );
  }

  // Convert DJ[] to Record<string, unknown>[] for DataTableComponent
  djsAsRows = computed((): Record<string, unknown>[] =>
    this.djs().map(d => ({ id: d.id, name: d.name, genre_tags: d.genre_tags }) as Record<string, unknown>)
  );

  // Reactive column labels — re-evaluates when language changes
  djColumns = computed((): TableColumn[] => {
    this.langService.currentLang(); // register dependency so labels update on language switch
    return [
      { key: 'name', label: this.translate.instant('djs.name') },
      { key: 'genre_tags', label: this.translate.instant('djs.genres'), sortable: false },
      { key: 'actions', label: this.translate.instant('actions.title'), sortable: false, searchable: false },
    ];
  });

  genreTags(row: Record<string, unknown>): string[] {
    return Array.isArray(row['genre_tags']) ? (row['genre_tags'] as string[]) : [];
  }

  rowId(row: Record<string, unknown>): string {
    return String(row['id']);
  }

  async addDJ() {
    const name = this.newDJ().name;
    const genres = this.newDJGenres()
      .split(',')
      .map(g => g.trim())
      .filter(g => g.length > 0);

    if (!name || genres.length === 0) {
      await this.dialog.alert({
        title: this.translate.instant('dialog.validationTitle'),
        message: this.translate.instant('djs.fillRequired'),
      });
      return;
    }

    this.api.createDJ({ name, genre_tags: genres }).subscribe(() => {
      this.newDJ.set({ name: '' });
      this.newDJGenres.set('');
      this.loadDJs();
    });
  }

  async deleteDJ(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('djs.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    this.api.deleteDJ(id).subscribe(() => this.loadDJs());
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
