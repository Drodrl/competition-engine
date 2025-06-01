import { Component, OnInit } from '@angular/core';
import { CompetitionService } from '../competition-create/competition.service';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';


@Component({
  selector: 'app-landing-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './landing-page.component.html',
  styleUrls: ['./landing-page.component.scss'],
  providers: [CompetitionService]
})
export class LandingPageComponent implements OnInit {
  finishedCompetitions: any[] = [];
  ongoingCompetitions: any[] = [];
  error: string | null = null;

  constructor(private svc: CompetitionService, private router: Router) {}

  ngOnInit(): void {
    this.fetchCompetitions();
  }

  fetchCompetitions(): void {
    this.svc.getAllCompetitions().subscribe({
      next: (data: any[]) => {
        // log fetched competitions
        console.log('Fetched competitions:', data);
        // Filter finished competitions
        this.finishedCompetitions = data.filter(comp => comp.status === 3);
        console.log('Finished competitions:', this.finishedCompetitions);
        // Filter ongoing competitions 
        this.ongoingCompetitions = data.filter(comp => comp.status === 1 || comp.status === 2);
        console.log('Ongoing competitions:', this.ongoingCompetitions);
      },
      error: (err) => {
        this.error = 'Failed to fetch competitions';
        console.error(err);
      }
    });
  }

  goToLogin(): void {
    this.router.navigate(['/login-page']);
  }
}