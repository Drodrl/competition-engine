import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { TeamSignupComponent } from './team-signup.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

describe('TeamSignupComponent', () => {
  let component: TeamSignupComponent;
  let fixture: ComponentFixture<TeamSignupComponent>;
  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [TeamSignupComponent, HttpClientTestingModule]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TeamSignupComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    spyOn(sessionStorage, 'getItem').and.returnValue('42');
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    expect(component).toBeTruthy();
  });

  it('should fetch competitions on init', () => {
    const mockCompetitions = [
      { competition_id: 1, competition_name: 'Comp1', sport_id: 2, start_date: new Date() }
    ];
    const req = httpMock.expectOne('/api/handlers/competitions');
    expect(req.request.method).toBe('GET');
    req.flush(mockCompetitions);

    expect(component.competitions.length).toBe(1);
    expect(component.competitions[0].competition_name).toBe('Comp1');
  });

  it('should open modal and fetch teams', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.userId = 42;
    component.openModal(1);

    const req = httpMock.expectOne(r =>
      r.url === '/api/handlers/teams' && r.params.get('user_id') === '42'
    );
    expect(req.request.method).toBe('GET');
    req.flush([{ team_id: 10, team_name: 'Team A' }]);

    expect(component.teams.length).toBe(1);
    expect(component.showModal).toBeTrue();
  });

  it('should alert if no competition selected when signing up', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.selectedCompetitionId = null;
    spyOn(window, 'alert');
    component.signUp(10);
    expect(window.alert).toHaveBeenCalledWith('No competition selected');
  });

  it('should POST signup and alert on success', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.selectedCompetitionId = 1;
    spyOn(window, 'alert');
    component.signUp(10);

    const req = httpMock.expectOne('/handlers/team_signup');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ competition_id: 1, team_id: 10 });

    req.flush({});
    expect(window.alert).toHaveBeenCalledWith('Team signed up successfully!');
    expect(component.showModal).toBeFalse();
  });

  it('should alert on signup error with message', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.selectedCompetitionId = 1;
    spyOn(window, 'alert');
    component.signUp(10);

    const req = httpMock.expectOne('/handlers/team_signup');
    req.flush({ message: 'Signup failed' }, { status: 400, statusText: 'Bad Request' });

    expect(window.alert).toHaveBeenCalledWith('Signup failed');
  });

  it('should alert "Signup failed" if error message is missing', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.selectedCompetitionId = 1;
    spyOn(window, 'alert');
    component.signUp(10);

    const req = httpMock.expectOne('/handlers/team_signup');
    req.flush({}, { status: 400, statusText: 'Bad Request' });

    expect(window.alert).toHaveBeenCalledWith('Signup failed');
  });
});