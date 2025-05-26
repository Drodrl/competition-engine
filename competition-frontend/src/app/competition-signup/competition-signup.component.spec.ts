import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { CompetitionSignupComponent } from './competition-signup.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';

describe('CompetitionSignupComponent', () => {
  let component: CompetitionSignupComponent;
  let fixture: ComponentFixture<CompetitionSignupComponent>;
  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [CompetitionSignupComponent, HttpClientTestingModule]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CompetitionSignupComponent);
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

  it('should alert if user not logged in when signing up', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.userId = null;
    spyOn(window, 'alert');
    component.signUp(1);
    expect(window.alert).toHaveBeenCalledWith('User not logged in');
  });

  it('should POST signup and alert on success', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.userId = 42;
    spyOn(window, 'alert');
    component.signUp(1);

    const req = httpMock.expectOne('/handlers/user_signup');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ competition_id: 1, user_id: 42 });

    req.flush({});
    expect(window.alert).toHaveBeenCalledWith('Successfully signed up!');
  });

  it('should alert on signup error', () => {
    httpMock.expectOne('/api/handlers/competitions').flush([]);
    component.userId = 42;
    spyOn(window, 'alert');
    component.signUp(1);

    const req = httpMock.expectOne('/handlers/user_signup');
    req.flush({ message: 'Signup failed' }, { status: 400, statusText: 'Bad Request' });

    expect(window.alert).toHaveBeenCalledWith('Signup failed');
  });

  it('should alert "Signup failed" if error message is missing', () => {
  httpMock.expectOne('/api/handlers/competitions').flush([]);
  component.userId = 42;
  spyOn(window, 'alert');
  component.signUp(1);

  const req = httpMock.expectOne('/handlers/user_signup');
  req.flush({}, { status: 400, statusText: 'Bad Request' });

  expect(window.alert).toHaveBeenCalledWith('Signup failed');
});
});