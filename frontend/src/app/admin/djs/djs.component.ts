import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';
import { ApiService, DJ } from '../../services/api.service';
import { DjsStore } from '../../services/djs-store';
import { DialogService } from '../../shared/dialog.service';
import { LanguageService } from '../../services/language.service';
import { DataTableComponent, TableColumn } from '../../shared/data-table.component';
import { ColumnDefDirective } from '../../shared/column-def.directive';
import { ButtonComponent } from '../../shared/button.component';
import { AdminShellComponent } from '../../shared/admin-shell.component';
import { StatusBadgeComponent } from '../../shared/status-badge.component';
import { LoadingMeterComponent } from '../../shared/loading-meter.component';
import { SubmitGuard } from '../../shared/submit-guard';

// CERT_OPTIONS are the school's standard curriculum genres (EL-020). DJs can
// also have custom (free-text) certifications beyond this list.
export const CERT_OPTIONS = ['House', 'Hip Hop', 'Pop', 'Techno', 'Trance', 'Drum & Bass', 'R&B'];

@Component({
  selector: 'app-djs',
  standalone: true,
  imports: [FormsModule, TranslatePipe, DataTableComponent, ColumnDefDirective, ButtonComponent, AdminShellComponent, StatusBadgeComponent, LoadingMeterComponent],
  templateUrl: './djs.component.html',
  styleUrl: './djs.component.css',
})
export class DjsComponent {
  private api = inject(ApiService);
  private translate = inject(TranslateService);
  private dialog = inject(DialogService);
  private langService = inject(LanguageService);
  private store = inject(DjsStore);

  // Backed by the root DjsStore (EL-092): a re-visit paints from cache with no
  // spinner while the roster revalidates in the background.
  djs = this.store.djs;
  loading = this.store.loading;
  loadError = this.store.error;
  readonly addGuard = new SubmitGuard();
  readonly saveGuard = new SubmitGuard();
  // Add-a-DJ form state. The form lives in a modal opened from the topbar
  // "+ Add" button (EL-099), not an always-visible inline card.
  showAddModal = signal(false);
  newDJ = signal({ name: '' });
  newDJGenres = signal('');

  // EL-020: certifications edit panel. `editing` holds the DJ being edited (the
  // panel is open when non-null); the rest are its working-copy form fields.
  readonly certOptions = CERT_OPTIONS;
  editing = signal<DJ | null>(null);
  editName = signal('');
  editIsStudent = signal(true);
  editCerts = signal<string[]>([]);
  editCustom = signal('');

  constructor() {
    this.store.ensure();
  }

  retry() {
    this.store.ensure();
  }

  /** Open the add-DJ modal from the topbar "+ Add" button, with a fresh form. */
  openAddModal() {
    this.newDJ.set({ name: '' });
    this.newDJGenres.set('');
    this.showAddModal.set(true);
  }

  cancelAdd() {
    this.showAddModal.set(false);
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

    this.addGuard.run(this.api.createDJ({ name, genre_tags: genres }), {
      next: () => {
        this.newDJ.set({ name: '' });
        this.newDJGenres.set('');
        this.showAddModal.set(false);
        this.store.refresh();
      },
    });
  }

  copyPortalLink(id: string) {
    this.api.generateDJPortalToken(id).subscribe(({ portal_url, expires_at }) => {
      this.dialog.shareLink({
        title: this.translate.instant('dj.portal_link.title'),
        message: this.translate.instant('dj.portal_link.warning'),
        url: portal_url,
        expiresLabel: expires_at
          ? this.translate.instant('dj.portal_link.expires', { date: this.formatExpiry(expires_at) })
          : undefined,
        copyLabel: this.translate.instant('dj.portal_link.copy'),
        copiedLabel: this.translate.instant('dj.portal_link.copied'),
      });
    });
  }

  /** Formats an ISO timestamp to a short local date like "17 Jul 2026". */
  private formatExpiry(iso: string): string {
    const d = new Date(iso);
    return isNaN(d.getTime())
      ? iso
      : d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  }

  async deleteDJ(id: string) {
    const ok = await this.dialog.confirm({
      title: this.translate.instant('dialog.deleteTitle'),
      message: this.translate.instant('djs.deleteConfirm'),
      confirmLabel: this.translate.instant('actions.delete'),
      variant: 'danger',
    });
    if (!ok) return;
    this.api.deleteDJ(id).subscribe(() => this.store.refresh());
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
    this.saveGuard.run(this.api.updateDJ(dj.id, {
      name: this.editName().trim(),
      genre_tags: dj.genre_tags,
      certifications: this.editCerts(),
      is_student: this.editIsStudent(),
    }), {
      next: () => {
        this.cancelEdit();
        this.store.refresh();
      },
    });
  }

  cancelEdit() {
    this.editing.set(null);
  }
}
