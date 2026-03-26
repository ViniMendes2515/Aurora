import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

interface MenuItem {
  label: string;
  icon: string;
  route?: string;
  action?: string;
}

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.css'
})
export class SidebarComponent {
  @Input() isOpen: boolean = true;
  @Input() isCollapsed: boolean = false;
  @Output() logout = new EventEmitter<void>();
  @Output() menuClick = new EventEmitter<void>();

  menuItems: MenuItem[] = [
    { label: 'Home', icon: 'pi-home', route: '/dashboard' },
    { label: 'Iluminação', icon: 'pi-sun', route: '/dashboard/devices' },
    { label: 'Sensores', icon: 'pi-wifi', route: '/dashboard/sensors' },
    { label: 'Regras', icon: 'pi-cog', route: '/dashboard/rules' },
    { label: 'Notificações', icon: 'pi-bell', route: '/dashboard/notifications' }, 
  ];

  onMenuClick(): void {
    this.menuClick.emit();
  }

  onLogout(): void {
    this.logout.emit();
  }
}
