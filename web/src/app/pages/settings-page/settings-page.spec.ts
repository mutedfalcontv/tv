import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SettingsPage } from './settings-page';
import { ApiService } from '../../services/api';

describe('SettingsPage', () => {
  let component: SettingsPage;
  let fixture: ComponentFixture<SettingsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SettingsPage],
      providers: [
        {
          provide: ApiService,
          useValue: {
            config: { value: () => ({ tv_ip: '', default_player: '' }), reload: () => {} },
            updateConfig: () => ({ pipe: () => ({ subscribe: () => {} }) }),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SettingsPage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
