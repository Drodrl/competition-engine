import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { CompetitionService, StageDTO } from '../competition-create/competition.service';

@Component({
  selector: 'app-manage-competition',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './manage-competition.component.html',
  styleUrl: './manage-competition.component.scss',
  providers: [CompetitionService],
})
export class ManageCompetitionComponent implements OnInit {
  competitionId!: number;
  competition: any;
  stages: StageDTO[] = [];
  participants: any[] = [];
  loading = true;
  error: string | null = null;
  signupClosed = false;
  canAdvanceStage = false;
  selectedStageId!: number;
  winner: any = null;

  // Cache for match completion status per stage
  private matchesCompleteCache: { [stageId: number]: boolean } = {};

  constructor(
    private route: ActivatedRoute,
    private svc: CompetitionService,
    private router: Router
  ) {}

  ngOnInit() {
    this.competitionId = Number(this.route.snapshot.paramMap.get('id'));
    this.svc.getCompetitionById(this.competitionId).subscribe({
      next: data => {
        this.competition = data;
        this.loading = false;
        if (data.winner) {
          this.winner = data.winner;
        }
      },
      error: err => {
        this.error = 'Failed to load competition';
        this.loading = false;
      }
    });

    this.svc.getStagesByCompetitionId(this.competitionId).subscribe({
      next: data => this.stages = data,
      error: () => this.stages = []
    });

    this.svc.getParticipantsByCompetitionId(this.competitionId).subscribe({
      next: data => this.participants = data,
      error: () => this.participants = []
    });
  }

  closeSignup() {
    this.svc.changeCompetitionStatus(this.competitionId, 2).subscribe({
      next: () => {
        this.signupClosed = true;
        alert('Signup period closed!');
      },
      error: err => alert('Error closing signup: ' + (err.error?.error || err.error || err.message))
    });
  }

  goToStage(stageId: any) {
    this.router.navigate(['/manage-competition', this.competitionId, 'stage', stageId]);
  }

  canAdvanceStageFor(stage: StageDTO): boolean {
    // Only allow for round robin stages and when all matches are complete
    if (stage.tourney_format_id !== 3 || stage.stage_id === undefined) return false;
    return this.allMatchesCompleteForStage(stage.stage_id);
  }

  allMatchesCompleteForStage(stageId: number): boolean {
    // If already checked, return cached value
    if (this.matchesCompleteCache[stageId] !== undefined) {
      return this.matchesCompleteCache[stageId];
    }

    // Default to false until checked
    this.matchesCompleteCache[stageId] = false;

    // Call backend to check if all matches are complete
    this.svc.getCanGenerateNextRound(stageId).subscribe({
      next: res => {
        // If canGenerate is true, all matches are complete for the current round
        this.matchesCompleteCache[stageId] = !!res.canGenerate;
      },
      error: () => {
        this.matchesCompleteCache[stageId] = false;
      }
    });

    // On first call, return false until async check completes
    return false;
  }

  advanceStage(stageId: number) {
    this.svc.advanceAfterRoundRobin(stageId).subscribe({
      next: res => {
        if (res.finished) {
          alert('Competition finished!');
        } else {
          alert('Advanced to next stage!');
        }
      },
      error: err => alert('Error: ' + (err.error || err.message))
    });
  }

  canFinishCompetition(): boolean {
    if (!this.stages || this.stages.length === 0) return false;
    const lastStage = this.stages[this.stages.length - 1];
    if (lastStage.stage_id === undefined) return false;
    return this.allMatchesCompleteForStage(lastStage.stage_id);
  }

  finishCompetition() {
    this.svc.finishCompetition(this.competitionId).subscribe({
      next: res => {
        alert('Competition finished!');
        if (res && res.winner) {
          this.winner = res.winner;
        }
      },
      error: err => alert('Error: ' + (err.error || err.message))
    });
  }

  goBackToMyCompetitions() {
    this.router.navigate(['/my-competitions']);
  }
}