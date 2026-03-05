import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { LightingService, Light } from '@core/services/lighting.service';
import { LightingSocketService } from '@core/services/lighting-socket.service';

@Component({
  selector: 'app-devices',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-aurora-dark">Dispositivos</h1>
          <p class="text-aurora-dark/60 mt-1">Controle suas luzes e atuadores</p>
        </div>
        <div class="flex items-center gap-3">
          <!-- Indicador de conexão WS -->
          <span class="flex items-center gap-1.5 text-xs font-medium"
                [class]="wsConnected ? 'text-green-600' : 'text-gray-400'">
            <span class="w-2 h-2 rounded-full"
                  [class]="wsConnected ? 'bg-green-400 animate-pulse' : 'bg-gray-300'"></span>
            {{ wsConnected ? 'Tempo real' : 'Desconectado' }}
          </span>
          <!-- Controles globais -->
          <div class="flex items-center gap-2">
            <button
              (click)="turnAllOn()"
              [disabled]="loading || bulkLoading || lights.length === 0"
              class="px-3 py-2 rounded-lg text-xs font-medium bg-yellow-400 text-white hover:bg-yellow-500
                     disabled:opacity-40 disabled:cursor-not-allowed flex items-center">
              <i class="pi pi-sun mr-1"></i>Ligar todas
            </button>
            <button
              (click)="turnAllOff()"
              [disabled]="loading || bulkLoading || lights.length === 0"
              class="px-3 py-2 rounded-lg text-xs font-medium bg-gray-200 text-gray-700 hover:bg-gray-300
                     disabled:opacity-40 disabled:cursor-not-allowed flex items-center">
              <i class="pi pi-moon mr-1"></i>Desligar todas
            </button>
          </div>
          <button (click)="loadLights()" class="btn-primary">
            <i class="pi pi-refresh mr-2"></i>Atualizar
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div *ngIf="loading" class="flex justify-center py-12">
        <i class="pi pi-spin pi-spinner text-4xl text-aurora-primary"></i>
      </div>

      <!-- Grid de luzes -->
      <div *ngIf="!loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div *ngFor="let light of lights"
             class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 transition-all"
             [class.border-yellow-300]="light.state === 'on'"
             [class.bg-yellow-50]="light.state === 'on'">

          <!-- Ícone e nome -->
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-full flex items-center justify-center transition-colors"
                   [class]="light.state === 'on' ? 'bg-yellow-200' : 'bg-gray-100'">
                <i class="pi pi-lightbulb text-2xl transition-colors"
                   [class]="light.state === 'on' ? 'text-yellow-500' : 'text-gray-400'"></i>
              </div>
              <div>
                <h3 class="font-semibold text-aurora-dark">{{ light.name }}</h3>
                <p class="text-sm text-aurora-dark/60">{{ light.location }}</p>
              </div>
            </div>
            <span class="w-3 h-3 rounded-full"
                  [class]="light.state === 'on' ? 'bg-green-400' : 'bg-gray-300'"></span>
          </div>

          <!-- Status -->
          <p class="text-center font-medium mb-4"
             [class]="light.state === 'on' ? 'text-yellow-600' : 'text-gray-400'">
            {{ light.state === 'on' ? 'Ligada' : 'Desligada' }}
          </p>

          <!-- Controles -->
          <div class="flex gap-2">
            <button (click)="turnOn(light)"
                    [disabled]="light.state === 'on' || loadingId === light.id"
                    class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors
                           bg-yellow-400 text-white hover:bg-yellow-500
                           disabled:opacity-40 disabled:cursor-not-allowed">
              <i class="pi pi-sun mr-1"></i>Ligar
            </button>
            <button (click)="turnOff(light)"
                    [disabled]="light.state === 'off' || loadingId === light.id"
                    class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors
                           bg-gray-200 text-gray-700 hover:bg-gray-300
                           disabled:opacity-40 disabled:cursor-not-allowed">
              <i class="pi pi-moon mr-1"></i>Desligar
            </button>
          </div>
        </div>

        <!-- Nenhum dispositivo -->
        <div *ngIf="lights.length === 0" class="col-span-3 bg-white rounded-xl shadow-sm p-12 text-center">
          <i class="pi pi-lightbulb text-5xl text-aurora-dark/20 mb-4"></i>
          <h3 class="text-lg font-medium text-aurora-dark/60">Nenhum dispositivo encontrado</h3>
          <p class="text-sm text-aurora-dark/40 mt-1">Os dispositivos aparecerão aqui quando o ESP32 estiver conectado</p>
        </div>
      </div>

      <!-- Feedback de erro -->
      <div *ngIf="error" class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700 text-sm">
        <i class="pi pi-exclamation-triangle mr-2"></i>{{ error }}
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
