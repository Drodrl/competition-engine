import { Routes } from '@angular/router';
import { LoginComponent } from './login/login.component';
import { CompetitionCreateComponent } from './competition-create/competition-create.component';

export const appRoutes: Routes = [
  { path: 'login-page', component: LoginComponent },
  { path: 'create-competition', component: CompetitionCreateComponent }
];
