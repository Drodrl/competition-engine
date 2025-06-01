import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface AthleteStats {
  totalCompetitions: number;
  competitionsWon: number;
  ongoingCompetitions: number;
  upcomingCompetitions: number;
  winPercentage: number;
}

export interface Competition {
  competition_id: number;
  competition_name: string;
  sport_name: string;
  start_date: Date;
  end_date?: Date;
  status: number;
}

@Injectable({
  providedIn: 'root'
})
export class AthleteDashboardService {
  constructor(private http: HttpClient) {}

  getAthleteStats(userId: number): Observable<AthleteStats> {
    return this.http.get<AthleteStats>(`/api/athletes/${userId}/stats`);
  }

  getAthleteCompetitions(userId: number): Observable<Competition[]> {
    return this.http.get<Competition[]>(`/api/athletes/${userId}/competitions`);
  }
}