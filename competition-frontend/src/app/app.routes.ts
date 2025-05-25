import { Routes } from '@angular/router';
import { LoginComponent } from './login/login.component';
import { CompetitionCreateComponent } from './competition-create/competition-create.component';
import { OrganizerDashboardComponent } from './organizer-dashboard/organizer-dashboard.component';
import { MyCompetitionsComponent } from './my-competitions/my-competitions.component';
import { EditCompetitionComponent } from './edit-competition/edit-competition.component';

export const appRoutes: Routes = [
  { path: '', redirectTo: 'login-page', pathMatch: 'full' }, // Redirect root to login
  { path: 'login-page', component: LoginComponent },
  { path: 'organizer-dashboard', component: OrganizerDashboardComponent },
  { path: 'create-competition', component: CompetitionCreateComponent },
  { path: 'my-competitions', component: MyCompetitionsComponent},
  { path: 'edit-competition/:id', component: EditCompetitionComponent}
];
