import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormArray, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CompetitionService, Item } from './competition.service';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';

@Component({
  selector: 'app-competition-create',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  providers: [CompetitionService],
  templateUrl: './competition-create.component.html'
})
export class CompetitionCreateComponent implements OnInit {
  form: FormGroup;
  sports: Item[] = [];
  loading = true;
  error: string | null = null;
  organizerId: string | null = null;

  constructor(private fb: FormBuilder, private svc: CompetitionService, private router: Router) {
    this.form = this.fb.group({
      competition_name: ['', Validators.required],
      sport_id: [null, Validators.required],
      start_date: ['', Validators.required],
      end_date: ['', Validators.required],
      max_participants: [null, [Validators.required, Validators.min(2)]],
      flag_teams: [false, Validators.required]
    });
  }

  ngOnInit() {
    this.svc.getSports().subscribe(data => this.sports = data);
    this.organizerId = sessionStorage.getItem('userId');
    if (!this.organizerId){
      this.error = 'You must be logged in as an organizer';
      this.loading = false;
      return;
    }
  }

  checkCompetitionDraft(): boolean {
    if (this.form.value.end_date < this.form.value.start_date) {
      alert('End date cannot be before start date');
      return false;
    } else if (this.form.value.start_date < new Date().toISOString().split('T')[0]) {
      alert('Start date cannot be in the past');
      return false;
    } else if (this.form.value.max_participants < 2 || this.form.value.max_participants > 100 || this.form.value.max_participants % 2 !== 0) {
      alert('Max participants must be an even number between 2 and 100');
      return false;
    }
    return true;
  }

  submit() {
    this.form.markAllAsTouched();
    if (this.form.invalid) return;
    if (!this.checkCompetitionDraft()) return;

    const formValue = this.form.value;

    const payload = {
      competition_name: formValue.competition_name,
      sport_id: +formValue.sport_id,
      start_date: formValue.start_date,
      end_date: formValue.end_date,
      organizer_id: this.organizerId,
      max_participants: +formValue.max_participants,
      flag_teams: !!formValue.flag_teams
    };

    console.log('Submitting payload:', payload);
    this.svc.createCompetition(payload).subscribe({
      next: res => this.router.navigate(['/edit-competition', res.competition_id]),
      error: err => alert('Error: ' + (err.error || err.message))
    });
  }
}
