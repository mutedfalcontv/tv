import { Injectable, signal } from '@angular/core';

export interface LogFrame {
  line: string;
  timestamp: string;
}

@Injectable({ providedIn: 'root' })
export class WebSocketService {
  private ws: WebSocket | null = null;
  readonly connected = signal(false);
  readonly logs = signal<LogFrame[]>([]);

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) return;
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${proto}//${window.location.host}/ws/logs`;
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.connected.set(true);
    };

    this.ws.onmessage = (event) => {
      try {
        const frame: LogFrame = JSON.parse(event.data);
        this.logs.update((prev) => {
          const next = [...prev, frame];
          return next.length > 500 ? next.slice(-500) : next;
        });
      } catch {
        // ignore parse errors
      }
    };

    this.ws.onclose = () => {
      this.connected.set(false);
      this.ws = null;
    };
  }

  disconnect() {
    this.ws?.close();
  }

  filter(opts: { level?: string; tag?: string; pid?: number }) {
    this.ws?.send(JSON.stringify({ type: 'filter', ...opts }));
  }
}
