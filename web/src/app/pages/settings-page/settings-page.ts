import { Component, effect, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { ApiService } from '../../services/api';

@Component({
  selector: 'app-settings-page',
  imports: [FormsModule],
  templateUrl: './settings-page.html',
  styleUrl: './settings-page.scss',
})
export class SettingsPage {
  private api = inject(ApiService);

  readonly config = this.api.config;
  tvIP = '';
  saving = false;
  readonly error = signal<string | null>(null);
  private dirty = false;

  constructor() {
    // Seed the input once the resource loads, without overwriting user edits.
    effect(() => {
      const cfg = this.config.value();
      if (cfg && !this.dirty) {
        this.tvIP = cfg.tv_ip;
      }
    });
  }

  markDirty() {
    this.dirty = true;
  }

  save() {
    this.saving = true;
    this.error.set(null);
    this.api
      .updateConfig({ tv_ip: this.tvIP })
      .pipe(finalize(() => (this.saving = false)))
      .subscribe({
        next: () => {
          this.dirty = false;
          this.config.reload?.();
        },
        error: (err) => {
          this.error.set(err?.error?.error ?? 'Failed to save settings');
        },
      });
  }
}
