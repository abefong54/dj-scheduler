import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DataTableComponent, TableColumn } from './data-table.component';
import { provideTranslateService } from '@ngx-translate/core';

const COLS: TableColumn[] = [
  { key: 'name', label: 'Name' },
  { key: 'genre', label: 'Genre' },
];
const DATA: Record<string, unknown>[] = [
  { id: '1', name: 'DJ Alpha', genre: 'House' },
  { id: '2', name: 'DJ Beta', genre: 'Techno' },
  { id: '3', name: 'MC Gamma', genre: 'House' },
];

describe('DataTableComponent', () => {
  let fixture: ComponentFixture<DataTableComponent>;
  let component: DataTableComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DataTableComponent],
      providers: [provideTranslateService()],
    }).compileComponents();
    fixture = TestBed.createComponent(DataTableComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('columns', COLS);
    fixture.componentRef.setInput('data', DATA);
    fixture.detectChanges();
  });

  it('shows all rows by default', () => {
    expect(component.filteredData().length).toBe(3);
  });

  it('filters by search query case-insensitively', () => {
    component.searchQuery.set('dj');
    expect(component.filteredData().length).toBe(2);
    expect(component.filteredData().map(r => r['name'])).toEqual(['DJ Alpha', 'DJ Beta']);
  });

  it('filters across all searchable columns', () => {
    component.searchQuery.set('house');
    expect(component.filteredData().length).toBe(2);
  });

  it('sorts ascending on first header click', () => {
    component.sort(COLS[0]);
    expect(component.filteredData()[0]['name']).toBe('DJ Alpha');
  });

  it('sorts descending on second click of same column', () => {
    component.sort(COLS[0]);
    component.sort(COLS[0]);
    expect(component.filteredData()[0]['name']).toBe('MC Gamma');
  });

  it('resets to ascending when a different column is clicked', () => {
    component.sort(COLS[0]);
    component.sort(COLS[0]); // now descending
    component.sort(COLS[1]); // new column — should reset to asc
    expect(component.sortDir()).toBe('asc');
  });

  it('excludes searchable:false columns from search', () => {
    const cols: TableColumn[] = [
      { key: 'name', label: 'Name', searchable: false },
      { key: 'genre', label: 'Genre' },
    ];
    fixture.componentRef.setInput('columns', cols);
    component.searchQuery.set('dj'); // only matches 'name', which is non-searchable
    expect(component.filteredData().length).toBe(0);
  });

  it('does not sort when sortable is false', () => {
    const col: TableColumn = { key: 'name', label: 'Name', sortable: false };
    component.sort(col);
    expect(component.sortKey()).toBe('');
  });
});
