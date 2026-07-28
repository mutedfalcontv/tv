import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppsPage } from './apps-page';

describe('AppsPage', () => {
  let component: AppsPage;
  let fixture: ComponentFixture<AppsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppsPage],
    }).compileComponents();

    fixture = TestBed.createComponent(AppsPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
