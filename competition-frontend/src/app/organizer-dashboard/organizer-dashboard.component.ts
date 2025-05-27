import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-organizer-dashboard',
  imports: [],
  standalone: true,
  templateUrl: './organizer-dashboard.component.html',
  styleUrls: ['./organizer-dashboard.component.scss']
})
export class OrganizerDashboardComponent {
    constructor(private router: Router) {}

    navigateToCreateCompetition(): void {
      this.router.navigate(['/create-competition']);
    }

    navigateToMyCompetitions(): void {
      this.router.navigate(['/my-competitions']);
    }
}
