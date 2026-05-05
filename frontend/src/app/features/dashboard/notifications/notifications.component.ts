import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { SecurityService, AlarmEvent } from '@core/services/security.service';
import {
  NotificationsService,
  TelegramPreference,
  ALL_NOTIFICATION_TYPES
} from '@core/services/notifications.service';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './notifications.component.html',
})
export class NotificationsComponent implements OnInit {
  alarms: AlarmEvent[] = [];
  loadingAlarms = false;

  // Telegram
  telegramPref: TelegramPreference | null = null;
  loadingTelegram = false;
  linkToken: string | null = null;
  linkMessage: string | null = null;
  generatingToken = false;
  unlinking = false;
  savingPrefs = false;
  prefsSaved = false;
  telegramError: string | null = null;

  allTypes = ALL_NOTIFICATION_TYPES;
  selectedTypes: Set<string> = new Set();

  constructor(
    private securityService: SecurityService,
    private notificationsService: NotificationsService,
  ) {}

  ngOnInit(): void {
    this.loadAlarms();
    this.loadTelegramPreferences();
  }

  // ── Alarmes ───────────────────────────────────────────────────────────────

  loadAlarms(): void {
    this.loadingAlarms = true;
    this.securityService.getRecentAlarms().subscribe({
      next: (alarms) => { this.alarms = alarms || []; this.loadingAlarms = false; },
      error: () => { this.loadingAlarms = false; }
    });
  }

  silenceAlarm(alarm: AlarmEvent): void {
    this.securityService.silenceAlarm(alarm.id).subscribe({
      next: () => { alarm.status = 'silenced'; }
    });
  }

  triggerTestAlarm(): void {
    this.securityService.triggerAlarm('Teste Manual').subscribe({
      next: (alarm) => { this.alarms = [alarm, ...this.alarms]; }
    });
  }

  // ── Telegram ──────────────────────────────────────────────────────────────

  loadTelegramPreferences(): void {
    this.loadingTelegram = true;
    this.telegramError = null;
    this.notificationsService.getTelegramPreferences().subscribe({
      next: (pref) => {
        this.telegramPref = pref;
        this.selectedTypes = new Set(pref.enabled_types);
        this.loadingTelegram = false;
      },
      error: (err) => {
        // 404 = não vinculado, não é erro para o usuário
        if (err.status !== 404) {
          this.telegramError = 'Erro ao carregar preferências Telegram.';
        }
        this.loadingTelegram = false;
      }
    });
  }

  generateToken(): void {
    this.generatingToken = true;
    this.linkToken = null;
    this.linkMessage = null;
    this.telegramError = null;
    this.notificationsService.generateLinkToken().subscribe({
      next: (res) => {
        this.linkToken = res.token;
        this.linkMessage = res.message;
        this.generatingToken = false;
      },
      error: () => {
        this.telegramError = 'Erro ao gerar token. Tente novamente.';
        this.generatingToken = false;
      }
    });
  }

  unlinkTelegram(): void {
    this.unlinking = true;
    this.notificationsService.unlinkTelegram().subscribe({
      next: () => {
        this.telegramPref = null;
        this.linkToken = null;
        this.selectedTypes.clear();
        this.unlinking = false;
      },
      error: () => {
        this.telegramError = 'Erro ao desvincular conta.';
        this.unlinking = false;
      }
    });
  }

  toggleType(value: string): void {
    if (this.selectedTypes.has(value)) {
      this.selectedTypes.delete(value);
    } else {
      this.selectedTypes.add(value);
    }
  }

  isTypeSelected(value: string): boolean {
    return this.selectedTypes.has(value);
  }

  savePreferences(): void {
    this.savingPrefs = true;
    this.prefsSaved = false;
    this.notificationsService.updateTelegramPreferences(Array.from(this.selectedTypes)).subscribe({
      next: () => {
        this.savingPrefs = false;
        this.prefsSaved = true;
        setTimeout(() => { this.prefsSaved = false; }, 3000);
        if (this.telegramPref) {
          this.telegramPref.enabled_types = Array.from(this.selectedTypes);
        }
      },
      error: () => {
        this.telegramError = 'Erro ao salvar preferências.';
        this.savingPrefs = false;
      }
    });
  }

  copyToken(): void {
    if (this.linkToken) {
      navigator.clipboard.writeText(`/start ${this.linkToken}`);
    }
  }
}
