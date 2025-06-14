import { Routes } from '@angular/router';
import { LoginComponent } from './login/login.component';
import { CompetitionCreateComponent } from './competition-create/competition-create.component';
import { OrganizerDashboardComponent } from './organizer-dashboard/organizer-dashboard.component';
import { MyCompetitionsComponent } from './my-competitions/my-competitions.component';
import { EditCompetitionComponent } from './edit-competition/edit-competition.component';
import { AthleteDashboardComponent} from './athlete-dashboard/athlete-dashboard.component';
import { CompetitionSignupComponent } from './competition-signup/competition-signup.component';
import { TeamCreateComponent } from './team-create/team-create.component';
import { TeamSignupComponent } from './team-signup/team-signup.component';  
import { ManageCompetitionComponent } from './manage-competition/manage-competition.component';
import { ManageStageComponent } from './manage-stage/manage-stage.component';
import { TeamManageComponent } from './team-manage/team-manage.component';
import { LandingPageComponent } from './landing-page/landing-page.component';

export const appRoutes: Routes = [
  { path: '', redirectTo: 'landing-page', pathMatch: 'full' }, // Redirect root to landing page
  { path: 'landing-page', component: LandingPageComponent }, 
  { path: 'login-page', component: LoginComponent },
  { path: 'organizer-dashboard', component: OrganizerDashboardComponent },
  { path: 'create-competition', component: CompetitionCreateComponent },
  { path: 'my-competitions', component: MyCompetitionsComponent},
  { path: 'edit-competition/:id', component: EditCompetitionComponent},
  { path: 'athlete-dashboard', component: AthleteDashboardComponent },
  { path: 'competition-signup', component: CompetitionSignupComponent },
  { path: 'team-create', component: TeamCreateComponent },
  { path: 'team-signup', component: TeamSignupComponent },
  { path: 'manage-competition/:id', component: ManageCompetitionComponent},
  { path: 'manage-competition/:id/stage/:stageId', component: ManageStageComponent },
  { path: 'team-manage', component: TeamManageComponent },
];
