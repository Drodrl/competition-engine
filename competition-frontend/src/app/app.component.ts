import { Component } from '@angular/core';
import { CommonModule } from '@angular/common'; 
import { LoginComponent } from './login/login.component';
import { RouterModule } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';
import { CompetitionCreateComponent } from './competition-create/competition-create.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, LoginComponent, RouterModule, HttpClientModule, CompetitionCreateComponent],
  templateUrl: './app.component.html',
  //styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'Competition Engine';
}
