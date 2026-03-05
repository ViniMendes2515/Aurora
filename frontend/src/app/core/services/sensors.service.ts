import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface Sensor {
  ID: string;
  Name: string;
  Type: string;
  Location: string;
  Owner_id: string;
  Active: boolean;
  Created_at: string;
  Updated_at: string;
}

export interface MotionRecord {
  ID: string;
  SensorID: string;
  UserID: string;
  DetectedAt: string;
}

export interface LightRecord {
  ID: string;
  SensorID: string;
  Value: number;
  Raw: number;
  RecordedAt: string;
}

@Injectable({ providedIn: 'root' })
export class SensorsService {
  private readonly baseUrl = environment.sensorsApiUrl;

  constructor(private http: HttpClient) {}

  listSensors(): Observable<Sensor[]> {
    return this.http.get<Sensor[]>(`${this.baseUrl}/sensors`);
  }

  getMotionHistory(sensorId: string): Observable<MotionRecord[]> {
    return this.http.get<MotionRecord[]>(`${this.baseUrl}/sensors/${sensorId}/motion`);
  }

  getLightHistory(sensorId: string): Observable<LightRecord[]> {
    return this.http.get<LightRecord[]>(`${this.baseUrl}/sensors/${sensorId}/light`);
  }

  triggerMotion(sensorId: string): Observable<any> {
    return this.http.post(`${this.baseUrl}/sensors/${sensorId}/motion`, {});
  }
}
