import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { ApiService } from '../../services/api';

@Component({
  selector: 'app-player-page',
  imports: [FormsModule],
  templateUrl: './player-page.html',
  styleUrl: './player-page.scss',
})
export class PlayerPage {
  private api = inject(ApiService);

  readonly players = this.api.players;
  readonly config = this.api.config;
  url = '';
  selectedPlayer = '';
  sending = false;

  setDefault(pkg: string) {
    this.api.setDefaultPlayer(pkg).subscribe(() => this.config.reload?.());
  }

  play() {
    if (!this.url.trim()) return;
    this.sending = true;
    this.api.play(this.url, this.selectedPlayer || undefined).pipe(finalize(() => (this.sending = false))).subscribe();
  }
}
