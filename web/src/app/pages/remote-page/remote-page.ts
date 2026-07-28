import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpResourceRef } from '@angular/common/http';
import { finalize } from 'rxjs';
import { ApiService, Status } from '../../services/api';

@Component({
  selector: 'app-remote-page',
  imports: [FormsModule],
  templateUrl: './remote-page.html',
  styleUrl: './remote-page.scss',
})
export class RemotePage {
  readonly status: HttpResourceRef<Status | undefined>;
  readonly keys: HttpResourceRef<string[] | undefined>;
  text = '';
  sending = false;

  constructor(private api: ApiService) {
    this.status = this.api.status;
    this.keys = this.api.keys;
  }

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
