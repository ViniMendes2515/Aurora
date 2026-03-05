import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface AlarmEvent {
  id: string;
  trigger_type: string;
  sensor_id: string;
  location: string;
  status: 'triggered' | 'silenced';
  triggered_at: string;
}

@Injectable({ providedIn: 'root' })
export class SecurityService {
  private readonly baseUrl = environment.securityApiUrl;

  constructor(private http: HttpClient) {}

  getRecentAlarms(): Observable<AlarmEvent[]> {
    return this.http.get<AlarmEvent[]>(`${this.baseUrl}/alarms`);
  }

  triggerAlarm(location: string): Observable<AlarmEvent> {
    return this.http.post<AlarmEvent>(`${this.baseUrl}/alarms/trigger`, { location });
  }

  silenceAlarm(alarmId: string): Observable<AlarmEvent> {
    return this.http.post<AlarmEvent>(`${this.baseUrl}/alarms/${alarmId}/silence`, {});
  }
}
