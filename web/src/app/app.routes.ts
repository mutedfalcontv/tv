import { Routes } from '@angular/router';
import { RemotePage } from './pages/remote-page/remote-page';
import { AppsPage } from './pages/apps-page/apps-page';
import { LogsPage } from './pages/logs-page/logs-page';
import { PlayerPage } from './pages/player-page/player-page';
import { SettingsPage } from './pages/settings-page/settings-page';

export const routes: Routes = [
  { path: '', component: RemotePage },
  { path: 'apps', component: AppsPage },
  { path: 'logs', component: LogsPage },
  { path: 'player', component: PlayerPage },
  { path: 'settings', component: SettingsPage },
];
