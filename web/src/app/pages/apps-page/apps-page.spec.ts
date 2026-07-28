import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AppsPage } from './apps-page';
import { ApiService } from '../../services/api';

describe('AppsPage', () => {
  let component: AppsPage;
  let fixture: ComponentFixture<AppsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppsPage],
      providers: [
        {
          provide: ApiService,
          useValue: {
            apps: { value: () => [], reload: () => {} },
            launch: () => ({ subscribe: () => {} }),
            kill: () => ({ subscribe: () => {} }),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(AppsPage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
