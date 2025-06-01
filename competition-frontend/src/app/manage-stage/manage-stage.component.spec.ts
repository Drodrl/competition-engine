import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageStageComponent } from './manage-stage.component';

describe('ManageStageComponent', () => {
  let component: ManageStageComponent;
  let fixture: ComponentFixture<ManageStageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ManageStageComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ManageStageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
