import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SecurityService, AlarmEvent } from '@core/services/security.service';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-aurora-dark">Notificações</h1>
          <p class="text-aurora-dark/60 mt-1">Histórico de alarmes e eventos de segurança</p>
        </div>
        <div class="flex gap-2">
          <button (click)="triggerTestAlarm()" class="btn-danger">
            <i class="pi pi-bell mr-2"></i>Testar Alarme
          </button>
          <button (click)="loadAlarms()" class="btn-primary">
            <i class="pi pi-refresh mr-2"></i>Atualizar
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div *ngIf="loading" class="flex justify-center py-12">
        <i class="pi pi-spin pi-spinner text-4xl text-aurora-primary"></i>
      </div>

      <!-- Lista de alarmes -->
      <div *ngIf="!loading" class="space-y-3">
        <div *ngFor="let alarm of alarms"
             class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 flex items-start justify-between">
          <div class="flex items-start gap-4">
            <div class="w-10 h-10 rounded-full flex items-center justify-center mt-0.5"
                 [class]="alarm.status === 'triggered' ? 'bg-red-100' : 'bg-gray-100'">
              <i class="pi text-lg"
                 [class]="alarm.status === 'triggered' ? 'pi-exclamation-triangle text-red-500' : 'pi-check text-gray-400'"></i>
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="font-semibold text-aurora-dark">
                  {{ alarm.trigger_type === 'motion' ? 'Movimento Detectado' : 'Alarme Manual' }}
                </h3>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                      [class]="alarm.status === 'triggered' ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-500'">
                  {{ alarm.status === 'triggered' ? 'Ativo' : 'Silenciado' }}
                </span>
              </div>
              <p class="text-sm text-aurora-dark/60 mt-0.5">
                <i class="pi pi-map-marker mr-1"></i>{{ alarm.location || 'Local não especificado' }}
              </p>
              <p class="text-xs text-aurora-dark/40 mt-1">
                {{ alarm.triggered_at | date:'dd/MM/yyyy HH:mm:ss' }}
              </p>
            </div>
          </div>

          <button *ngIf="alarm.status === 'triggered'"
                  (click)="silenceAlarm(alarm)"
                  class="text-xs px-3 py-1.5 rounded-lg bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors whitespace-nowrap">
            <i class="pi pi-volume-off mr-1"></i>Silenciar
          </button>
        </div>

        <!-- Nenhum evento -->
        <div *ngIf="alarms.length === 0" class="bg-white rounded-xl shadow-sm p-12 text-center">
          <i class="pi pi-bell text-5xl text-aurora-dark/20 mb-4"></i>
          <h3 class="text-lg font-medium text-aurora-dark/60">Nenhum evento registrado</h3>
          <p class="text-sm text-aurora-dark/40 mt-1">Os alarmes aparecerão aqui quando forem acionados</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .btn-primary {
      @apply bg-aurora-primary text-white px-4 py-2 rounded-lg text-sm font-medium
             hover:bg-aurora-primary/90 transition-colors flex items-center;
    }
    .btn-danger {
      @apply bg-red-500 text-white px-4 py-2 rounded-lg text-sm font-medium
             hover:bg-red-600 transition-colors flex items-center;
    }
  `]
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
