import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PlayerPage } from './player-page';
import { ApiService } from '../../services/api';

describe('PlayerPage', () => {
  let component: PlayerPage;
  let fixture: ComponentFixture<PlayerPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PlayerPage],
      providers: [
        {
          provide: ApiService,
          useValue: {
            players: { value: () => [], reload: () => {} },
            config: { value: () => ({ tv_ip: '', default_player: '' }), reload: () => {} },
            setDefaultPlayer: () => ({ subscribe: () => {} }),
            play: () => ({ subscribe: () => {} }),
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(PlayerPage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
