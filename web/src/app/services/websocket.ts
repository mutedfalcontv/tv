import { Injectable, signal, DestroyRef, inject } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { Subject, scan, share, takeUntil, filter as rxFilter } from 'rxjs';

export interface LogFrame {
  line: string;
  timestamp: string;
}

@Injectable({ providedIn: 'root' })
export class WebSocketService {
  private destroyRef = inject(DestroyRef);
  private ws: WebSocketSubject<LogFrame | { type: string; [k: string]: unknown }> | null = null;
  private disconnect$ = new Subject<void>();

  readonly connected = signal(false);
  readonly logs = signal<LogFrame[]>([]);

  connect() {
    if (this.ws) return;
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${proto}//${window.location.host}/ws/logs`;

    this.ws = webSocket<LogFrame | { type: string; [k: string]: unknown }>({
      url,
      openObserver: { next: () => this.connected.set(true) },
      closeObserver: { next: () => { this.connected.set(false); this.ws = null; } },
    });

    this.ws.pipe(
      takeUntil(this.disconnect$),
      rxFilter((msg): msg is LogFrame => 'line' in msg),
      scan<LogFrame, LogFrame[]>((acc, frame) => {
        const next = [...acc, frame];
        return next.length > 500 ? next.slice(-500) : next;
      }, []),
      share(),
    ).subscribe((frames) => this.logs.set(frames));

    this.destroyRef.onDestroy(() => this.disconnect());
  }

  disconnect() {
    this.disconnect$.next();
    this.ws?.complete();
    this.ws = null;
    this.connected.set(false);
  }

  filter(opts: { level?: string; tag?: string; pid?: number }) {
    this.ws?.next({ type: 'filter', ...opts });
  }
}
