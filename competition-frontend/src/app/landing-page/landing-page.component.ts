import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-landing-page',
  standalone: true,
  imports: [],
  templateUrl: './landing-page.component.html',
  styleUrls: ['./landing-page.component.scss']
})
export class LandingPageComponent implements OnInit {
  finishedCompetitions: any[] = [];
  ongoingCompetitions: any[] = [];
  error: string = '';

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.fetchFinishedCompetitions();
    this.fetchOngoingCompetitions();
  }

  fetchFinishedCompetitions(): void {
    this.http.get('/api/competitions').subscribe({
      next: (data: any) => {
        const competitions = data.filter((comp: any) => comp.status === 3);
        competitions.forEach((comp: any) => {
          this.http.get(`/api/competitions/${comp.competition_id}/participants`).subscribe({
            next: (participants: any) => {
              const winner = participants.find((p: any) => p.is_winner);
              this.finishedCompetitions.push({
                ...comp,
                winner: winner ? winner.name : 'N/A',
              });
            },
            error: (err) => {
              console.error(`Failed to fetch participants for competition ${comp.competition_id}`, err);
            }
          });
        });
      },
      error: (err) => {
        this.error = 'Failed to fetch finished competitions';
        console.error(err);
      }
    });
  }

  fetchOngoingCompetitions(): void {
    this.http.get('/api/competitions').subscribe({
      next: (data: any) => {
        this.ongoingCompetitions = data.filter((comp: any) => comp.status === 2);
      },
      error: (err) => {
        this.error = 'Failed to fetch ongoing competitions';
        console.error(err);
      }
    });
  }
}