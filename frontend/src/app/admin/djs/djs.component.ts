import { Component, computed, inject, signal, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { ApiService, DJ } from '../../services/api.service';
import { DialogService } from '../../shared/dialog.service';
import { LanguageService } from '../../services/language.service';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';

// CERT_OPTIONS are the school's standard curriculum genres (EL-020). DJs can
// also have custom (free-text) certifications beyond this list.
export const CERT_OPTIONS = ['House', 'Hip Hop', 'Pop', 'Techno', 'Trance', 'Drum & Bass', 'R&B'];

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
  // ID of the DJ whose portal link was just copied (drives the "Copied!" flash).
  copiedDJId = signal<string | null>(null);

  // EL-020: certifications edit panel. `editing` holds the DJ being edited (the
  // panel is open when non-null); the rest are its working-copy form fields.
  readonly certOptions = CERT_OPTIONS;
  editing = signal<DJ | null>(null);
  editName = signal('');
  editIsStudent = signal(true);
  editCerts = signal<string[]>([]);
  editCustom = signal('');

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
    this.djs().map(d => ({
      id: d.id,
      name: d.name,
      genre_tags: d.genre_tags,
      certifications: d.certifications ?? [],
      is_student: d.is_student ?? true,
    }) as Record<string, unknown>)
  );

  // Reactive column labels — re-evaluates when language changes
  djColumns = computed((): TableColumn[] => {
    this.langService.currentLang(); // register dependency so labels update on language switch
    return [
      { key: 'name', label: this.translate.instant('djs.name') },
      { key: 'genre_tags', label: this.translate.instant('djs.genres'), sortable: false },
      { key: 'certifications', label: this.translate.instant('djs.certifications'), sortable: false },
      { key: 'actions', label: this.translate.instant('actions.title'), sortable: false, searchable: false },
    ];
  });

  certs(row: Record<string, unknown>): string[] {
    return Array.isArray(row['certifications']) ? (row['certifications'] as string[]) : [];
  }

  isStudent(row: Record<string, unknown>): boolean {
    return row['is_student'] !== false;
  }

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

  copyPortalLink(id: string) {
    this.api.generateDJPortalToken(id).subscribe(async ({ portal_url }) => {
      await navigator.clipboard.writeText(portal_url);
      this.copiedDJId.set(id);
      setTimeout(() => {
        if (this.copiedDJId() === id) this.copiedDJId.set(null);
      }, 2000);
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

  // --- EL-020: certifications edit panel ---

  // openEdit loads a DJ into the edit panel from the table row.
  openEdit(row: Record<string, unknown>) {
    const dj = this.djs().find(d => d.id === this.rowId(row));
    if (!dj) return;
    this.editing.set(dj);
    this.editName.set(dj.name);
    this.editIsStudent.set(dj.is_student ?? true);
    this.editCerts.set([...(dj.certifications ?? [])]);
    this.editCustom.set('');
  }

  isCertSelected(genre: string): boolean {
    return this.editCerts().includes(genre);
  }

  toggleCert(genre: string) {
    this.editCerts.update(certs =>
      certs.includes(genre) ? certs.filter(c => c !== genre) : [...certs, genre]);
  }

  addCustomCert() {
    const value = this.editCustom().trim();
    if (value && !this.editCerts().includes(value)) {
      this.editCerts.update(certs => [...certs, value]);
    }
    this.editCustom.set('');
  }

  saveEdit() {
    const dj = this.editing();
    if (!dj || !this.editName().trim()) return;
    this.api.updateDJ(dj.id, {
      name: this.editName().trim(),
      genre_tags: dj.genre_tags,
      certifications: this.editCerts(),
      is_student: this.editIsStudent(),
    }).subscribe(() => {
      this.cancelEdit();
      this.loadDJs();
    });
  }

  cancelEdit() {
    this.editing.set(null);
  }

  ngOnDestroy() {
    this.subscriptions.forEach(sub => sub.unsubscribe());
  }
}
