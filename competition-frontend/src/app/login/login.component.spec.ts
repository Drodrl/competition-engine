import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoginComponent } from './login.component';
import { HttpTestingController, HttpClientTestingModule} from '@angular/common/http/testing';

describe('LoginComponent', () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginComponent, HttpClientTestingModule]
    })
    .compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController); 
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should call login API and log response on success', () => {
    const consoleSpy = spyOn(console, 'log');
    const formData = { username: 'test', password: '1234' };
    component.onLogin(formData);

    const req = httpMock.expectOne('/api/login');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(formData);

    const mockResponse = { token: 'fake-jwt' };
    req.flush(mockResponse);

    expect(consoleSpy).toHaveBeenCalledWith('Login response:', mockResponse);
  });

  it('should log error on login API failure', () => {
    const consoleErrorSpy = spyOn(console, 'error');
    const formData = { username: 'test', password: 'wrong' };
    component.onLogin(formData);

    const req = httpMock.expectOne('/api/login');
    req.flush({ message: 'Invalid credentials' }, { status: 401, statusText: 'Unauthorized' });

    expect(consoleErrorSpy).toHaveBeenCalled();
  });
});
