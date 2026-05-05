import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface TelegramPreference {
  id: string;
  user_id: string;
  chat_id: number;
  enabled_types: string[];
  active: boolean;
  linked_at: string;
}

export interface TelegramLinkResponse {
  token: string;
  message: string;
  expires: string;
}

export const ALL_NOTIFICATION_TYPES: { value: string; label: string }[] = [
  { value: 'motion_detected',   label: 'Movimento detectado' },
  { value: 'light_on',          label: 'Luz ligada' },
  { value: 'light_off',         label: 'Luz desligada' },
  { value: 'alarm_triggered',   label: 'Alarme acionado' },
  { value: 'schedule_executed', label: 'Ações agendadas' },
];

@Injectable({ providedIn: 'root' })
export class NotificationsService {
  private readonly baseUrl = '/api/notifications';

  constructor(private http: HttpClient) {}

  generateLinkToken(): Observable<TelegramLinkResponse> {
    return this.http.post<TelegramLinkResponse>(`${this.baseUrl}/telegram/link`, {});
  }

  unlinkTelegram(): Observable<any> {
    return this.http.delete(`${this.baseUrl}/telegram/link`);
  }

  getTelegramPreferences(): Observable<TelegramPreference> {
    return this.http.get<TelegramPreference>(`${this.baseUrl}/telegram/preferences`);
  }

  updateTelegramPreferences(enabledTypes: string[]): Observable<any> {
    return this.http.put(`${this.baseUrl}/telegram/preferences`, { enabled_types: enabledTypes });
  }
}
