import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface Schedule {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  schedule_type: 'cron' | 'one_shot';
  cron_expr: string;
  run_at?: string;
  days_filter: number;
  timezone: string;
  action_target: 'nats' | 'http';
  action_type: string;
  action_device_id: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  last_run_at?: string;
  next_run_at?: string;
}

export interface ScheduleExecution {
  id: string;
  schedule_id: string;
  executed_at: string;
  status: 'success' | 'failed' | 'skipped';
  error_message?: string;
}

export interface PreviewResponse {
  schedule_id: string;
  next: string[];
}

export interface CreateScheduleRequest {
  name: string;
  description?: string;
  schedule_type: 'cron' | 'one_shot';
  cron_expr?: string;
  run_at?: string;
  days_filter?: number;
  timezone?: string;
  action_target: 'nats' | 'http';
  action_type: string;
  action_device_id?: string;
}

export type UpdateScheduleRequest = Partial<CreateScheduleRequest>;

@Injectable({ providedIn: 'root' })
export class SchedulesService {
  private readonly baseUrl = environment.schedulesApiUrl;

  constructor(private http: HttpClient) {}

  list(): Observable<Schedule[]> {
    return this.http.get<Schedule[]>(this.baseUrl);
  }

  create(req: CreateScheduleRequest): Observable<Schedule> {
    return this.http.post<Schedule>(this.baseUrl, req);
  }

  update(id: string, req: UpdateScheduleRequest): Observable<Schedule> {
    return this.http.put<Schedule>(`${this.baseUrl}/${id}`, req);
  }

  toggle(id: string): Observable<Schedule> {
    return this.http.patch<Schedule>(`${this.baseUrl}/${id}/toggle`, {});
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  listExecutions(id: string, limit: number): Observable<ScheduleExecution[]> {
    return this.http.get<ScheduleExecution[]>(`${this.baseUrl}/${id}/history?limit=${limit}`);
  }

  preview(scheduleId: string, count: number): Observable<PreviewResponse> {
    return this.http.post<PreviewResponse>(`${this.baseUrl}/preview`, { schedule_id: scheduleId, count });
  }
}
