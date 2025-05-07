import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Item { id: number; name: string; }

@Injectable()
export class CompetitionService {
  constructor(private http: HttpClient) {}
  createCompetition(data: any): Observable<{competition_id: number}> {
    return this.http.post<{competition_id: number}>('/api/competitions', data);
  }

  getSports(): Observable<Item[]> {
    return this.http.get<Item[]>('/api/sports');
  }
  getStructureTypes(): Observable<Item[]> {
    return this.http.get<Item[]>('/api/structure-types');
  }
  getActivityTypes(): Observable<Item[]> {
    return this.http.get<Item[]>('/api/activity-types');
  }
  getTournamentFormats(): Observable<Item[]> {
    return this.http.get<Item[]>('/api/tourney-formats');
  }
}
