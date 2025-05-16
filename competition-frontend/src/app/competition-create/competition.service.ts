import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Item { id: number; name: string; }

@Injectable({ providedIn: 'root' })
export class CompetitionService {
  private api = '/api/competitions';

  constructor(private http: HttpClient) {}

  createCompetitionDraft(payload: any): Observable<any> {
    return this.http.post(`${this.api}/draft`, payload);
  }

  getCompetitionsByOrganizer(organizerId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.api}/organizer/${organizerId}`);
  }

  getCompetitionById(id: number): Observable<any> {
    return this.http.get<any>(`${this.api}/${id}`);
  }

  updateCompetition(id: number, payload: any): Observable<any> {
    return this.http.put(`${this.api}/${id}`, payload);
  }

  changeCompetitionStatus(id: number, status: number): Observable<any> {
    return this.http.patch(`${this.api}/${id}/status`, { status });
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
