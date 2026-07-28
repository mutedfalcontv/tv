import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LogsPage } from './logs-page';
import { WebSocketService } from '../../services/websocket';

describe('LogsPage', () => {
  let component: LogsPage;
  let fixture: ComponentFixture<LogsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LogsPage],
      providers: [
        {
          provide: WebSocketService,
          useValue: {
            logs: { value: () => [], set: () => {} },
            connected: { value: () => false },
            connect: () => {},
            disconnect: () => {},
          },
        },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LogsPage);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
