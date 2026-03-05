import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { SensorsService, Sensor, MotionRecord, LightRecord } from '@core/services/sensors.service';
import { SensorsSocketService, SensorSocketEvent } from '@core/services/sensors-socket.service';

@Component({
  selector: 'app-sensors',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-aurora-dark">Sensores</h1>
          <p class="text-aurora-dark/60 mt-1">Monitore seus sensores em tempo real</p>
        </div>
        <button (click)="loadSensors()" class="btn-primary">
          <i class="pi pi-refresh mr-2"></i>Atualizar
        </button>
      </div>

      <!-- Loading -->
      <div *ngIf="loading" class="flex justify-center py-12">
        <i class="pi pi-spin pi-spinner text-4xl text-aurora-primary"></i>
      </div>

      <!-- Lista de sensores -->
      <div *ngIf="!loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div *ngFor="let sensor of sensors"
             class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 cursor-pointer hover:shadow-md transition-shadow"
             [class.border-aurora-primary]="selectedSensor?.ID === sensor.ID"
             (click)="selectSensor(sensor)">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full flex items-center justify-center"
                   [class]="sensor.Type === 'motion' ? 'bg-orange-100' : 'bg-yellow-100'">
                <i class="pi text-xl"
                   [class]="sensor.Type === 'motion' ? 'pi-eye text-orange-500' : 'pi-sun text-yellow-500'"></i>
              </div>
              <div>
                <h3 class="font-semibold text-aurora-dark">{{ sensor.Name }}</h3>
                <p class="text-sm text-aurora-dark/60">{{ sensor.Location }}</p>
              </div>
            </div>
            <span class="px-2 py-1 rounded-full text-xs font-medium"
                  [class]="sensor.Active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
              {{ sensor.Active ? 'Ativo' : 'Inativo' }}
            </span>
          </div>

          <div class="mt-4 flex gap-4 text-sm text-aurora-dark/60">
            <span class="capitalize">
              <i class="pi pi-tag mr-1"></i>{{ sensor.Type === 'motion' ? 'Presença' : 'Luminosidade' }}
            </span>
            <span>
              <i class="pi pi-map-marker mr-1"></i>{{ sensor.Location }}
            </span>
          </div>
        </div>

        <!-- Nenhum sensor -->
        <div *ngIf="sensors.length === 0" class="col-span-2 bg-white rounded-xl shadow-sm p-12 text-center">
          <i class="pi pi-wifi text-5xl text-aurora-dark/20 mb-4"></i>
          <h3 class="text-lg font-medium text-aurora-dark/60">Nenhum sensor encontrado</h3>
          <p class="text-sm text-aurora-dark/40 mt-1">Os sensores aparecerão aqui quando forem registrados</p>
        </div>
      </div>

      <!-- Histórico do sensor selecionado -->
      <div *ngIf="selectedSensor" class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold text-aurora-dark">
            Histórico — {{ selectedSensor.Name }}
          </h2>
          <span class="flex items-center gap-1.5 text-xs font-medium"
                [class]="wsConnected ? 'text-green-600' : 'text-gray-400'">
            <span class="w-2 h-2 rounded-full"
                  [class]="wsConnected ? 'bg-green-400 animate-pulse' : 'bg-gray-300'"></span>
            {{ wsConnected ? 'Tempo real' : 'Offline' }}
          </span>
        </div>

        <!-- Histórico de movimento -->
        <div *ngIf="selectedSensor.Type === 'motion'">
          <div *ngIf="motionRecords.length === 0" class="text-center py-8 text-aurora-dark/40">
            <i class="pi pi-clock text-3xl mb-2"></i>
            <p>Nenhum registro de movimento</p>
          </div>
          <div *ngFor="let record of motionRecords"
               class="flex items-center gap-3 py-3 border-b border-gray-50 last:border-0">
            <div class="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
              <i class="pi pi-eye text-orange-500 text-sm"></i>
            </div>
            <div>
              <p class="font-medium text-aurora-dark text-sm">Movimento detectado</p>
              <p class="text-xs text-aurora-dark/50">{{ record.DetectedAt | date:'dd/MM/yyyy HH:mm:ss' }}</p>
            </div>
          </div>
        </div>

        <!-- Histórico de luminosidade -->
        <div *ngIf="selectedSensor.Type === 'light'">
          <div *ngIf="lightRecords.length === 0" class="text-center py-8 text-aurora-dark/40">
            <i class="pi pi-sun text-3xl mb-2"></i>
            <p>Nenhum registro de luminosidade</p>
          </div>
          <div *ngFor="let record of lightRecords"
               class="flex items-center justify-between py-3 border-b border-gray-50 last:border-0">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center">
                <i class="pi pi-sun text-yellow-500 text-sm"></i>
              </div>
              <div>
                <p class="font-medium text-aurora-dark text-sm">Leitura de luminosidade</p>
                <p class="text-xs text-aurora-dark/50">{{ record.RecordedAt | date:'dd/MM/yyyy HH:mm:ss' }}</p>
              </div>
            </div>
            <div class="text-right">
              <span class="text-lg font-bold" [class]="record.Value < 30 ? 'text-orange-500' : 'text-green-500'">
                {{ record.Value | number:'1.0-1' }}%
              </span>
              <p class="text-xs text-aurora-dark/40">Raw: {{ record.Raw }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .btn-primary {
      @apply bg-aurora-primary text-white px-4 py-2 rounded-lg text-sm font-medium
             hover:bg-aurora-primary/90 transition-colors flex items-center;
    }
  `]
})
export class SensorsComponent implements OnInit, OnDestroy {
  sensors: Sensor[] = [];
  selectedSensor: Sensor | null = null;
  motionRecords: MotionRecord[] = [];
  lightRecords: LightRecord[] = [];
  loading = false;
  wsConnected = false;

  private wsSub?: Subscription;

  constructor(
    private sensorsService: SensorsService,
    private socketService: SensorsSocketService
  ) {}

  ngOnInit(): void {
    this.loadSensors();
    this.connectWebSocket();
  }

  ngOnDestroy(): void {
    this.wsSub?.unsubscribe();
    this.socketService.disconnect();
  }

  loadSensors(): void {
    this.loading = true;
    this.sensorsService.listSensors().subscribe({
      next: (sensors) => {
        this.sensors = sensors || [];
        this.loading = false;
      },
      error: () => {
        this.loading = false;
      }
    });
  }

  private connectWebSocket(): void {
    this.socketService.connect();

    const checkInterval = setInterval(() => {
      this.wsConnected = this.socketService.isConnected;
    }, 1000);

    this.wsSub = this.socketService.events.subscribe((event: SensorSocketEvent) => {
      this.wsConnected = true;

      // Atualiza somente o sensor atualmente selecionado, se for o mesmo ID
      if (!this.selectedSensor || this.selectedSensor.ID !== event.sensor_id) {
        return;
      }

      if (event.type === 'motion_detected' && event.detected_at) {
        const record: MotionRecord = {
          ID: crypto.randomUUID(),
          SensorID: event.sensor_id,
          UserID: 'device',
          DetectedAt: event.detected_at
        };
        this.motionRecords = [record, ...this.motionRecords].slice(0, 20);
      }

      if (event.type === 'light_updated' && event.recorded_at != null && event.value != null && event.raw != null) {
        const record: LightRecord = {
          ID: crypto.randomUUID(),
          SensorID: event.sensor_id,
          Value: event.value,
          Raw: event.raw,
          RecordedAt: event.recorded_at
        };
        this.lightRecords = [record, ...this.lightRecords].slice(0, 20);
      }
    });

    this.wsSub.add(() => clearInterval(checkInterval));
  }

  selectSensor(sensor: Sensor): void {
    this.selectedSensor = sensor;
    this.motionRecords = [];
    this.lightRecords = [];

    if (sensor.Type === 'motion') {
      this.sensorsService.getMotionHistory(sensor.ID).subscribe({
        next: (records) => this.motionRecords = records || []
      });
    } else if (sensor.Type === 'light') {
      this.sensorsService.getLightHistory(sensor.ID).subscribe({
        next: (records) => this.lightRecords = records || []
      });
    }
  }
}
