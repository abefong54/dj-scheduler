import { Component, contentChildren, input, TemplateRef } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslatePipe } from '@ngx-translate/core';
import { ColumnDefDirective } from './column-def.directive';
import { TableSort } from './table-sort';

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

  private table = new TableSort<Record<string, unknown>>(() => this.data(), {
    searchKeys: () => this.columns().filter(c => c.searchable !== false).map(c => c.key),
  });

  searchQuery = this.table.search;
  sortKey = this.table.sortKey;
  sortDir = this.table.sortDir;
  filteredData = this.table.view;

  sort(col: TableColumn) {
    if (col.sortable === false) return;
    this.table.toggle(col.key);
  }

  getTemplate(key: string): TemplateRef<{ row: Record<string, unknown> }> | null {
    return this.defs().find(d => d.columnKey() === key)?.template ?? null;
  }
}
