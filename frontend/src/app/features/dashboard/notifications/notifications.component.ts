import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SecurityService, AlarmEvent } from '@core/services/security.service';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notifications.component.html',
})
export class NotificationsComponent implements OnInit {
  alarms: AlarmEvent[] = [];
  loading = false;

  constructor(private securityService: SecurityService) {}

  ngOnInit(): void {
    this.loadAlarms();
  }

  loadAlarms(): void {
    this.loading = true;
    this.securityService.getRecentAlarms().subscribe({
      next: (alarms) => {
        this.alarms = alarms || [];
        this.loading = false;
      },
      error: () => {
        this.loading = false;
      }
    });
  }

  silenceAlarm(alarm: AlarmEvent): void {
    this.securityService.silenceAlarm(alarm.id).subscribe({
      next: () => {
        alarm.status = 'silenced';
      }
    });
  }

  triggerTestAlarm(): void {
    this.securityService.triggerAlarm('Teste Manual').subscribe({
      next: (alarm) => {
        this.alarms = [alarm, ...this.alarms];
      }
    });
  }
}
