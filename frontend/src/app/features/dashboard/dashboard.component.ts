import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';

import { NavbarComponent } from '@shared/components/navbar/navbar.component';
import { SidebarComponent } from '@shared/components/sidebar/sidebar.component';
import { AuthService } from '@core/services/auth.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    NavbarComponent,
    SidebarComponent
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  isSidebarOpen = true;
  isSidebarCollapsed = false;
  userEmail: string;

  constructor(private authService: AuthService) {
    this.userEmail = this.authService.getUserEmail();
  }

  ngOnInit(): void {
    this.checkScreenSize();
  }

  @HostListener('window:resize')
  onResize(): void {
    this.checkScreenSize();
  }

  private checkScreenSize(): void {
    const width = window.innerWidth;
    
    if (width < 768) {
      // Mobile: sidebar fecha completamente
      this.isSidebarOpen = false;
      this.isSidebarCollapsed = false;
    } else if (width < 1024) {
      // Tablet: sidebar recolhida (só ícones)
      this.isSidebarOpen = true;
      this.isSidebarCollapsed = true;
    } else {
      // Desktop: sidebar expandida
      this.isSidebarOpen = true;
      this.isSidebarCollapsed = false;
    }
  }

  toggleSidebar(): void {
    const width = window.innerWidth;
    
    if (width < 768) {
      // Mobile: abre/fecha completamente
      this.isSidebarOpen = !this.isSidebarOpen;
    } else {
      // Tablet/Desktop: expande/recolhe
      this.isSidebarCollapsed = !this.isSidebarCollapsed;
    }
  }

  closeSidebar(): void {
    if (window.innerWidth < 768) {
      this.isSidebarOpen = false;
    }
  }

  onLogout(): void {
    this.authService.logout();
  }
}
