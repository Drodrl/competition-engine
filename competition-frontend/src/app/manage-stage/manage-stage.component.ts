import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { CompetitionService } from '../competition-create/competition.service';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

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
  roundMatchCounts: { [roundId: number]: { completed: number; total: number } } = {};
  competitionStatus: number = 0;
  generatingRound = false;


  constructor(
    private route: ActivatedRoute,
    private svc: CompetitionService,
    private router: Router
  ) {}

  ngOnInit() {
    this.competitionId = +this.route.snapshot.paramMap.get('id')!;
    this.stageId       = +this.route.snapshot.paramMap.get('stageId')!;
    this.loadRounds();
    this.checkCanGenerateNextRound();

    // Fetch competition status
    this.svc.getCompetitionById(this.competitionId).subscribe(data => {
      this.competitionStatus = data.status;
    });
  }

  loadRounds() {
    this.svc.getRoundsByStageId(this.stageId)
      .subscribe(rounds => {
        this.rounds = rounds;
        this.checkCanGenerateNextRound();
        this.loadMatchCountsForRounds();
      });
  }

  checkCanGenerateNextRound() {
    this.svc.getCanGenerateNextRound(this.stageId).subscribe(res => {
      this.canGenerateNextRound = res.canGenerate;
      this.cannotGenerateReason = res.reason || '';
    });
  }

  generateRound() {
    if (this.generatingRound) return;
    this.generatingRound = true;
    this.svc.generateNextRound(this.stageId)
      .subscribe({
        next: () => {
          alert('Next round generated.');
          this.loadRounds();
          this.checkCanGenerateNextRound();
          this.generatingRound = false;
        },
        error: err => {
          alert(err.error || err.message || 'Could not generate next round.');
          this.checkCanGenerateNextRound();
          this.generatingRound = false;
        }
      });
  }

  loadMatchCountsForRounds() {
    this.roundMatchCounts = {};
    this.rounds.forEach(round => {
      this.svc.getMatchesByRoundId(round.round_id).subscribe(matches => {
        const total = matches.length;
        const completed = matches.filter(m => m.completed_at).length;
        this.roundMatchCounts[round.round_id] = { completed, total };
      });
    });
  }

  setWinner(matchId: number, winnerIndex: number) {
    const participants = this.matchParticipants[matchId];
    if (participants) {
      participants.forEach((p, idx) => p.is_winner = idx === winnerIndex);
    }
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
    const participants = this.matchParticipants[matchId];
    if (!participants || participants.length !== 2) {
      alert('There must be exactly 2 participants per match.');
      return;
    }

    const [p1, p2] = participants;
    const s1 = p1.score;
    const s2 = p2.score;

    // Check: both scores empty or both filled
    const bothEmpty = (s1 === null || s1 === undefined || s1 === '') && (s2 === null || s2 === undefined || s2 === '');
    const bothFilled = (s1 !== null && s1 !== undefined && s1 !== '') && (s2 !== null && s2 !== undefined && s2 !== '');

    if (!bothEmpty && !bothFilled) {
      alert('Either both scores must be filled or both left empty.');
      return;
    }

    // If both filled, check winner's score >= loser's score
    if (bothFilled) {
      const winnerIdx = participants.findIndex(p => p.is_winner);
      const loserIdx = winnerIdx === 0 ? 1 : 0;
      const winnerScore = participants[winnerIdx].score;
      const loserScore = participants[loserIdx].score;
      if (winnerScore < loserScore) {
        alert('The winner\'s score cannot be lower than the loser\'s score.');
        return;
      }
    }

    const payload = participants.map(p => ({
      participant_id: p.user_id ?? p.team_id,
      score:         p.score,
      is_winner:     p.is_winner
    }));
    this.svc.updateMatchResults(matchId, payload)
      .subscribe(() => alert('Results saved.'));
  }

  goBackToManageCompetition() {
    this.router.navigate(['/manage-competition', this.competitionId]);
  }
}
