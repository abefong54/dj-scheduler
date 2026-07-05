import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideTranslateService } from '@ngx-translate/core';
import { LeadFormComponent } from './lead-form.component';

describe('LeadFormComponent (EL-085)', () => {
  let fixture: ComponentFixture<LeadFormComponent>;
  let http: HttpTestingController;
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LeadFormComponent],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideTranslateService()],
    }).compileComponents();
    fixture = TestBed.createComponent(LeadFormComponent);
    http = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
  });
  afterEach(() => http.verify());

  it('does not submit when email is empty', () => {
    const c = fixture.componentInstance;
    c.name.set('Aria');
    c.submit();
    http.expectNone('/api/leads');
    expect(c.state()).toBe('idle');
  });

  it('submits a valid lead and moves to success', () => {
    const c = fixture.componentInstance;
    c.name.set('Aria');
    c.email.set('aria@example.com');
    c.submit();
    http.expectOne('/api/leads').flush({ id: '1' });
    expect(c.state()).toBe('success');
  });

  it('moves to error when the request fails', () => {
    const c = fixture.componentInstance;
    c.name.set('Aria');
    c.email.set('aria@example.com');
    c.submit();
    http
      .expectOne('/api/leads')
      .flush({ error: 'bad' }, { status: 400, statusText: 'Bad Request' });
    expect(c.state()).toBe('error');
  });
});
