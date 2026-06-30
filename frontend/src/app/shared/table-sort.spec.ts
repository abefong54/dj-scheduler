import { signal } from '@angular/core';
import { TableSort } from './table-sort';

interface Row {
  id: string;
  name: string;
  genre: string;
}

const ROWS: Row[] = [
  { id: '1', name: 'DJ Alpha', genre: 'House' },
  { id: '2', name: 'DJ Beta', genre: 'Techno' },
  { id: '3', name: 'MC Gamma', genre: 'House' },
];

function make(rows: Row[] = ROWS, initialSortKey?: string) {
  return new TableSort<Row>(() => rows, {
    searchKeys: () => ['name', 'genre'],
    initialSortKey,
  });
}

describe('TableSort', () => {
  it('returns all rows in source order by default', () => {
    expect(make().view().map(r => r.id)).toEqual(['1', '2', '3']);
  });

  it('filters by search query case-insensitively', () => {
    const t = make();
    t.search.set('dj');
    expect(t.view().map(r => r.name)).toEqual(['DJ Alpha', 'DJ Beta']);
  });

  it('filters across every search key', () => {
    const t = make();
    t.search.set('house');
    expect(t.view().length).toBe(2);
  });

  it('toggle sorts ascending on first call', () => {
    const t = make();
    t.toggle('name');
    expect(t.view()[0].name).toBe('DJ Alpha');
    expect(t.sortDir()).toBe('asc');
  });

  it('toggle reverses to descending on second call of same key', () => {
    const t = make();
    t.toggle('name');
    t.toggle('name');
    expect(t.view()[0].name).toBe('MC Gamma');
    expect(t.sortDir()).toBe('desc');
  });

  it('toggle resets to ascending when the key changes', () => {
    const t = make();
    t.toggle('name');
    t.toggle('name'); // desc
    t.toggle('genre'); // new key
    expect(t.sortDir()).toBe('asc');
    expect(t.sortKey()).toBe('genre');
  });

  it('composes search then sort — sort applies to the filtered set only', () => {
    const t = make();
    t.search.set('dj');
    t.toggle('name');
    t.sortDir.set('desc');
    expect(t.view().map(r => r.name)).toEqual(['DJ Beta', 'DJ Alpha']);
  });

  it('indicator reflects active key and direction', () => {
    const t = make();
    expect(t.indicator('name')).toBe('');
    t.toggle('name');
    expect(t.indicator('name')).toBe('↑');
    t.toggle('name');
    expect(t.indicator('name')).toBe('↓');
    expect(t.indicator('genre')).toBe('');
  });

  it('honors an initial sort key', () => {
    const t = make(ROWS, 'name');
    expect(t.sortKey()).toBe('name');
    expect(t.view()[0].name).toBe('DJ Alpha');
  });

  it('reacts to a changing reactive source', () => {
    const src = signal<Row[]>(ROWS);
    const t = new TableSort<Row>(() => src(), { searchKeys: () => ['name', 'genre'] });
    expect(t.view().length).toBe(3);
    src.set([ROWS[0]]);
    expect(t.view().length).toBe(1);
  });
});
