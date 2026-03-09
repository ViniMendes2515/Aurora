import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { LightingService, Light } from '@core/services/lighting.service';
import { LightingSocketService } from '@core/services/lighting-socket.service';

@Component({
  selector: 'app-devices',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './devices.component.html',
  styleUrls: ['./devices.component.scss']
})
export class DevicesComponent implements OnInit, OnDestroy {
  lights: Light[] = [];
  loading = false;
  bulkLoading = false;
  loadingId: string | null = null;
  error: string | null = null;
  wsConnected = false;

  private wsSub?: Subscription;

  constructor(
    private lightingService: LightingService,
    private socketService: LightingSocketService
  ) {}

  ngOnInit(): void {
    this.loadLights();
    this.connectWebSocket();
  }

  ngOnDestroy(): void {
    this.wsSub?.unsubscribe();
    this.socketService.disconnect();
  }

  private connectWebSocket(): void {
    this.socketService.connect();

    // Monitora o status da conexão verificando periodicamente
    const checkInterval = setInterval(() => {
      this.wsConnected = this.socketService.isConnected;
    }, 1000);

    this.wsSub = this.socketService.stateChanges.subscribe(event => {
      this.wsConnected = true;
      const light = this.lights.find(l => l.id === event.light_id);
      if (light) {
        light.state = event.state;
      }
    });

    // Limpa o interval quando o componente for destruído
    this.wsSub.add(() => clearInterval(checkInterval));
  }

  loadLights(): void {
    this.loading = true;
    this.error = null;
    this.lightingService.listLights().subscribe({
      next: (lights) => {
        this.lights = lights || [];
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.error = 'Não foi possível carregar os dispositivos. Verifique se os serviços estão rodando.';
      }
    });
  }

  turnOn(light: Light): void {
    this.loadingId = light.id;
    this.error = null;
    this.lightingService.turnOn(light.id).subscribe({
      next: () => {
        light.state = 'on';
        this.loadingId = null;
      },
      error: () => {
        this.loadingId = null;
        this.error = `Não foi possível ligar "${light.name}". Verifique se o ESP32 está conectado.`;
      }
    });
  }

  turnOff(light: Light): void {
    this.loadingId = light.id;
    this.error = null;
    this.lightingService.turnOff(light.id).subscribe({
      next: () => {
        light.state = 'off';
        this.loadingId = null;
      },
      error: () => {
        this.loadingId = null;
        this.error = `Não foi possível desligar "${light.name}". Verifique se o ESP32 está conectado.`;
      }
    });
  }

  turnAllOn(): void {
    const targets = this.lights.filter(l => l.state === 'off');
    if (targets.length === 0) return;

    this.bulkLoading = true;
    this.error = null;

    let remaining = targets.length;
    const finalize = () => {
      remaining--;
      if (remaining <= 0) {
        this.bulkLoading = false;
      }
    };

    for (const light of targets) {
      this.lightingService.turnOn(light.id).subscribe({
        next: () => {
          light.state = 'on';
          finalize();
        },
        error: () => {
          this.error = 'Não foi possível ligar todas as luzes. Verifique se o ESP32 está conectado.';
          finalize();
        }
      });
    }
  }

  turnAllOff(): void {
    const targets = this.lights.filter(l => l.state === 'on');
    if (targets.length === 0) return;

    this.bulkLoading = true;
    this.error = null;

    let remaining = targets.length;
    const finalize = () => {
      remaining--;
      if (remaining <= 0) {
        this.bulkLoading = false;
      }
    };

    for (const light of targets) {
      this.lightingService.turnOff(light.id).subscribe({
        next: () => {
          light.state = 'off';
          finalize();
        },
        error: () => {
          this.error = 'Não foi possível desligar todas as luzes. Verifique se o ESP32 está conectado.';
          finalize();
        }
      });
    }
  }
}
