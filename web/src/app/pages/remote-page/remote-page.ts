import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { ApiService } from '../../services/api';

@Component({
  selector: 'app-remote-page',
  imports: [FormsModule],
  templateUrl: './remote-page.html',
  styleUrl: './remote-page.scss',
})
export class RemotePage {
  private api = inject(ApiService);

  readonly status = this.api.status;
  readonly keys = this.api.keys;
  text = '';
  sending = false;

  press(key: string) {
    this.sending = true;
    this.api.press(key).pipe(finalize(() => (this.sending = false))).subscribe();
  }

  sendText() {
    if (!this.text.trim()) return;
    this.sending = true;
    this.api.type(this.text).pipe(finalize(() => (this.sending = false))).subscribe();
    this.text = '';
  }
}
