import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormArray, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CompetitionService, Item } from './competition.service';
import { CommonModule } from '@angular/common';

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
  structures: Item[] = [];
  activities: Item[] = [];
  formats: Item[] = [];

  constructor(private fb: FormBuilder, private svc: CompetitionService) {
    this.form = this.fb.group({
      competition_name: ['', Validators.required],
      sport_id: [null, Validators.required],
      start_date: ['', Validators.required],
      end_date: ['', Validators.required],
      structure_type_id: [null, Validators.required],
      max_participants: [null, [Validators.required, Validators.min(1)]],
      stages: this.fb.array([])
    });
  }

  ngOnInit() {
    this.addStage(false);
    this.svc.getSports().subscribe(data => this.sports = data);
    this.svc.getStructureTypes().subscribe(data => this.structures = data);
    this.svc.getActivityTypes().subscribe(data => this.activities = data);
    this.svc.getTournamentFormats().subscribe(data => this.formats = data);

    // Listen for changes to structure_type_id
    this.form.get('structure_type_id')?.valueChanges.subscribe(value => {
      this.handleStructureTypeChange(value);
    });
  }

  get stages(): FormArray {
    return this.form.get('stages') as FormArray;
  }

  private buildStage(index: number) {
    const stage = this.fb.group({
      stage_name: ['', Validators.required],
      stage_order: [index + 1],
      tourney_format_id: [null, Validators.required],
      activity_type_id: [null, Validators.required],
      groups: [null],
      participants: [{ value: null, disabled: true }]
    });
  
    stage.get('tourney_format_id')?.valueChanges.subscribe(format => {
      if (format === 4) { // Groups format
        stage.get('groups')?.setValidators([Validators.required, Validators.min(1)]);
      } else {
        stage.get('groups')?.clearValidators();
      }
      stage.get('groups')?.updateValueAndValidity();
    });
  
    return stage;
  }

  private handleStructureTypeChange(structureTypeID: number) {
    if (structureTypeID === 1) {
      while (this.stages.length > 1) {
        this.stages.removeAt(this.stages.length - 1);
      }
      if (this.stages.length === 0) {
        this.addStage();
      }
    }
  }

  addStage(calculate = true) {
    if (this.stages.length >= 3) return;
    this.stages.push(this.buildStage(this.stages.length));
    this.updateStageOrders();
    if (calculate) {
      //this.calculateParticipants();
    }
  }

  removeStage(i: number) {
    if (this.stages.length <= 1 || this.form.get('structure_type_id')?.value === 1) return;
    this.stages.removeAt(i);
    this.updateStageOrders();
    //this.calculateParticipants();
  }

  private updateStageOrders() {
    this.stages.controls.forEach((ctrl, i) =>
      ctrl.get('stage_order')!.setValue(i + 1, { emitEvent: false }));
  }

  submit() {
    this.form.markAllAsTouched();
    if (this.form.invalid) return;

    const formValue = this.form.value;

    const payload = {
      competition_name: formValue.competition_name,
      sport_id: +formValue.sport_id,
      start_date: formValue.start_date,
      end_date: formValue.end_date,
      organizer_id: 1,
      structure_type_id: +formValue.structure_type_id,
      stages: formValue.stages.map((stage: any, index: number) => ({
        stage_name: stage.stage_name,
        stage_order: index + 1,
        tourney_format_id: +stage.tourney_format_id,
        activity_type_id: +stage.activity_type_id,
        groups: stage.groups,
        participants: null // Set to null for now, will be calculated in the backend
      }))
    };

    console.log('Submitting payload:', payload);
    this.svc.createCompetition(payload).subscribe({
      next: res => alert('Created ID: ' + res.competition_id),
      error: err => alert('Error: ' + (err.error || err.message))
    });
  }
}
