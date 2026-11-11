import { Service, signal, DestroyRef, inject } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import { Subject, Subscription, scan, share, takeUntil, filter as rxFilter, timer } from 'rxjs';

export interface LogFrame {
  line: string;
  timestamp: string;
}

type Filter = { level?: string; tag?: string; pid?: number };

@Service()
export class WebSocketService {
  private destroyRef = inject(DestroyRef);
  private ws: WebSocketSubject<LogFrame | { type: string; [k: string]: unknown }> | null = null;
  private disconnect$ = new Subject<void>();
  private reconnectSub: Subscription | null = null;
  private attempt = 0;
  private manualClose = false;
  private pendingFilter: Filter | null = null;

  readonly connected = signal(false);
  readonly connecting = signal(false);
  readonly error = signal<string | null>(null);
  readonly logs = signal<LogFrame[]>([]);

  constructor() {
    this.destroyRef.onDestroy(() => this.disconnect());
  }

  connect() {
    this.manualClose = false;
    if (this.ws) return;
    this.clearReconnect();
    this.open();
  }

  private open() {
    if (this.manualClose) return;
    this.connecting.set(true);
    this.error.set(null);

    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${proto}//${window.location.host}/ws/logs`;

    this.ws = webSocket<LogFrame | { type: string; [k: string]: unknown }>({
      url,
      openObserver: {
        next: () => {
          this.connected.set(true);
          this.connecting.set(false);
          this.attempt = 0;
          this.error.set(null);
          if (this.pendingFilter) {
            this.ws?.next({ type: 'filter', ...this.pendingFilter });
          }
        },
      },
      closeObserver: {
        next: () => {
          this.connected.set(false);
          this.connecting.set(false);
          this.ws = null;
          this.scheduleReconnect();
        },
      },
    });

    this.ws
      .pipe(
        takeUntil(this.disconnect$),
        rxFilter((msg): msg is LogFrame => 'line' in msg),
        scan<LogFrame, LogFrame[]>((acc, frame) => {
          const next = [...acc, frame];
          return next.length > 500 ? next.slice(-500) : next;
        }, []),
        share(),
      )
      .subscribe({
        next: (frames) => this.logs.set(frames),
        error: (err) => {
          this.error.set(err instanceof Error && err.message ? err.message : 'log stream disconnected');
        },
      });
  }

  private scheduleReconnect() {
    if (this.manualClose || this.reconnectSub) return;
    const delay = Math.min(1000 * 2 ** this.attempt, 10000);
    this.attempt++;
    this.reconnectSub = timer(delay).subscribe(() => {
      this.reconnectSub = null;
      this.open();
    });
  }

  private clearReconnect() {
    this.reconnectSub?.unsubscribe();
    this.reconnectSub = null;
  }

  disconnect() {
    this.manualClose = true;
    this.clearReconnect();
    this.disconnect$.next();
    this.ws?.complete();
    this.ws = null;
    this.connected.set(false);
    this.connecting.set(false);
  }

  filter(opts: Filter) {
    this.pendingFilter = opts;
    if (this.ws) {
      this.ws.next({ type: 'filter', ...opts });
    }
  }
}
