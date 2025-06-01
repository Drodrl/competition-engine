import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { TeamSignupComponent } from './team-signup.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';

describe('TeamSignupComponent', () => {
  let component: TeamSignupComponent;
  let fixture: ComponentFixture<TeamSignupComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TeamSignupComponent, HttpClientTestingModule]
    }).compileComponents();

    fixture = TestBed.createComponent(TeamSignupComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);

    // Mock sessionStorage
    spyOn(sessionStorage, 'getItem').and.returnValue('1');
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    fixture.detectChanges();
    httpMock.expectOne('/api/competitions').flush([]);
    httpMock.expectOne('/api/sports').flush([]);
    expect(component).toBeTruthy();
  });

  it('should load competitions and sports on init', fakeAsync(() => {
    fixture.detectChanges();

    // Mock competitions
    const req1 = httpMock.expectOne('/api/competitions/flag_teams/true');
    expect(req1.request.method).toBe('GET');
    // req1.flush([{ competition_id: 1, competition_name: 'Comp', sport_id: 2, start_date: new Date() }]);
    req1.flush([]);

    // Mock sports
    const req2 = httpMock.expectOne('/api/sports');
    expect(req2.request.method).toBe('GET');
    // req2.flush([{ id: 2, name: 'Soccer' }]);
    req2.flush([]);

    tick();
    expect(component.competitions.length).toBe(1);
    expect(component.sports.length).toBe(1);
  }));

  it('should open modal and load teams', fakeAsync(() => {
    fixture.detectChanges();

    // Flush initial requests
    httpMock.expectOne('/api/competitions').flush([]);
    httpMock.expectOne('/api/sports').flush([]);

    component.userId = 42;
    component.openModal(5);

    const req = httpMock.expectOne(r => r.url === '/api/handlers/teams' && r.params.get('user_id') === '42');
    expect(req.request.method).toBe('GET');
    req.flush([{ team_id: 1, team_name: 'Team A' }]);

    tick();
    expect(component.teams.length).toBe(1);
    expect(component.showModal).toBeTrue();
  }));

  it('should call signUp and close modal on success', fakeAsync(() => {
    fixture.detectChanges();
    httpMock.expectOne('/api/competitions').flush([]);
    httpMock.expectOne('/api/sports').flush([]);

    component.selectedCompetitionId = 10;
    spyOn(window, 'alert');
    component.showModal = true;

    component.signUp(7);

    const req = httpMock.expectOne('/handlers/team_signup');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ competition_id: 10, team_id: 7 });
    req.flush({ message: 'Signup successful' });

    tick();
    expect(window.alert).toHaveBeenCalledWith('Team signed up successfully!');
    expect(component.showModal).toBeFalse();
  }));

  it('should alert error message on signup error', fakeAsync(() => {
    fixture.detectChanges();
    httpMock.expectOne('/api/competitions').flush([]);
    httpMock.expectOne('/api/sports').flush([]);

    component.selectedCompetitionId = 10;
    spyOn(window, 'alert');
    component.showModal = true;

    component.signUp(7);

    const req = httpMock.expectOne('/handlers/team_signup');
    req.flush({ message: 'Signup failed: already signed up' }, { status: 400, statusText: 'Bad Request' });

    tick();
    expect(window.alert).toHaveBeenCalledWith('Signup failed: already signed up');
    expect(component.showModal).toBeTrue();
  }));

  it('should alert if no competition selected on signUp', () => {
    spyOn(window, 'alert');
    component.selectedCompetitionId = null;
    component.signUp(1);
    expect(window.alert).toHaveBeenCalledWith('No competition selected');
  });

  it('should get sport name by id', () => {
    component.sports = [{ id: 3, name: 'Basketball' }];
    expect(component.getSportName(3)).toBe('Basketball');
    expect(component.getSportName(99)).toBe('99');
  });

  it('should get competition status', () => {
    component.competitions = [{ competition_id: 1, competition_name: '', sport_id: 0, start_date: new Date(), status: 1 }];
    expect(component.getCompStatus(component.competitions[0])).toBe('Open');
    component.competitions[0].status = 0;
    expect(component.getCompStatus(component.competitions[0])).toBe('Closed');
    component.competitions[0].status = 99 as any;
    expect(component.getCompStatus(component.competitions[0])).toBe('Unknown');
  });
});