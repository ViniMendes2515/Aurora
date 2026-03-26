import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  SchedulesService,
  Schedule,
  ScheduleExecution,
  PreviewResponse,
  CreateScheduleRequest
} from '@core/services/schedules.service';
import { LightingService, Light } from '@core/services/lighting.service';

// Repeat options map to cron weekday syntax
type RepeatMode = 'daily' | 'weekdays' | 'weekends' | 'custom';

const WEEKDAYS = [
  { label: 'Dom', cronDay: 0 },
  { label: 'Seg', cronDay: 1 },
  { label: 'Ter', cronDay: 2 },
  { label: 'Qua', cronDay: 3 },
  { label: 'Qui', cronDay: 4 },
  { label: 'Sex', cronDay: 5 },
  { label: 'Sáb', cronDay: 6 }
];

@Component({
  selector: 'app-schedules',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './schedules.component.html'
})
export class SchedulesComponent implements OnInit {
  schedules: Schedule[] = [];
  lights: Light[] = [];
  loading = false;
  error: string | null = null;
  showForm = false;
  editingId: string | null = null;

  selectedSchedule: Schedule | null = null;
  activeTab: 'history' | 'preview' = 'history';
  executions: ScheduleExecution[] = [];
  loadingExecutions = false;
  previewDates: string[] = [];
  loadingPreview = false;

  readonly weekdays = WEEKDAYS;

  // User-friendly form fields
  formName = '';
  formScheduleType: 'cron' | 'one_shot' = 'cron';
  formHour = '07';
  formMinute = '00';
  formRepeat: RepeatMode = 'daily';
  formCustomDays: boolean[] = [false, false, false, false, false, false, false];
  formRunAt = '';
  formActionType = 'turn_on_light';
  formDeviceId = '';

  readonly hours = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'));
  readonly minutes = ['00', '05', '10', '15', '20', '25', '30', '35', '40', '45', '50', '55'];

  readonly scheduleTypeOptions: { v: 'cron' | 'one_shot'; l: string }[] = [
    { v: 'cron', l: 'Repetir' },
    { v: 'one_shot', l: 'Uma vez' }
  ];

  readonly repeatOptions: { v: RepeatMode; l: string }[] = [
    { v: 'daily', l: 'Todos os dias' },
    { v: 'weekdays', l: 'Dias úteis' },
    { v: 'weekends', l: 'Fim de semana' },
    { v: 'custom', l: 'Dias específicos' }
  ];

  readonly actionOptions: { v: string; l: string; icon: string }[] = [
    { v: 'turn_on_light', l: 'Ligar luz', icon: 'pi-sun' },
    { v: 'turn_off_light', l: 'Desligar luz', icon: 'pi-moon' },
    { v: 'trigger_alarm', l: 'Acionar alarme', icon: 'pi-bell' },
    { v: 'silence_alarm', l: 'Silenciar alarme', icon: 'pi-bell-slash' }
  ];

  constructor(
    private schedulesService: SchedulesService,
    private lightingService: LightingService
  ) {}

  ngOnInit(): void {
    this.loadSchedules();
    this.loadLights();
  }

  loadLights(): void {
    this.lightingService.listLights().subscribe({
      next: (data) => this.lights = data || [],
      error: () => this.lights = []
    });
  }

  get isLightAction(): boolean {
    return this.formActionType === 'turn_on_light' || this.formActionType === 'turn_off_light';
  }

  get isAlarmAction(): boolean {
    return this.formActionType === 'trigger_alarm' || this.formActionType === 'silence_alarm';
  }

  get deviceLabel(): string {
    if (this.isLightAction) return 'Qual luz?';
    if (this.isAlarmAction) return 'Local do alarme';
    return 'Dispositivo';
  }

  get devicePlaceholder(): string {
    if (this.isAlarmAction) return 'Ex: sala, entrada, garagem';
    return '';
  }

  private buildCronExpr(): string {
    const m = this.formMinute;
    const h = this.formHour;
    switch (this.formRepeat) {
      case 'daily': return `${m} ${h} * * *`;
      case 'weekdays': return `${m} ${h} * * 1-5`;
      case 'weekends': return `${m} ${h} * * 0,6`;
      case 'custom': {
        const selected = WEEKDAYS
          .filter((_, i) => this.formCustomDays[i])
          .map(d => d.cronDay);
        const dayStr = selected.length > 0 ? selected.join(',') : '*';
        return `${m} ${h} * * ${dayStr}`;
      }
    }
  }

  private parseCronIntoForm(cron: string): void {
    const parts = cron.trim().split(/\s+/);
    if (parts.length < 5) return;
    this.formMinute = parts[0].padStart(2, '0');
    this.formHour = parts[1].padStart(2, '0');
    const dayPart = parts[4];
    if (dayPart === '*') {
      this.formRepeat = 'daily';
    } else if (dayPart === '1-5') {
      this.formRepeat = 'weekdays';
    } else if (dayPart === '0,6' || dayPart === '6,0') {
      this.formRepeat = 'weekends';
    } else {
      this.formRepeat = 'custom';
      const nums = dayPart.split(',').map(Number);
      this.formCustomDays = WEEKDAYS.map(d => nums.includes(d.cronDay));
    }
  }

  private resetForm(): void {
    this.formName = '';
    this.formScheduleType = 'cron';
    this.formHour = '07';
    this.formMinute = '00';
    this.formRepeat = 'daily';
    this.formCustomDays = [false, false, false, false, false, false, false];
    this.formRunAt = '';
    this.formActionType = 'turn_on_light';
    this.formDeviceId = '';
  }

  loadSchedules(): void {
    this.loading = true;
    this.error = null;
    this.schedulesService.list().subscribe({
      next: (data) => {
        this.schedules = data || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = err?.error?.message || 'Erro ao carregar agendamentos.';
        this.loading = false;
      }
    });
  }

  toggleForm(): void {
    this.showForm = !this.showForm;
    if (!this.showForm) {
      this.editingId = null;
      this.resetForm();
    }
  }

  startEdit(schedule: Schedule, event: Event): void {
    event.stopPropagation();
    this.editingId = schedule.id;
    this.showForm = true;
    this.formName = schedule.name;
    this.formScheduleType = schedule.schedule_type;
    this.formActionType = schedule.action_type;
    this.formDeviceId = schedule.action_device_id || '';
    this.formRunAt = schedule.run_at || '';
    if (schedule.schedule_type === 'cron' && schedule.cron_expr) {
      this.parseCronIntoForm(schedule.cron_expr);
    }
  }

  canSubmit(): boolean {
    if (!this.formName.trim()) return false;
    if (this.formScheduleType === 'cron' && this.formRepeat === 'custom') {
      if (!this.formCustomDays.some(Boolean)) return false;
    }
    if (this.formScheduleType === 'one_shot' && !this.formRunAt) return false;
    if (this.isLightAction && !this.formDeviceId) return false;
    return true;
  }

  submitForm(): void {
    const payload: CreateScheduleRequest = {
      name: this.formName.trim(),
      schedule_type: this.formScheduleType,
      action_target: 'http',
      action_type: this.formActionType,
      action_device_id: this.formDeviceId,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'America/Sao_Paulo',
      days_filter: 0,
      ...(this.formScheduleType === 'cron'
        ? { cron_expr: this.buildCronExpr() }
        : { run_at: this.formRunAt ? new Date(this.formRunAt).toISOString() : undefined })
    };

    if (this.editingId) {
      this.schedulesService.update(this.editingId, payload).subscribe({
        next: (updated) => {
          this.schedules = this.schedules.map(s => s.id === updated.id ? updated : s);
          if (this.selectedSchedule?.id === updated.id) this.selectedSchedule = updated;
          this.showForm = false;
          this.editingId = null;
          this.resetForm();
        },
        error: (err) => {
          this.error = err?.error?.message || 'Erro ao atualizar agendamento.';
        }
      });
    } else {
      this.schedulesService.create(payload).subscribe({
        next: (created) => {
          this.schedules = [created, ...this.schedules];
          this.showForm = false;
          this.resetForm();
        },
        error: (err) => {
          this.error = err?.error?.message || 'Erro ao criar agendamento.';
        }
      });
    }
  }

  toggleSchedule(schedule: Schedule, event: Event): void {
    event.stopPropagation();
    this.schedulesService.toggle(schedule.id).subscribe({
      next: (updated) => {
        this.schedules = this.schedules.map(s => s.id === updated.id ? updated : s);
        if (this.selectedSchedule?.id === updated.id) this.selectedSchedule = updated;
      },
      error: (err) => {
        this.error = err?.error?.message || 'Erro ao alterar status.';
      }
    });
  }

  deleteSchedule(id: string, event: Event): void {
    event.stopPropagation();
    this.schedulesService.delete(id).subscribe({
      next: () => {
        this.schedules = this.schedules.filter(s => s.id !== id);
        if (this.selectedSchedule?.id === id) this.selectedSchedule = null;
      },
      error: (err) => {
        this.error = err?.error?.message || 'Erro ao deletar agendamento.';
      }
    });
  }

  selectSchedule(schedule: Schedule): void {
    if (this.selectedSchedule?.id === schedule.id) {
      this.selectedSchedule = null;
      return;
    }
    this.selectedSchedule = schedule;
    this.activeTab = 'history';
    this.executions = [];
    this.previewDates = [];
    this.loadHistory();
  }

  closeDetail(): void {
    this.selectedSchedule = null;
  }

  setTab(tab: 'history' | 'preview'): void {
    this.activeTab = tab;
    if (tab === 'history' && this.executions.length === 0) this.loadHistory();
  }

  loadHistory(): void {
    if (!this.selectedSchedule) return;
    this.loadingExecutions = true;
    this.schedulesService.listExecutions(this.selectedSchedule.id, 20).subscribe({
      next: (data) => {
        this.executions = data || [];
        this.loadingExecutions = false;
      },
      error: () => {
        this.executions = [];
        this.loadingExecutions = false;
      }
    });
  }

  loadPreview(): void {
    if (!this.selectedSchedule) return;
    this.loadingPreview = true;
    this.schedulesService.preview(this.selectedSchedule.id, 5).subscribe({
      next: (data: PreviewResponse) => {
        this.previewDates = data.next || [];
        this.loadingPreview = false;
      },
      error: () => {
        this.previewDates = [];
        this.loadingPreview = false;
      }
    });
  }

  describeSchedule(schedule: Schedule): string {
    if (schedule.schedule_type === 'one_shot') {
      return `Uma vez — ${this.formatDate(schedule.run_at)}`;
    }
    const parts = schedule.cron_expr?.trim().split(/\s+/) || [];
    if (parts.length < 5) return schedule.cron_expr || '';
    const m = parts[0];
    const h = parts[1];
    const dayPart = parts[4];
    const time = `${h.padStart(2,'0')}:${m.padStart(2,'0')}`;
    if (dayPart === '*') return `Todos os dias às ${time}`;
    if (dayPart === '1-5') return `Dias úteis às ${time}`;
    if (dayPart === '0,6' || dayPart === '6,0') return `Finais de semana às ${time}`;
    const names = ['Dom','Seg','Ter','Qua','Qui','Sex','Sáb'];
    const days = dayPart.split(',').map(n => names[Number(n)] || n).join(', ');
    return `${days} às ${time}`;
  }

  lightName(deviceId: string): string {
    const l = this.lights.find(x => x.id === deviceId);
    return l ? `${l.name} (${l.location})` : deviceId;
  }

  actionLabel(type: string): string {
    const map: Record<string, string> = {
      turn_on_light: 'Ligar luz',
      turn_off_light: 'Desligar luz',
      trigger_alarm: 'Acionar alarme',
      silence_alarm: 'Silenciar alarme'
    };
    return map[type] || type;
  }

  formatDate(iso?: string): string {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('pt-BR', {
      day: '2-digit', month: '2-digit', year: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }
}
