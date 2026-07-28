import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { httpResource } from '@angular/common/http';

export interface Status {
  connected: boolean;
  tv_ip: string;
}

export interface AppInfo {
  PackageName: string;
  Name: string;
}

export interface TvConfig {
  tv_ip: string;
  default_player: string;
}

export interface LogResult {
  logs: string;
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  constructor(private http: HttpClient) {}

  readonly status = httpResource<Status>(() => '/api/status');
  readonly keys = httpResource<string[]>(() => '/api/status/keys');
  readonly apps = httpResource<string[]>(() => '/api/apps');
  readonly players = httpResource<AppInfo[]>(() => '/api/players');
  readonly config = httpResource<TvConfig>(() => '/api/config');

  getLogs(opts?: { level?: string; tag?: string; lines?: number; pkg?: string }) {
    const params = new URLSearchParams();
    if (opts?.level) params.set('level', opts.level);
    if (opts?.tag) params.set('tag', opts.tag);
    if (opts?.lines) params.set('lines', String(opts.lines));
    if (opts?.pkg) params.set('pkg', opts.pkg);
    const qs = params.toString();
    return this.http.get<LogResult>(`/api/logs${qs ? '?' + qs : ''}`);
  }

  press(key: string) {
    return this.http.post<{ ok: boolean }>('/api/remote/press', { key });
  }

  type(text: string) {
    return this.http.post<{ ok: boolean }>('/api/remote/type', { text });
  }

  launch(pkg: string) {
    return this.http.post<{ ok: boolean }>('/api/apps/launch', { package: pkg });
  }

  kill(pkg: string) {
    return this.http.post<{ ok: boolean }>('/api/apps/kill', { package: pkg });
  }

  setDefaultPlayer(pkg: string) {
    return this.http.put<{ ok: boolean }>('/api/player/default', { package: pkg });
  }

  play(url: string, player?: string) {
    return this.http.post<{ ok: boolean }>('/api/play', { url, player });
  }

  updateConfig(cfg: Partial<TvConfig>) {
    return this.http.put<{ ok: boolean }>('/api/config', cfg);
  }
}
