import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { SensorsService, Sensor, MotionRecord, LightRecord } from '@core/services/sensors.service';
import { SensorsSocketService, SensorSocketEvent } from '@core/services/sensors-socket.service';

@Component({
  selector: 'app-sensors',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './sensors.component.html',
  styleUrls: ['./sensors.component.scss']
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
