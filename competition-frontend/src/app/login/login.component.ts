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
    this.http.post('/login', formValues)
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
