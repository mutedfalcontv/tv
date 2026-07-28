import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { WebSocketService } from '../../services/websocket';

@Component({
  selector: 'app-logs-page',
  templateUrl: './logs-page.html',
  styleUrl: './logs-page.scss',
})
export class LogsPage implements OnInit, OnDestroy {
  protected ws = inject(WebSocketService);

  readonly logs = this.ws.logs;
  readonly connected = this.ws.connected;

  ngOnInit() {
    this.ws.connect();
  }

  ngOnDestroy() {
    this.ws.disconnect();
  }

  clear() {
    this.logs.set([]);
  }
}
