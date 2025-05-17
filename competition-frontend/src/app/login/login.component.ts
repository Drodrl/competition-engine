<<<<<<< Updated upstream
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  //styleUrls: ['./login.component.css']
})
export class LoginComponent {
  constructor(private http: HttpClient, private router: Router) {}

  onLogin(formValues: any): void {
    console.log('Logging in:', formValues);
    this.http.post<{ userId: number; role: string }>('/login', formValues)
    .subscribe({
      next: (response) => {
        console.log('Login response:', response);
        if (response.role === 'organizer') {
          sessionStorage.setItem('userId', response.userId.toString()); // Save user ID in session
          this.router.navigate(['/organizer-dashboard']); // Redirect to organizer dashboard
        } else { //Add rest of the roles here
          alert('Only organizers can log in.');
        }
      },
      error: (error) => {
        console.error('Login error:', error);
        alert('Invalid email or password.');
      }
    });
  }
}
=======
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  //styleUrls: ['./login.component.css']
})
export class LoginComponent {
  constructor(private http: HttpClient) {}

  onLogin(formValues: any): void {
    console.log('Logging in:', formValues);
    this.http.post('/api/login', formValues)
      .subscribe({
        next: (response) => {
          console.log('Login response:', response);
        },
        error: (error) => {
          console.error('Login error:', error);
        }
      });
  }
}
>>>>>>> Stashed changes
