import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RemotePage } from './remote-page';

describe('RemotePage', () => {
  let component: RemotePage;
  let fixture: ComponentFixture<RemotePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RemotePage],
    }).compileComponents();

    fixture = TestBed.createComponent(RemotePage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
