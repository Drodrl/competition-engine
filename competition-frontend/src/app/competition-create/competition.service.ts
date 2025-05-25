import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Item { 
  id: number; 
  name: string; 
}

export interface TournamentFormat {
  id: number;
  name: string;
  minimum_participants: number;
}

export interface StageDTO {
  stage_id?: number;
  stage_name: string;
  stage_order: number;
  tourney_format_id: number;
  participants_at_start: number;
  participants_at_end: number;
}

@Injectable()
export class CompetitionService {
  constructor(private http: HttpClient) {}

  statusLabel(status: number): string {
    switch (status) {
      case 0: return 'Draft';
      case 1: return 'Open';
      case 2: return 'Ongoing';
      case 3: return 'Finished';
      case 4: return 'Cancelled';
      default: return 'Unknown';
    }
  }

  createCompetition(data: any): Observable<{competition_id: number}> {
    return this.http.post<{competition_id: number}>('/api/competitions/draft', data);
  }

  deleteCompetition(id: number): Observable<any> {
    return this.http.delete(`/api/competitions/${id}`);
  }

  getSports(): Observable<Item[]> {
    return this.http.get<Item[]>('/api/sports');
  }

  getTournamentFormats(): Observable<TournamentFormat[]> {
    return this.http.get<TournamentFormat[]>('/api/tourney-formats');
  }

  getCompetitionsByOrganizer(organizerId: string): Observable<any[]> {
    return this.http.get<any[]>(`/api/competitions/organizer/${organizerId}`);
  }

  getCompetitionById(id: number): Observable<any> {
    return this.http.get<any>(`/api/competitions/${id}`);
  }

  getStagesByCompetitionId(id: number): Observable<StageDTO[]> {
    return this.http.get<StageDTO[]>(`/api/competitions/${id}/stages`);
  }

  updateCompetition(id: number, data: any): Observable<any> {
    return this.http.put(`/api/competitions/${id}`, data);
  }

  addStage(competitionId: number, stage: StageDTO): Observable<any> {
    return this.http.post(`/api/competitions/${competitionId}/stages`, stage);
  }

  updateStage(competitionId: number, stageId: number, stage: StageDTO): Observable<any> {
    return this.http.put(`/api/competitions/${competitionId}/stages/${stageId}`, stage);
  }

  deleteStage(competitionId: number, stageId: number): Observable<any> {
    return this.http.delete(`/api/competitions/${competitionId}/stages/${stageId}`);
  }

  changeCompetitionStatus(id: number, status: number): Observable<any> {
    return this.http.patch(`/api/competitions/${id}/status`, { status });
  }
}
