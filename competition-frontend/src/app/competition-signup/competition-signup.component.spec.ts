import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { CompetitionSignupComponent } from './competition-signup.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

describe('CompetitionSignupComponent', () => {
  let component: CompetitionSignupComponent;
  let fixture: ComponentFixture<CompetitionSignupComponent>;
  let httpMock: HttpTestingController;

  const mockCompetitions = [
    { competition_id: 1, competition_name: 'Comp1', sport_id: 10, start_date: new Date(), status: 1 },
    { competition_id: 2, competition_name: 'Comp2', sport_id: 20, start_date: new Date(), status: 0 }
  ];
  const mockSports = [
    { id: 10, name: 'Football' },
    { id: 20, name: 'Basketball' }
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CompetitionSignupComponent, HttpClientTestingModule]
    }).compileComponents();

    fixture = TestBed.createComponent(CompetitionSignupComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    sessionStorage.setItem('userId', '123');
  });

  afterEach(() => {
    httpMock.verify();
    sessionStorage.clear();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // it('should fetch competitions and sports on init', fakeAsync(() => {
  //   fixture.detectChanges();
  //   const req1 = httpMock.expectOne('/api/competitions');
  //   req1.flush(mockCompetitions);
  //   const req2 = httpMock.expectOne('/api/sports');
  //   req2.flush(mockSports);
  //   tick();
  //   expect(component.competitions.length).toBe(2);
  //   expect(component.sports.length).toBe(2);
  // }));

  // it('getSportName should return sport name', () => {
  //   component.sports = mockSports;
  //   expect(component.getSportName(10)).toBe('Football');
  //   expect(component.getSportName(20)).toBe('Basketball');
  //   expect(component.getSportName(99)).toBe('99');
  //   expect(component.getSportName(null)).toBe('null');
  // });

//   it('should fetch competitions and sports on init', () => {
//   component.ngOnInit();

//   // Expect the request to fetch competitions
//   const competitionsRequest = httpMock.expectOne('/api/competitions/flag_teams/false');
//   expect(competitionsRequest.request.method).toBe('GET');
//   competitionsRequest.flush([]); // Respond with an empty array

//   // Expect the request to fetch sports
//   const sportsRequest = httpMock.expectOne('/api/sports');
//   expect(sportsRequest.request.method).toBe('GET');
//   sportsRequest.flush([]); // Respond with an empty array

//   expect(component.competitions).toEqual([]);
//   expect(component.sports).toEqual([]);

// });

  it('getCompStatus should return correct status', () => {
    component.competitions = mockCompetitions;
    expect(component.getCompStatus(mockCompetitions[0])).toBe('Open');
    expect(component.getCompStatus(mockCompetitions[1])).toBe('Closed');
    expect(component.getCompStatus(undefined)).toBe('Unknown');
  });

  it('signUp should alert if user not logged in', () => {
    spyOn(window, 'alert');
    component.userId = null;
    component.signUp(1);
    expect(window.alert).toHaveBeenCalledWith('User not logged in');
  });

  it('signUp should POST and alert on success', fakeAsync(() => {
    spyOn(window, 'alert');
    component.userId = 123;
    component.signUp(1);
    const req = httpMock.expectOne('/handlers/user_signup');
    expect(req.request.method).toBe('POST');
    req.flush({});
    tick();
    expect(window.alert).toHaveBeenCalledWith('Successfully signed up!');
  }));

  it('signUp should alert error message on failure', fakeAsync(() => {
    spyOn(window, 'alert');
    component.userId = 123;
    component.signUp(1);
    const req = httpMock.expectOne('/handlers/user_signup');
    req.flush({ message: 'Signup failed: already signed up' }, { status: 400, statusText: 'Bad Request' });
    tick();
    expect(window.alert).toHaveBeenCalledWith('Signup failed: already signed up');
  }));

  it('signUp should alert generic error if no message', fakeAsync(() => {
    spyOn(window, 'alert');
    component.userId = 123;
    component.signUp(1);
    const req = httpMock.expectOne('/handlers/user_signup');
    req.flush({}, { status: 400, statusText: 'Bad Request' });
    tick();
    expect(window.alert).toHaveBeenCalledWith('Signup failed');
  }));
});