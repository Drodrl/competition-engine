import { Routes } from '@angular/router';
import { LoginComponent } from './login/login.component';

export const appRoutes: Routes = [
  { path: '', redirectTo: 'login-page', pathMatch: 'full' }, // Redirect root to login
  { path: 'login-page', component: LoginComponent }//,
];
