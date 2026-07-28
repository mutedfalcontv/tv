import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RemotePage } from './remote-page';
import { ApiService } from '../../services/api';

describe('RemotePage', () => {
  let component: RemotePage;
  let fixture: ComponentFixture<RemotePage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RemotePage],
      providers: [
        {
          provide: ApiService,
          useValue: {
            status: { value: () => ({ connected: false, tv_ip: '' }) },
            keys: { value: () => [] },
            press: () => ({ subscribe: () => {} }),
            type: () => ({ subscribe: () => {} }),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(RemotePage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
