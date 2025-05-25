import { Routes } from '@angular/router';
import { LoginComponent } from './login/login.component';
import { AthleteDashboardComponent} from './athlete-dashboard/athlete-dashboard.component';
import { CompetitionSignupComponent } from './competition-signup/competition-signup.component';
import { TeamCreateComponent } from './team-create/team-create.component';
import { TeamSignupComponent } from './team-signup/team-signup.component';  

export const appRoutes: Routes = [
  { path: '', redirectTo: 'login-page', pathMatch: 'full' }, // Redirect root to login
  { path: 'login-page', component: LoginComponent },
  { path: 'athlete-dashboard', component: AthleteDashboardComponent },
  { path: 'competition-signup', component: CompetitionSignupComponent },
  { path: 'team-create', component: TeamCreateComponent },
  { path: 'team-signup', component: TeamSignupComponent } 
];
