import { Injectable, OnDestroy } from '@angular/core';
import { Subject, Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface LightStateEvent {
  type: 'light_state_changed';
  light_id: string;
  name: string;
  location: string;
  state: 'on' | 'off';
}

@Injectable({ providedIn: 'root' })
export class LightingSocketService implements OnDestroy {
  private ws: WebSocket | null = null;
  private events$ = new Subject<LightStateEvent>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private destroyed = false;

  get stateChanges(): Observable<LightStateEvent> {
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
      console.log('[WS] Conectado ao lighting-service');
    };

    this.ws.onmessage = (msg) => {
      try {
        const event: LightStateEvent = JSON.parse(msg.data);
        if (event.type === 'light_state_changed') {
          this.events$.next(event);
        }
      } catch {
        console.warn('[WS] Mensagem inválida:', msg.data);
      }
    };

    this.ws.onerror = (err) => {
      console.warn('[WS] Erro na conexão:', err);
    };

    this.ws.onclose = () => {
      console.log('[WS] Conexão encerrada — reconectando em 5s...');
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
    const httpBase = environment.lightingApiUrl;
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${proto}//${host}${httpBase}/ws`;
  }
}
