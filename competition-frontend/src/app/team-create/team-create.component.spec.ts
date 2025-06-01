import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { TeamCreateComponent } from './team-create.component';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

describe('TeamCreateComponent', () => {
  let component: TeamCreateComponent;
  let fixture: ComponentFixture<TeamCreateComponent>;
  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [TeamCreateComponent, HttpClientTestingModule, ReactiveFormsModule]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TeamCreateComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    expect(component).toBeTruthy();
  });

  it('should fetch users on init', () => {
    const mockUsers = [
      { id: 1, first_name: 'A', last_name: 'B' }
    ];
    const req = httpMock.expectOne('/api/handlers/athletes');
    expect(req.request.method).toBe('GET');
    req.flush(mockUsers);

    expect(component.users.length).toBe(1);
    expect(component.users[0].first_name).toBe('A');
  });

  it('should add user id on checkbox select', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    const event = { target: { checked: true } } as any;
    component.onUserSelect(event, 5);
    expect(component.selectedUserIds).toContain(5);
  });

  it('should remove user id on checkbox deselect', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    component.selectedUserIds = [5];
    const event = { target: { checked: false } } as any;
    component.onUserSelect(event, 5);
    expect(component.selectedUserIds).not.toContain(5);
  });

  it('should POST team and alert on success', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    spyOn(window, 'alert');
    component.teamForm.setValue({ teamName: 'Equipo' });
    component.selectedUserIds = [1, 2];
    component.onSubmit();

    const req = httpMock.expectOne('/handlers/team_create');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ teamName: 'Equipo', userIds: [1, 2] });

    req.flush({});
    expect(window.alert).toHaveBeenCalledWith('Team created successfully!');
  });

  it('should alert error message from backend', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    spyOn(window, 'alert');
    component.teamForm.setValue({ teamName: 'Equipo' });
    component.selectedUserIds = [1];
    component.onSubmit();

    const req = httpMock.expectOne('/handlers/team_create');
    req.flush({ message: 'Ya existe' }, { status: 400, statusText: 'Bad Request' });

    expect(window.alert).toHaveBeenCalledWith('Ya existe');
  });

  it('should alert default error message if backend gives none', () => {
    httpMock.expectOne('/api/handlers/athletes').flush([]);
    spyOn(window, 'alert');
    component.teamForm.setValue({ teamName: 'Equipo' });
    component.selectedUserIds = [1];
    component.onSubmit();

    const req = httpMock.expectOne('/handlers/team_create');
    req.flush({}, { status: 400, statusText: 'Bad Request' });

    expect(window.alert).toHaveBeenCalledWith('Create team failed');
  });
});