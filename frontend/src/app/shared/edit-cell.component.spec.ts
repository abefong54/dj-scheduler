import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EditCellComponent } from './edit-cell.component';

describe('EditCellComponent', () => {
  let fixture: ComponentFixture<EditCellComponent>;
  let cmp: EditCellComponent;

  function render(inputs: Record<string, unknown> = {}) {
    fixture = TestBed.createComponent(EditCellComponent);
    cmp = fixture.componentInstance;
    for (const [k, v] of Object.entries(inputs)) fixture.componentRef.setInput(k, v);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [EditCellComponent] }).compileComponents();
  });

  it('starts in display mode', () => {
    render({ value: 'The Warehouse' });
    expect(cmp.editing()).toBe(false);
  });

  it('enters edit mode and seeds the draft from value', () => {
    render({ value: 'The Warehouse' });
    cmp.startEdit();
    expect(cmp.editing()).toBe(true);
  });

  it('commits a changed value and emits it', () => {
    render({ value: 'Old Venue' });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.draft.set('New Venue');
    cmp.commitEdit();
    expect(emitted).toEqual([{ value: 'New Venue', endValue: '' }]);
    expect(cmp.editing()).toBe(false);
  });

  it('does not emit when the value is unchanged', () => {
    render({ value: 'Same' });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.commitEdit();
    expect(emitted).toEqual([]);
    expect(cmp.editing()).toBe(false);
  });

  it('reverts (no emit) on cancel', () => {
    render({ value: 'Original' });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.draft.set('Changed');
    cmp.cancelEdit();
    expect(emitted).toEqual([]);
    expect(cmp.editing()).toBe(false);
  });

  it('reverts instead of emitting when a required field is cleared', () => {
    render({ value: 'Name', required: true });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.draft.set('   ');
    cmp.commitEdit();
    expect(emitted).toEqual([]);
    expect(cmp.editing()).toBe(false);
  });

  it('date-range reverts when end is before start', () => {
    render({ type: 'date-range', value: '2026-07-04', endValue: '2026-07-06' });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.draftEnd.set('2026-07-01');
    cmp.commitEdit();
    expect(emitted).toEqual([]);
  });

  it('date-range commits both dates when valid', () => {
    render({ type: 'date-range', value: '2026-07-04', endValue: '2026-07-04' });
    const emitted: unknown[] = [];
    cmp.commit.subscribe(v => emitted.push(v));
    cmp.startEdit();
    cmp.draftEnd.set('2026-07-06');
    cmp.commitEdit();
    expect(emitted).toEqual([{ value: '2026-07-04', endValue: '2026-07-06' }]);
  });

  it('does not enter edit mode when disabled', () => {
    render({ value: 'x', disabled: true });
    cmp.startEdit();
    expect(cmp.editing()).toBe(false);
  });
});
