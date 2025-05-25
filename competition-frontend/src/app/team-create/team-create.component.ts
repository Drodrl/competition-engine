import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

interface User {
  id: number;
  first_name: string;
  last_name: string;
}

@Component({
  selector: 'app-team-create',
  standalone: true,
  templateUrl: './team-create.component.html',
//   styleUrls: ['./team-create.component.scss']
  imports: [CommonModule, RouterModule, ReactiveFormsModule],
})

export class TeamCreateComponent implements OnInit {
  teamForm: FormGroup;
  users: User[] = [];
  selectedUserIds: number[] = [];

  constructor(private fb: FormBuilder, private http: HttpClient) {
    this.teamForm = this.fb.group({
      teamName: ['']
    });
  }

  ngOnInit() {
    this.http.get<User[]>('/api/handlers/athletes').subscribe((data: any) => {
      this.users = data;
    });
  }

  onUserSelect(event: Event, userId: number) {
    const checkbox = event.target as HTMLInputElement;
    if (checkbox.checked) {
      this.selectedUserIds.push(userId);
    } else {
      this.selectedUserIds = this.selectedUserIds.filter((id) => id !== userId);
    }
  }

  onSubmit() {
    const payload = {
      teamName: this.teamForm.value.teamName,
      userIds: this.selectedUserIds
    };

    this.http.post('/handlers/team_create', payload).subscribe({
      next: () => alert('Team created successfully!'),
      error: (err: any) => {
            const errorMessage = err.error?.message || 'Create team failed';
            alert(errorMessage); }
    });
  }
}