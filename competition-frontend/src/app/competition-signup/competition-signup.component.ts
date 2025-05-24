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

@Component({
  selector: 'app-competition-signup',
  standalone: true,
  templateUrl: './competition-signup.component.html',
  imports: [CommonModule, RouterModule],
})

export class CompetitionSignupComponent implements OnInit {
  competitions: Competition[] = [];
  userId: number | null = null;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.userId = Number(sessionStorage.getItem('userId'));
    this.http.get<Competition[]>('/api/competitions').subscribe((data: any) => {
      this.competitions = data;
    });
  }

  signUp(competitionId: number) {
    if (!this.userId) {
      alert('User not logged in');
      return;
    }


    const payload = { competition_id: competitionId, user_id: this.userId };
    this.http.post('/user_signup', payload).subscribe({
      next: () => alert('Successfully signed up!'),
      error: (err: any) => {
            const errorMessage = err.error?.message || 'Signup failed';
            alert(errorMessage); }
    });
  }
}