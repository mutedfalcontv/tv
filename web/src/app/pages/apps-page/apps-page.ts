import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api';

@Component({
  selector: 'app-apps-page',
  imports: [FormsModule],
  templateUrl: './apps-page.html',
  styleUrl: './apps-page.scss',
})
export class AppsPage {
  private api = inject(ApiService);

  readonly apps = this.api.apps;
  filter = '';

  get filteredApps(): string[] {
    const all = this.apps.value() ?? [];
    if (!this.filter) return all;
    const q = this.filter.toLowerCase();
    return all.filter((p) => p.toLowerCase().includes(q));
  }

  launch(pkg: string) {
    this.api.launch(pkg).subscribe();
  }

  kill(pkg: string) {
    this.api.kill(pkg).subscribe();
  }

  refresh() {
    this.apps.reload?.();
  }
}
