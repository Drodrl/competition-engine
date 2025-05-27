import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CompetitionService, StageDTO, TournamentFormat } from '../competition-create/competition.service';
import { FormBuilder, FormArray, FormGroup, Validators, ReactiveFormsModule, FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';


@Component({
  selector: 'app-edit-competition',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormsModule],
  providers: [CompetitionService],
  templateUrl: './edit-competition.component.html',
  //styleUrl: './edit-competition.component.scss'
})
export class EditCompetitionComponent implements OnInit {
  form: FormGroup;
  competitionId!: number;
  loading = true;
  error: string | null = null;
  status: number = 0;
  sports: any[] = [];
  stages: StageDTO[] = [];
  tourneyFormats: TournamentFormat[] = [];
  loadingStages = false;
  addStageDialogOpen = false;
  newStage: StageDTO = {
    stage_name: '',
    stage_order: this.stages.length +1,
    tourney_format_id: 0,
    participants_at_start: 0,
    participants_at_end: 1
  };
  editStageDialogOpen = false;
  editStageData: StageDTO | null = null;

  constructor(
    private route: ActivatedRoute,
    private svc: CompetitionService,
    private fb: FormBuilder,
    private router: Router
  ) {
    this.form = this.fb.group({
      competition_name: ['', Validators.required],
      sport_id: [{ value: null, disabled: true }, Validators.required],
      start_date: ['', Validators.required],
      end_date: ['', Validators.required],
      max_participants: [null, [Validators.required, Validators.min(2)]],
      flag_teams: [false, Validators.required]
    });
  }

  ngOnInit() {
    this.competitionId = Number(this.route.snapshot.paramMap.get('id'));
    this.svc.getSports().subscribe(data => this.sports = data);
    this.svc.getTournamentFormats().subscribe(
      data => {
      this.tourneyFormats = data,
      console.log('Tournaments loaded:', data);
    });

    this.svc.getCompetitionById(this.competitionId).subscribe({
      next: data => {
        this.form.patchValue({
          competition_name: data.competition_name,
          sport_id: data.sport_id,
          start_date: data.start_date ? data.start_date.substring(0, 10) : '',
          end_date: data.end_date ? data.end_date.substring(0, 10) : '',
          max_participants: data.max_participants,
          flag_teams: data.flag_teams
        });
        this.status = data.status;
        this.loading = false;
      },
      error: err => {
        this.error = 'Failed to load competition';
        this.loading = false;
      }
    });

    this.loadStages();
  }

  save() {
    if (this.form.invalid) return;
    const payload = {
      competition_name: this.form.value.competition_name,
      start_date: this.form.value.start_date,
      end_date: this.form.value.end_date,
      max_participants: this.form.value.max_participants,
      flag_teams: this.form.value.flag_teams
    };

    const prevMaxParticipants = this.stages.length > 0 ? this.stages[0].participants_at_start : null;
    const newMaxParticipants = this.form.value.max_participants;

    if (!this.checkCompetition()) return;
    if (!this.checkStages(newMaxParticipants)) return;
    
    if (prevMaxParticipants !== null && prevMaxParticipants !== newMaxParticipants) {
      // Update first stage
      const firstStage = { ...this.stages[0], participants_at_start: newMaxParticipants };
      this.svc.updateStage(this.competitionId, firstStage.stage_id!, firstStage).subscribe({
          next: () => {
            this.loadStages();
            this.editStageDialogOpen = false;
          },
          error: err => {
            alert('Error updating previous stage: ' + (err.error || err.message));
            this.loadStages();
            this.editStageDialogOpen = false;
          }
      });
  
      this.svc.updateCompetition(this.competitionId, payload).subscribe({
        next: () => {
          alert('Competition and stages updated successfully!');
          this.loadStages();
        },
        error: err => alert('Error updating competition: ' + (err.error || err.message))
      });
     
    } else {
      // No change in max_participants, just update competition
      this.svc.updateCompetition(this.competitionId, payload).subscribe({
        next: () => alert('Competition updated successfully!'),
        error: err => alert('Error updating competition: ' + (err.error || err.message))
      });
    }
  }

  checkCompetition(): boolean {
    if (this.form.value.end_date < this.form.value.start_date) {
      alert('End date cannot be before start date');
      return false;
    } else if (this.form.value.start_date < new Date().toISOString().split('T')[0]) {
      alert('Start date cannot be in the past');
      return false;
    } else if (this.form.value.max_participants < 2 || this.form.value.max_participants > 100 || this.form.value.max_participants % 2 !== 0) {
      alert('Max participants must be an even number between 2 and 100');
      return false;
    } else if (this.stages.length === 0) {
      alert('Please add at least one stage before saving.');
      return false;
    }
    return true;
  }

  checkStages(max_participants: number): boolean {
    let prevParticipants = max_participants;
    for (let i = 0; i < this.stages.length; i++) {
      const stage = this.stages[i];
      if (!stage.stage_name || !stage.tourney_format_id || !stage.participants_at_start) {
        alert('Please fill all required fields for each stage.');
        return false;
      }
      const format = this.tourneyFormats.find(tf => tf.id === stage.tourney_format_id);
      if (!format) {
        alert('Please select a valid tournament format for each stage.');
        return false;
      }
      if (stage.participants_at_start < format.minimum_participants) {
        alert(`Stage "${stage.stage_name}" requires at least ${format.minimum_participants} participants.`);
        return false;
      }
      if (i !== 0) {
        if (stage.participants_at_start > prevParticipants - 2) {
          alert(`Stage "${stage.stage_name}" cannot have more participants at start than previous stage's end minus 2 (${prevParticipants - 2}).`);
          return false;
        }
      }
      prevParticipants = stage.participants_at_start;
    }
    return true;
  }

  loadStages() {
    this.loadingStages = true;
    this.svc.getStagesByCompetitionId(this.competitionId).subscribe({
      next: data => {
        this.stages = Array.isArray(data) ? data : [];
        this.loadingStages = false;
      },
      error: err => {
        this.error = 'Failed to load stages';
        this.stages = []; // fallback to empty array on error
        this.loadingStages = false;
      }
    });
  }

  openCompetition() {
    this.save();

    this.svc.changeCompetitionStatus(this.competitionId, 1).subscribe({
      next: () => {
        alert('Competition opened successfully!');
        this.status = 1;
        this.router.navigate(['/my-competitions']);
      },
      error: err => alert('Error opening competition: ' + (err.error || err.message))
    });
  }

  addStageValid(): boolean {
    if (!Array.isArray(this.stages) || this.stages.length >= 3) {
      return false;
    }
    return true;
  }

  openAddStageDialog() {
    this.addStageDialogOpen = true;
    let participantCount = 0;
    if (this.stages.length === 0) {
      participantCount = this.form.value.max_participants;
    } else {
      const lastStage = this.stages[this.stages.length - 1];
      participantCount = lastStage.participants_at_end || 1;
    }
    
    this.newStage = {
      stage_name: '',
      stage_order: this.stages.length + 1,
      tourney_format_id: 0,
      participants_at_start: participantCount,
      participants_at_end: 1
    };
  }

  checkStageData(stage: StageDTO): boolean {
    const tourneyFormatId = +stage.tourney_format_id;
    const format = this.tourneyFormats.find(tf => tf.id === tourneyFormatId);
    const participantsStartOfLastStage = this.stages.length > 0 ? this.stages[this.stages.length - 1].participants_at_start : this.form.value.max_participants;
    if (!format) {
      alert('Please select a valid tournament format.');
      return false;
    }
    const minimumParticipants = format.minimum_participants;
    if ((stage.participants_at_start || 0) < minimumParticipants) {
      alert(`This format requires at least ${minimumParticipants} participants.`);
      return false;
    } else if (stage.participants_at_start % 2 !== 0) {
      alert('Participants at start must be an even number.');
      return false;
    }
    if (stage.participants_at_start > participantsStartOfLastStage - 2 && stage.stage_order > 1) {
      alert('Participants at start cannot exceed participants at start of previous stage (at least 2 less).');
      return false;
    }
    return true;
  }

  addStageConfirmed() {
    if (!this.addStageValid()) return;

    // Coerce types (important!)
    this.newStage.tourney_format_id = +this.newStage.tourney_format_id;
    this.newStage.stage_order = +this.newStage.stage_order;
    this.newStage.participants_at_start = +this.newStage.participants_at_start;
    this.newStage.participants_at_end = +this.newStage.participants_at_end;

    if (!this.checkStageData(this.newStage)) return;
    this.svc.addStage(this.competitionId, this.newStage).subscribe({
      next: () => {
        if (this.stages.length > 0) {
          const prevStage = this.stages[this.stages.length - 1];
          const updatedPrevStage = { ...prevStage, participants_at_end: this.newStage.participants_at_start };
          this.svc.updateStage(this.competitionId, prevStage.stage_id!, updatedPrevStage).subscribe({
            next: () => {
              this.loadStages();
              this.addStageDialogOpen = false;
            },
            error: err => {
              alert('Error updating previous stage: ' + (err.error || err.message));
              this.loadStages();
              this.addStageDialogOpen = false;
            }
          });
        } else {
          this.loadStages();
          this.addStageDialogOpen = false;
        }
      },
      error: err => alert('Error adding stage: ' + (err.error || err.message))
    });
  }

  openEditStageDialog(stage: StageDTO) {
    this.editStageDialogOpen = true;
    this.editStageData = { ...stage };
  }

  saveEditStage() {
    if (!this.editStageData) return;

    // Coerce types (important!)
    this.editStageData.tourney_format_id = +this.editStageData.tourney_format_id;
    this.editStageData.stage_order = +this.editStageData.stage_order;
    this.editStageData.participants_at_start = +this.editStageData.participants_at_start;
    this.editStageData.participants_at_end = +this.editStageData.participants_at_end;

    if (!this.checkStageData(this.editStageData)) return;
    const stageIndex = this.stages.findIndex(s => s.stage_id === this.editStageData!.stage_id);

    this.svc.updateStage(this.competitionId, this.editStageData.stage_id!, this.editStageData).subscribe({
      next: () => {
        if (stageIndex > 0) {
          const prevStage = this.stages[stageIndex - 1];
          const updatedPrevStage = { ...prevStage, participants_at_end: this.editStageData!.participants_at_start };
          this.svc.updateStage(this.competitionId, prevStage.stage_id!, updatedPrevStage).subscribe({
            next: () => {
              this.loadStages();
              this.editStageDialogOpen = false;
            },
            error: err => {
              alert('Error updating previous stage: ' + (err.error || err.message));
              this.loadStages();
              this.editStageDialogOpen = false;
            }
          });
        } else {
          this.loadStages();
          this.editStageDialogOpen = false;
        }
      },
      error: err => alert('Error editing stage: ' + (err.error || err.message))
    });
  }

  deleteStage(stage: StageDTO) {
    if (!confirm('Are you sure you want to delete this stage?')) return;
    const stageIndex = this.stages.findIndex(s => s.stage_id === stage.stage_id);
    this.svc.deleteStage(this.competitionId, stage.stage_id!).subscribe({
      next: () => {
        // If there is a previous and next stage, update previous stage's participants_at_end
        if (stageIndex > 0 && stageIndex < this.stages.length - 1) {
          const prevStage = this.stages[stageIndex - 1];
          const nextStage = this.stages[stageIndex + 1];
          const updatedPrevStage = { ...prevStage, participants_at_end: nextStage.participants_at_start };
          this.svc.updateStage(this.competitionId, prevStage.stage_id!, updatedPrevStage).subscribe({
            next: () => this.loadStages(),
            error: err => alert('Error updating previous stage: ' + (err.error || err.message))
          });
          const updatedNextStage = { ...nextStage, stage_order: prevStage.stage_order + 1 };
          this.svc.updateStage(this.competitionId, nextStage.stage_id!, updatedNextStage).subscribe({
            next: () => this.loadStages(),
            error: err => alert('Error updating next stage: ' + (err.error || err.message))
          });
        }
        else if (stageIndex > 0 && stageIndex === this.stages.length) {
          const prevStage = this.stages[stageIndex - 1];
          const updatedPrevStage = { ...prevStage, participants_at_end: 1 };
          this.svc.updateStage(this.competitionId, prevStage.stage_id!, updatedPrevStage).subscribe({
            next: () => this.loadStages(),
            error: err => alert('Error updating previous stage: ' + (err.error || err.message))
          });
        }
        this.loadStages();
      },
      error: err => alert('Error deleting stage: ' + (err.error || err.message))
    });
  }

  get statusLabel(): string {
    return this.svc.statusLabel(this.status) 
  }
      

  tourneyLabel(tourneyId: number): string {
    const found = this.tourneyFormats.find(tf => tf.id === tourneyId);
    return found ? found.name : 'Unknown';
  }

  deleteCompetition() {
    if (!confirm("Are you sure you want to delete this competition?")) return;
    this.svc.deleteCompetition(this.competitionId).subscribe({
      next: () => {
        alert('Competition deleted successfully!');
        this.router.navigate(['/my-competitions']);
      },
      error: err => alert('Error deleting competition: ' + (err.error || err.message))
    });
  }

  goBack() {
    if (!confirm("All unsaved changes will be lost. Are you sure you want to go back?")) return;
    this.router.navigate(['/my-competitions']);
  }

  isFirstStage(): boolean {
    return this.stages.length === 0;
  }

}
