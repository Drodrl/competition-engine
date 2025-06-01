import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CompetitionService } from '../competition-create/competition.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-my-competitions',
  imports: [CommonModule],
  templateUrl: './my-competitions.component.html',
  providers: [CompetitionService],
  //styleUrl: './my-competitions.component.scss'
})
export class MyCompetitionsComponent implements OnInit {
  competitions: any[] = [];
  loading = true;
  error: string | null = null;

  constructor(private svc: CompetitionService, private router: Router) {}

  ngOnInit() {
    const organizerId = sessionStorage.getItem('userId');
    if (!organizerId){
      this.error = 'You must be logged in as an organizer';
      this.loading = false;
      return;
    }
    this.svc.getCompetitionsByOrganizer(organizerId).subscribe({
      next: data => {
        console.log('Competitions loaded:', data);
        this.competitions = [...data];
        this.loading = false;
      },
      error: err => {
        this.error = 'Failed to load competitions';
        this.loading = false;
      }
    });
  }

  editCompetition(id: number, status: number) {
    if (status === 0){
      this.router.navigate(['/edit-competition', id]);
    }
    else{
      this.router.navigate(['/manage-competition', id]);
    }
  }

  statusLabel(status: number): string {
    return this.svc.statusLabel(status)
  }

  getActionLabel(status: number): string {
    if (status === 0) return 'Edit';        // Draft
    if (status === 1 || status === 2) return 'Manage'; // Open or Closed
    if (status === 3) return 'Details';     // Finished
    return 'Manage';
  }

  goBack() {
    if (!confirm("All unsaved changes will be lost. Are you sure you want to go back?")) return;
    this.router.navigate(['/organizer-dashboard']);
  }

}
