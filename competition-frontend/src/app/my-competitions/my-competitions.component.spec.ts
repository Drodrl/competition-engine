import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MyCompetitionsComponent } from './my-competitions.component';

describe('MyCompetitionsComponent', () => {
  let component: MyCompetitionsComponent;
  let fixture: ComponentFixture<MyCompetitionsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyCompetitionsComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MyCompetitionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
