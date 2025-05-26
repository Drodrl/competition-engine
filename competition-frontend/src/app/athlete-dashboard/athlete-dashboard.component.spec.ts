import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { AthleteDashboardComponent } from './athlete-dashboard.component';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';

describe('AthleteDashboardComponent', () => {
  let component: AthleteDashboardComponent;
  let fixture: ComponentFixture<AthleteDashboardComponent>;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [AthleteDashboardComponent],
      providers: [
        { provide: ActivatedRoute, useValue: { params: of({}) } }
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AthleteDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});