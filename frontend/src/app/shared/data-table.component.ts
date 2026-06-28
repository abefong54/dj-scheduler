import { Component, contentChildren, input, signal, computed, TemplateRef } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';
import { ColumnDefDirective } from './column-def.directive';

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;   // default: true
  searchable?: boolean; // default: true — participates in text filter
}

@Component({
  selector: 'app-data-table',
  standalone: true,
  imports: [FormsModule, NgTemplateOutlet, TranslatePipe, ColumnDefDirective],
  templateUrl: './data-table.component.html',
  styleUrl: './data-table.component.css',
})
export class DataTableComponent {
  columns = input<TableColumn[]>([]);
  data = input<Record<string, unknown>[]>([]);

  private defs = contentChildren(ColumnDefDirective);

  searchQuery = signal('');
  sortKey = signal('');
  sortDir = signal<'asc' | 'desc'>('asc');

  filteredData = computed(() => {
    const q = this.searchQuery().toLowerCase().trim();
    const key = this.sortKey();
    const dir = this.sortDir();
    const searchableCols = this.columns()
      .filter(c => c.searchable !== false)
      .map(c => c.key);

    let rows = this.data();

    if (q) {
      rows = rows.filter(row =>
        searchableCols.some(k => String(row[k] ?? '').toLowerCase().includes(q))
      );
    }

    if (key) {
      rows = [...rows].sort((a, b) => {
        const av = String(a[key] ?? '');
        const bv = String(b[key] ?? '');
        return dir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av);
      });
    }

    return rows;
  });

  sort(col: TableColumn) {
    if (col.sortable === false) return;
    if (this.sortKey() === col.key) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortKey.set(col.key);
      this.sortDir.set('asc');
    }
  }

  getTemplate(key: string): TemplateRef<{ row: Record<string, unknown> }> | null {
    return this.defs().find(d => d.columnKey() === key)?.template ?? null;
  }
}
