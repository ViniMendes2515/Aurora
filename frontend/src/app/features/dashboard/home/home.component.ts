import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { CardComponent } from '@shared/ui/card/card.component';
import { SensorsService } from '@core/services/sensors.service';
import { LightingService } from '@core/services/lighting.service';
import { SecurityService } from '@core/services/security.service';
import { RulesService } from '@core/services/rules.service';

interface MetricCard {
  title: string;
  value: string | number;
  icon: string;
  variant: 'primary' | 'secondary' | 'base' | 'dark';
  route?: string;
}

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, CardComponent, RouterModule],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css'
})
export class HomeComponent implements OnInit {
  metrics: MetricCard[] = [];
  currentDate: string = '';
  lastAlarm: string = 'Nenhum';
  lightOnCount = 0;

  ngOnInit(): void {
    this.currentDate = new Date().toLocaleDateString('pt-BR', {
      weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'
    });
    this.loadMetrics();
  }

  constructor(
    private sensorsService: SensorsService,
    private lightingService: LightingService,
    private securityService: SecurityService,
    private rulesService: RulesService
  ) {}

  private loadMetrics(): void {
    forkJoin({
      sensors: this.sensorsService.listSensors().pipe(catchError(() => of([]))),
      lights:  this.lightingService.listLights().pipe(catchError(() => of([]))),
      alarms:  this.securityService.getRecentAlarms().pipe(catchError(() => of([]))),
      rules:   this.rulesService.listRules().pipe(catchError(() => of([])))
    }).subscribe(({ sensors, lights, alarms, rules }) => {
      const activeSensors = sensors.filter((s: any) => s.active).length;
      const lightsOn = lights.filter((l: any) => l.state === 'on').length;
      const activeAlarms = alarms.filter((a: any) => a.status === 'triggered').length;
      const activeRules = rules.filter((r: any) => r.enabled).length;

      if (alarms.length > 0) {
        const last = alarms[0] as any;
        this.lastAlarm = new Date(last.triggered_at).toLocaleString('pt-BR');
      }

      this.lightOnCount = lightsOn;

      this.metrics = [
        {
          title: 'Sensores Ativos',
          value: activeSensors || sensors.length,
          icon: 'pi-wifi',
          variant: 'primary',
          route: '/dashboard/sensors'
        },
        {
          title: 'Luzes Ligadas',
          value: `${lightsOn}/${lights.length}`,
          icon: 'pi-lightbulb',
          variant: 'secondary',
          route: '/dashboard/devices'
        },
        {
          title: 'Alarmes Ativos',
          value: activeAlarms,
          icon: 'pi-bell',
          variant: 'base',
          route: '/dashboard/notifications'
        },
        {
          title: 'Regras Ativas',
          value: activeRules || rules.length,
          icon: 'pi-cog',
          variant: 'dark',
          route: '/dashboard/rules'
        }
      ];
    });
  }
}
