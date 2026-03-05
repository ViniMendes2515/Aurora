import { Injectable, OnDestroy } from '@angular/core';
import { Observable, Subject } from 'rxjs';
import { environment } from '@environments/environment';

export type SensorSocketEventType = 'motion_detected' | 'light_updated';

export interface SensorSocketEvent {
  type: SensorSocketEventType;
  sensor_id: string;
  sensor_type: 'motion' | 'light';
  detected_at?: string;
  value?: number;
  raw?: number;
  recorded_at?: string;
}

@Injectable({ providedIn: 'root' })
export class SensorsSocketService implements OnDestroy {
  private ws: WebSocket | null = null;
  private events$ = new Subject<SensorSocketEvent>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private destroyed = false;

  get events(): Observable<SensorSocketEvent> {
    return this.events$.asObservable();
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;

    const wsUrl = this.buildWsUrl();
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('[WS][sensors] Conectado');
    };

    this.ws.onmessage = (msg) => {
      try {
        const event: SensorSocketEvent = JSON.parse(msg.data);
        if (event && (event.type === 'motion_detected' || event.type === 'light_updated')) {
          this.events$.next(event);
        }
      } catch {
        console.warn('[WS][sensors] Mensagem inválida:', msg.data);
      }
    };

    this.ws.onerror = (err) => {
      console.warn('[WS][sensors] Erro na conexão:', err);
    };

    this.ws.onclose = () => {
      console.log('[WS][sensors] Conexão encerrada — tentando reconectar em 5s...');
      if (!this.destroyed) {
        this.reconnectTimer = setTimeout(() => this.connect(), 5000);
      }
    };
  }

  disconnect(): void {
    this.destroyed = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.ws = null;
  }

  ngOnDestroy(): void {
    this.disconnect();
    this.events$.complete();
  }

  private buildWsUrl(): string {
    const httpBase = environment.sensorsApiUrl; // "/api/sensors"
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${proto}//${host}${httpBase}/ws`;
  }
}

