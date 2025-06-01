import { Component } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';

interface Team {
  team_id: number;
  team_name: string;
}

interface User {
  id_user: number;
  name_user: string;
  lname1_user: string;
  remove?: boolean; // mark user to be removed
  add?: boolean; // mark user to be added
}

@Component({
  selector: 'app-team-manage',
  templateUrl: './team-manage.component.html',
  styleUrls: ['./team-manage.component.scss'],
  imports: [CommonModule, FormsModule, RouterModule],
  standalone: true
})
export class TeamManageComponent {
  teams: Team[] = []; 
  users: User[] = []; 
  athletes: User[] = []; 
  selectedTeam: any = null; 
  showEditModal = false; 
  showAddModal = false; 
  userId: number | null = null;

  constructor(private http: HttpClient, private router: Router) {}

  ngOnInit() {
    this.userId = Number(sessionStorage.getItem('userId'));
    if (!this.userId) {
      alert('User not logged in');
      return;
    }

    this.fetchTeams();
  }

  fetchTeams() {
    if (!this.userId) return;

    this.http.get(`/api/user-teams?user_id=${this.userId}`).subscribe({
      next: (data: any) => {
        this.teams = data;
      },
      error: (err: any) => {
        const errorMessage = err.error?.message || 'Failed to fetch teams';
        alert(errorMessage);
      }
    });
  }


  openEditModal(team: any) {
    this.selectedTeam = { ...team, participants: team.participants || [] };
    this.showEditModal = true;

    this.http.get<User[]>(`/api/team-participants?team_id=${team.team_id}`).subscribe({
        next: (data: User[]) => {
          this.users = data.map(user => ({
              ...user,
              remove: false 
          }));
        },
        error: (err: any) => {
          const errorMessage = err.error?.message || 'Failed to fetch participants';
          alert(errorMessage);
        }
    });
  }

  closeEditModal() {
    this.showEditModal = false;
    this.selectedTeam = null;
    this.users = [];
  }

  openAddModal() {
    this.showAddModal = true;

    this.http.get<User[]>('/api/handlers/athletes').subscribe({
        next: (data: User[]) => {
          this.athletes = data.map(athlete => ({
              ...athlete,
              add: false 
          }));
        },
        error: (err: any) => {
        const errorMessage = err.error?.message || 'Failed to fetch athletes';
        alert(errorMessage);
        }
    });

  }

  closeAddModal() {
    this.showAddModal = false;
    this.athletes = [];
  }

addParticipants() {
  const selectedAthletes = this.athletes.filter(athlete => athlete.add);

  this.users = [...this.users, ...selectedAthletes.map(athlete => ({
    id_user: athlete.id_user, 
    name_user: athlete.name_user,
    lname1_user: athlete.lname1_user,
    remove: false
  }))];


  this.closeAddModal();
}

  updateTeam() {
      if (!this.selectedTeam) return;

      const participants = this.selectedTeam.participants || [];

      const usersToRemove = this.users
        .filter(user => user.remove)
        .map(user => user.id_user);

      const usersToAdd = this.users
        .filter(user => !user.remove) 
        .filter(user => !participants.some((p: User) => p.id_user === user.id_user)) 
        .map(user => user.id_user) 
        .filter(userId => userId);

      if (usersToRemove.length > 0) {
          this.http.post('/api/remove-participants', { team_id: this.selectedTeam.team_id, user_ids: usersToRemove }).subscribe({
          next: () => {
              alert('Participants removed successfully');
          },
          error: (err: any) => {
              const errorMessage = err.error?.message || 'Failed to remove participants';
              alert(errorMessage);
          }
          });
      }

      if (usersToAdd.length > 0) {
          this.http.post('/api/add-participants', { team_id: this.selectedTeam.team_id, user_ids: usersToAdd }).subscribe({
          next: () => {
              alert('Participants added successfully');
          },
          error: (err: any) => {
              const errorMessage = err.error?.message || 'Failed to add participants';
              alert(errorMessage);
          }
          });
      }

      this.closeEditModal();
      this.fetchTeams();
  }

  leaveTeam(teamId: number) {

    if (confirm('Are you sure you want to leave this team?')) {
      const payload = {
        team_id: teamId,
        user_ids: [this.userId] 
      };

      this.http.post('/api/remove-participants', payload).subscribe({
        next: () => {
          this.teams = this.teams.filter(team => team.team_id !== teamId);
        },
        error: (err: any) => {
          const errorMessage = err.error?.message || 'Failed to leave the team';
          alert(errorMessage);
        }
      });
    }
  
  }

  editTeam(teamId: number) {
    const team = this.teams.find(t => t.team_id === teamId);

    if (!team) {
        alert('Team not found');
        return;
    }

    this.openEditModal(team);
  }

  goToAthleteDashboard() {
    this.router.navigate(['/athlete-dashboard']); // Replace with the actual route for the athlete dashboard
  }
}