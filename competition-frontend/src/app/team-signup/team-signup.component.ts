import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';

interface Competition {
  competition_id: number;
  competition_name: string;
  sport_id: number;
  start_date: Date;
}

interface Team {
  team_id: number;
  team_name: string;
}

@Component({
  selector: 'app-team-signup',
  standalone: true,
  templateUrl: './team-signup.component.html',
  styleUrls: ['./team-signup.component.scss'],
  imports: [CommonModule, RouterModule],
})

export class TeamSignupComponent implements OnInit {
  competitions: Competition[] = [];
  teams: Team[] = [];
  selectedCompetitionId: number | null = null;
  userId: number | null = null;
  showModal: boolean = false;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.userId = Number(sessionStorage.getItem('userId'));
    this.http.get<Competition[]>('/api/handlers/competitions').subscribe((data: any) => {
      this.competitions = data;
    });
  }

  openModal(competitionId: number) {
    this.selectedCompetitionId = competitionId;
    this.http.get<Team[]>('/api/handlers/teams', { params: { user_id: String(this.userId) } }).subscribe((data: any) => {
      this.teams = data;
      this.showModal = true; 
    });
  }

  signUp(teamId: number) {
    if (!this.selectedCompetitionId) {
      alert('No competition selected');
      return;
    }

    const payload = { competition_id: this.selectedCompetitionId, team_id: teamId };
    this.http.post('/handlers/team_signup', payload).subscribe({
      next: () => {
        alert('Team signed up successfully!');
        this.showModal = false;
      },
      error: (err: any) => {
        const errorMessage = err.error?.message || 'Signup failed';
        alert(errorMessage);
      },
    });
  }
}