import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { forkJoin, of, Subscription } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { CardComponent } from '@shared/ui/card/card.component';
import { SensorsService } from '@core/services/sensors.service';
import { LightingService } from '@core/services/lighting.service';
import { SecurityService } from '@core/services/security.service';
import { RulesService } from '@core/services/rules.service';
import { LightingSocketService } from '@core/services/lighting-socket.service';
import { SensorsSocketService } from '@core/services/sensors-socket.service';

interface MetricCard {
  title: string;
  value: string | number;
  icon: string;
  variant: 'primary' | 'secondary' | 'base' | 'dark';
  route?: string;
}

export interface ActivityItem {
  icon: string;
  iconVariant: 'primary' | 'secondary' | 'base' | 'dark' | 'alarm' | 'light-on' | 'light-off';
  text: string;
  time: Date;
  isNew?: boolean;
}

const MAX_ACTIVITIES = 5;

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, CardComponent, RouterModule],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css'
})
export class HomeComponent implements OnInit, OnDestroy {
  metrics: MetricCard[] = [];
  activities: ActivityItem[] = [];
  currentDate: string = '';
  loadingActivities = true;

  private subs: Subscription[] = [];

  constructor(
    private sensorsService: SensorsService,
    private lightingService: LightingService,
    private securityService: SecurityService,
    private rulesService: RulesService,
    private lightingSocket: LightingSocketService,
    private sensorsSocket: SensorsSocketService
  ) {}

  ngOnInit(): void {
    this.currentDate = new Date().toLocaleDateString('pt-BR', {
      weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'
    });
    this.loadMetrics();
    this.loadInitialActivities();
    this.connectWebSockets();
  }

  ngOnDestroy(): void {
    this.subs.forEach(s => s.unsubscribe());
    this.lightingSocket.disconnect();
    this.sensorsSocket.disconnect();
  }

  private loadMetrics(): void {
    forkJoin({
      sensors: this.sensorsService.listSensors().pipe(catchError(() => of([]))),
      lights:  this.lightingService.listLights().pipe(catchError(() => of([]))),
      alarms:  this.securityService.getRecentAlarms().pipe(catchError(() => of([]))),
      rules:   this.rulesService.listRules().pipe(catchError(() => of([])))
    }).subscribe(({ sensors, lights, alarms, rules }) => {
      const activeSensors = sensors?.filter((s: any) => s.Active).length ?? sensors.length;
      const lightsOn      = lights?.filter((l: any) => l.state === 'on').length ?? 0;
      const activeAlarms  = alarms?.filter((a: any) => a.status === 'triggered').length ?? 0;
      const activeRules   = rules?.filter((r: any) => r.enabled).length ?? rules?.length ?? 0;

      this.metrics = [
        { title: 'Sensores Ativos', value: activeSensors, icon: 'pi-wifi',      variant: 'primary',   route: '/dashboard/sensors' },
        { title: 'Luzes Ligadas',   value: `${lightsOn}/${lights.length}`,       icon: 'pi-lightbulb', variant: 'secondary', route: '/dashboard/devices' },
        { title: 'Alarmes Ativos',  value: activeAlarms,  icon: 'pi-bell',       variant: 'base',      route: '/dashboard/notifications' },
        { title: 'Regras Ativas',   value: activeRules,   icon: 'pi-cog',        variant: 'dark',      route: '/dashboard/rules' }
      ];
    });
  }

  private loadInitialActivities(): void {
    this.loadingActivities = true;

    this.securityService.getRecentAlarms().pipe(catchError(() => of([]))).subscribe((alarms: any[]) => {
      const alarmItems: ActivityItem[] = (alarms || []).map(a => ({
        icon: a.status === 'silenced' ? 'pi-lock' : 'pi-bell',
        iconVariant: a.status === 'silenced' ? 'dark' : 'alarm',
        text: a.status === 'silenced'
          ? `Alarme silenciado em ${a.location}`
          : `Alarme disparado em ${a.location} (${a.trigger_type})`,
        time: new Date(a.triggered_at)
      }));

      this.mergeActivities(alarmItems);
      this.loadingActivities = false;
    });
  }

  private connectWebSockets(): void {
    this.lightingSocket.connect();
    this.sensorsSocket.connect();

    const lightSub = this.lightingSocket.stateChanges.subscribe(event => {
      this.pushActivity({
        icon: event.state === 'on' ? 'pi-lightbulb' : 'pi-moon',
        iconVariant: event.state === 'on' ? 'light-on' : 'light-off',
        text: event.state === 'on'
          ? `Luz "${event.name}" ligada em ${event.location}`
          : `Luz "${event.name}" desligada em ${event.location}`,
        time: new Date(),
        isNew: true
      });
    });

    const sensorSub = this.sensorsSocket.events.subscribe(event => {
      if (event.type === 'motion_detected') {
        this.pushActivity({
          icon: 'pi-wifi',
          iconVariant: 'primary',
          text: `Movimento detectado pelo sensor ${event.sensor_id.slice(0, 8)}`,
          time: event.detected_at ? new Date(event.detected_at) : new Date(),
          isNew: true
        });
      }
    });

    this.subs.push(lightSub, sensorSub);
  }

  private pushActivity(item: ActivityItem): void {
    this.activities = [item, ...this.activities].slice(0, MAX_ACTIVITIES);
    setTimeout(() => { item.isNew = false; }, 4000);
  }

  private mergeActivities(items: ActivityItem[]): void {
    this.activities = [...items, ...this.activities]
      .sort((a, b) => b.time.getTime() - a.time.getTime())
      .slice(0, MAX_ACTIVITIES);
  }

  formatTime(date: Date): string {
    const diff = Math.floor((Date.now() - date.getTime()) / 1000);
    if (diff < 60)   return 'Agora mesmo';
    if (diff < 3600) return `Há ${Math.floor(diff / 60)} min`;
    if (diff < 86400) return `Há ${Math.floor(diff / 3600)}h`;
    return date.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
  }
}
