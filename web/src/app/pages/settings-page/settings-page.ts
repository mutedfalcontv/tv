import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpResourceRef } from '@angular/common/http';
import { finalize } from 'rxjs';
import { ApiService, TvConfig } from '../../services/api';

@Component({
  selector: 'app-settings-page',
  imports: [FormsModule],
  templateUrl: './settings-page.html',
  styleUrl: './settings-page.scss',
})
export class SettingsPage implements OnInit {
  readonly config: HttpResourceRef<TvConfig | undefined>;
  tvIP = '';
  saving = false;

  constructor(private api: ApiService) {
    this.config = this.api.config;
  }

  ngOnInit() {
    const cfg = this.config.value();
    if (cfg) this.tvIP = cfg.tv_ip;
  }

  save() {
    this.saving = true;
    this.api.updateConfig({ tv_ip: this.tvIP })
      .pipe(finalize(() => (this.saving = false)))
      .subscribe(() => this.config.reload?.());
  }
}
