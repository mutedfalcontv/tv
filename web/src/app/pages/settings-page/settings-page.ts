import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { ApiService } from '../../services/api';

@Component({
  selector: 'app-settings-page',
  imports: [FormsModule],
  templateUrl: './settings-page.html',
  styleUrl: './settings-page.scss',
})
export class SettingsPage implements OnInit {
  private api = inject(ApiService);

  readonly config = this.api.config;
  tvIP = '';
  saving = false;

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
