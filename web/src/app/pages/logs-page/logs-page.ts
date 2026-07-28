import { Component, OnInit, OnDestroy, Signal, WritableSignal } from '@angular/core';
import { WebSocketService, LogFrame } from '../../services/websocket';

@Component({
  selector: 'app-logs-page',
  templateUrl: './logs-page.html',
  styleUrl: './logs-page.scss',
})
export class LogsPage implements OnInit, OnDestroy {
  readonly logs: WritableSignal<LogFrame[]>;
  readonly connected: Signal<boolean>;

  constructor(protected ws: WebSocketService) {
    this.logs = this.ws.logs;
    this.connected = this.ws.connected;
  }

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
