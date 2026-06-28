import { Directive, TemplateRef, input } from '@angular/core';

@Directive({ selector: '[appColumnDef]', standalone: true })
export class ColumnDefDirective {
  columnKey = input.required<string>({ alias: 'appColumnDef' });
  constructor(public template: TemplateRef<{ row: Record<string, unknown> }>) {}
}
