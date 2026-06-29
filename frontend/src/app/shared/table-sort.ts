import { Signal, computed, signal } from '@angular/core';

/**
 * Reactive search + sort engine shared by every admin table. Holds the search
 * query and sort state as signals and exposes a `view` computed that filters
 * then sorts the source rows. Framework-light: no DI, instantiate it directly
 * in any component (see DataTableComponent and the event-detail slots table).
 */
export interface TableSortConfig<T> {
  /** Reactive list of field keys the text filter scans (case-insensitive). */
  searchKeys: () => string[];
  /** How to stringify a row for sorting; defaults to String(row[key] ?? ''). */
  sortValue?: (row: T, key: string) => string;
  /** Column key to sort by initially; omit for source order. */
  initialSortKey?: string;
}

export class TableSort<T> {
  readonly search = signal('');
  readonly sortKey = signal('');
  readonly sortDir = signal<'asc' | 'desc'>('asc');
  readonly view: Signal<T[]>;

  constructor(rows: () => T[], cfg: TableSortConfig<T>) {
    if (cfg.initialSortKey) this.sortKey.set(cfg.initialSortKey);
    const sortValue =
      cfg.sortValue ?? ((row: T, key: string) => String((row as Record<string, unknown>)[key] ?? ''));

    this.view = computed(() => {
      const q = this.search().toLowerCase().trim();
      const key = this.sortKey();
      const dir = this.sortDir();

      let out = rows();
      if (q) {
        const keys = cfg.searchKeys();
        out = out.filter(row =>
          keys.some(k => String((row as Record<string, unknown>)[k] ?? '').toLowerCase().includes(q)),
        );
      }
      if (key) {
        out = [...out].sort((a, b) => {
          const av = sortValue(a, key);
          const bv = sortValue(b, key);
          return dir === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av);
        });
      }
      return out;
    });
  }

  /** Sort by `key`; same key flips direction, a new key restarts at ascending. */
  toggle(key: string) {
    if (this.sortKey() === key) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortKey.set(key);
      this.sortDir.set('asc');
    }
  }

  /** Sort arrow for a header: '↑' / '↓' when active on `key`, else ''. */
  indicator(key: string): string {
    if (this.sortKey() !== key) return '';
    return this.sortDir() === 'asc' ? '↑' : '↓';
  }
}
