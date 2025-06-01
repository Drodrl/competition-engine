import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { CompetitionService } from '../competition-create/competition.service';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-manage-stage',
  imports: [CommonModule, FormsModule],
  templateUrl: './manage-stage.component.html',
  styleUrls: ['./manage-stage.component.scss'],
  providers: [CompetitionService],
})
export class ManageStageComponent implements OnInit {
  competitionId!: number;
  stageId!: number;
  rounds: any[] = [];
  matches: any[] = [];
  selectedRound: any = null;
  matchParticipants: { [matchId: number]: any[] } = {};
  canGenerateNextRound = false;
  cannotGenerateReason = '';

  constructor(
    private route: ActivatedRoute,
    private svc: CompetitionService
  ) {}

  ngOnInit() {
    this.competitionId = +this.route.snapshot.paramMap.get('id')!;
    this.stageId       = +this.route.snapshot.paramMap.get('stageId')!;
    this.loadRounds();
    this.checkCanGenerateNextRound();
  }

  loadRounds() {
    this.svc.getRoundsByStageId(this.stageId)
      .subscribe(rounds => {
        this.rounds = rounds;
        this.checkCanGenerateNextRound();
      });
  }

  checkCanGenerateNextRound() {
    this.svc.getCanGenerateNextRound(this.stageId).subscribe(res => {
      this.canGenerateNextRound = res.canGenerate;
      this.cannotGenerateReason = res.reason || '';
    });
  }

  generateRound() {
    this.svc.generateNextRound(this.stageId)
      .subscribe({
        next: () => {
          alert('Next round generated.');
          this.loadRounds();
          this.checkCanGenerateNextRound();
        },
        error: err => {
          alert(err.error || err.message || 'Could not generate next round.');
          this.checkCanGenerateNextRound();
        }
      });
  }

  selectRound(round: any) {
    this.selectedRound = round;
    this.svc.getMatchesByRoundId(round.round_id)
      .subscribe(matches => {
        this.matches = matches;
        this.matchParticipants = {};
        matches.forEach(m => {
          this.svc.getMatchParticipants(m.match_id)
            .subscribe(parts => this.matchParticipants[m.match_id] = parts);
        });
      });
  }

  saveParticipantResult(matchId: number) {
    const payload = this.matchParticipants[matchId].map(p => ({
      participant_id: p.user_id ?? p.team_id,
      score:         p.score,
      is_winner:     p.is_winner
    }));
    this.svc.updateMatchResults(matchId, payload)
      .subscribe(() => alert('Results saved.'));
  }
}
