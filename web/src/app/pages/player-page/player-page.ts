import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpResourceRef } from '@angular/common/http';
import { finalize } from 'rxjs';
import { ApiService, AppInfo, TvConfig } from '../../services/api';

@Component({
  selector: 'app-player-page',
  imports: [FormsModule],
  templateUrl: './player-page.html',
  styleUrl: './player-page.scss',
})
export class PlayerPage {
  readonly players: HttpResourceRef<AppInfo[] | undefined>;
  readonly config: HttpResourceRef<TvConfig | undefined>;
  url = '';
  selectedPlayer = '';
  sending = false;

  constructor(private api: ApiService) {
    this.players = this.api.players;
    this.config = this.api.config;
  }

  setDefault(pkg: string) {
    this.api.setDefaultPlayer(pkg).subscribe(() => this.config.reload?.());
  }

  play() {
    if (!this.url.trim()) return;
    this.sending = true;
    this.api.play(this.url, this.selectedPlayer || undefined).pipe(finalize(() => (this.sending = false))).subscribe();
  }
}
