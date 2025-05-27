import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';

interface Competition {
  competition_id: number;
  competition_name: string;
  sport_id: number;
  start_date: Date;
  end_date?: Date;
  status?: number;
}

interface Sport {
  id: number;
  name: string;
}

@Component({
  selector: 'app-competition-signup',
  standalone: true,
  templateUrl: './competition-signup.component.html',
  imports: [CommonModule, RouterModule],
})

export class CompetitionSignupComponent implements OnInit {
  competitions: Competition[] = [];
  sports: Sport[] = [];
  userId: number | null = null;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.userId = Number(sessionStorage.getItem('userId'));
    this.http.get<Competition[]>('/api/competitions').subscribe((data: any) => {
      this.competitions = data;
    });
    this.http.get<Sport[]>('/api/sports').subscribe((data: any) => {
      this.sports = data;
    });
  }

  getSportName(sportId: number): string {
    const sport = this.sports.find(s => s.id === sportId);
    return sport ? sport.name : sportId.toString();
  }

  getCompStatus(competition: Competition): string {
    const status = this.competitions.find(c => c.competition_id === competition.competition_id)?.status;
    if (status === 1) return "Open";
    if (status === 0) return "Closed";
    return "Unknown";
  }

  signUp(competitionId: number) {
    if (!this.userId) {
      alert('User not logged in');
      return;
    }


    const payload = { competition_id: competitionId, user_id: this.userId };
    this.http.post('/handlers/user_signup', payload).subscribe({
      next: () => alert('Successfully signed up!'),
      error: (err: any) => {
            const errorMessage = err.error?.message || 'Signup failed';
            alert(errorMessage); }
    });
  }
}