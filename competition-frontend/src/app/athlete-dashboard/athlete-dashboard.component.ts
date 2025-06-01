import { Component, OnInit} from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AthleteDashboardService, AthleteStats, Competition } from './athlete-dashboard.service';
import { HttpClientModule } from '@angular/common/http';

@Component({
  selector: 'app-athlete-dashboard',
  standalone: true,
  templateUrl: './athlete-dashboard.component.html',
  styleUrls: ['./athlete-dashboard.component.scss'],
  imports: [RouterModule, CommonModule, HttpClientModule],
  providers: [AthleteDashboardService]
})
export class AthleteDashboardComponent implements OnInit {
  userId: number | null = null;
  stats: AthleteStats | null = null;
  competitions: Competition[] = [];
  loading = true;
  error = '';

  constructor(private service: AthleteDashboardService) {}

  ngOnInit(): void {
    const userIdStr = sessionStorage.getItem('userId');
    if (!userIdStr) {
      this.error = 'User not logged in';
      this.loading = false;
      return;
    }
    
    this.userId = parseInt(userIdStr, 10);
    this.loadStats();
    this.loadCompetitions();
  }

  private loadStats(): void {
    if (!this.userId) return;
    
    this.service.getAthleteStats(this.userId).subscribe({
      next: (data) => {
        this.stats = data;
        this.loading = false;
      },
      error: (err) => {
        console.error('Failed to load athlete stats:', err);
        this.error = 'Failed to load statistics';
        this.loading = false;
      }
    });
  }

  private loadCompetitions(): void {
    if (!this.userId) return;
    
    this.service.getAthleteCompetitions(this.userId).subscribe({
      next: (data) => {
        this.competitions = data;
      },
      error: (err) => {
        console.error('Failed to load athlete competitions:', err);
        this.error = 'Failed to load competitions';
      }
    });
  }
  
  getStatusLabel(status: number): string {
    switch (status) {
      case 0: return 'Draft';
      case 1: return 'Open';
      case 2: return 'Ongoing';
      case 3: return 'Completed';
      default: return 'Draft';
    }
  }
}