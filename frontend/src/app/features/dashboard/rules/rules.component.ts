import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RulesService, Rule, CreateRuleRequest } from '@core/services/rules.service';
import { LightingService, Light } from '@core/services/lighting.service';

@Component({
  selector: 'app-rules',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-aurora-dark">Regras</h1>
          <p class="text-aurora-dark/60 mt-1">Automatize seu ambiente com regras inteligentes</p>
        </div>
        <button (click)="toggleForm()" class="btn-primary">
          <i class="pi mr-2" [class]="showForm ? 'pi-times' : 'pi-plus'"></i>
          {{ showForm ? 'Cancelar' : 'Nova Regra' }}
        </button>
      </div>

      <!-- Formulário de nova regra -->
      <div *ngIf="showForm" class="bg-white rounded-xl shadow-sm border border-aurora-primary/20 p-6">
        <h2 class="text-lg font-semibold text-aurora-dark mb-4">
          {{ editingRuleId ? 'Editar Regra de Automação' : 'Nova Regra de Automação' }}
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Nome da regra</label>
            <input [(ngModel)]="newRule.name" type="text" placeholder="Ex: Acender luz quando escurecer"
                   class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
          </div>
          <div>
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Gatilho</label>
            <select [(ngModel)]="newRule.trigger_type"
                    class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
              <option value="motion">Detecção de presença (PIR)</option>
              <option value="light_low">Luminosidade baixa (LDR)</option>
              <option value="light_high">Luminosidade alta (LDR)</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Sensor gatilho</label>
            <input [(ngModel)]="newRule.trigger_id" type="text"
                   [placeholder]="newRule.trigger_type === 'light_low' || newRule.trigger_type === 'light_high'
                     ? 'Ex: esp32-ldr-001' : 'Ex: esp32-pir-001'"
                   class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
          </div>
          <div *ngIf="newRule.trigger_type === 'light_low' || newRule.trigger_type === 'light_high'">
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">
              {{ newRule.trigger_type === 'light_low'
                ? 'Considerar luz baixa abaixo de (%)'
                : 'Considerar luz alta acima de (%)' }}
            </label>
            <input [(ngModel)]="newRule.trigger_threshold" type="number" min="1" max="100" step="1"
                   placeholder="30"
                   class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
            <p class="text-xs text-aurora-dark/50 mt-0.5">
              0–100.
              <ng-container *ngIf="newRule.trigger_type === 'light_low'; else highHint">
                Ex: 30 = acionar quando luminosidade &lt; 30%
              </ng-container>
              <ng-template #highHint>
                Ex: 70 = acionar quando luminosidade &gt; 70%
              </ng-template>
            </p>
          </div>
          <div>
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Ação</label>
            <select [(ngModel)]="newRule.action_type"
                    class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
              <option value="turn_on_light">Ligar luz</option>
              <option value="turn_off_light">Desligar luz</option>
              <option value="trigger_alarm">Acionar alarme</option>
            </select>
          </div>
          <div *ngIf="newRule.action_type === 'turn_on_light' || newRule.action_type === 'turn_off_light'">
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Qual luz?</label>
            <select [(ngModel)]="newRule.action_id"
                    class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
              <option value="">Selecione uma luz</option>
              <option *ngFor="let light of lights" [value]="light.id">{{ light.name }} ({{ light.location }})</option>
            </select>
          </div>
          <div *ngIf="newRule.action_type === 'trigger_alarm'">
            <label class="block text-sm font-medium text-aurora-dark/70 mb-1">Dispositivo alvo</label>
            <input [(ngModel)]="newRule.action_id" type="text" placeholder="alarme"
                   class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-aurora-primary">
          </div>
        </div>
        <div class="mt-4 flex justify-end">
          <button (click)="createRule()" [disabled]="!canCreate()" class="btn-primary">
            <i class="pi pi-check mr-2"></i>{{ editingRuleId ? 'Salvar Alterações' : 'Criar Regra' }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div *ngIf="loading" class="flex justify-center py-12">
        <i class="pi pi-spin pi-spinner text-4xl text-aurora-primary"></i>
      </div>

      <!-- Lista de regras -->
      <div *ngIf="!loading" class="space-y-3">
        <div *ngFor="let rule of rules"
             class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 bg-aurora-primary/10 rounded-full flex items-center justify-center">
              <i class="pi pi-cog text-aurora-primary"></i>
            </div>
            <div>
              <h3 class="font-semibold text-aurora-dark">{{ rule.name }}</h3>
              <p class="text-sm text-aurora-dark/60 mt-0.5">
                <span class="font-medium">Se:</span> {{ getTriggerLabel(rule) }}
                <span class="text-aurora-dark/40 mx-1">→</span>
                <span class="font-medium">Então:</span> {{ getActionLabel(rule.action_type) }}
                <span *ngIf="rule.action_id" class="text-aurora-dark/40"> ({{ getActionTargetLabel(rule) }})</span>
              </p>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                  [class]="rule.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
              {{ rule.enabled ? 'Ativa' : 'Inativa' }}
            </span>
            <button (click)="startEdit(rule)" class="text-aurora-primary hover:text-aurora-primary/80 transition-colors mr-2">
              <i class="pi pi-pencil"></i>
            </button>
            <button (click)="deleteRule(rule.id)" class="text-red-400 hover:text-red-600 transition-colors">
              <i class="pi pi-trash"></i>
            </button>
          </div>
        </div>

        <!-- Nenhuma regra -->
        <div *ngIf="rules.length === 0" class="bg-white rounded-xl shadow-sm p-12 text-center">
          <i class="pi pi-cog text-5xl text-aurora-dark/20 mb-4"></i>
          <h3 class="text-lg font-medium text-aurora-dark/60">Nenhuma regra configurada</h3>
          <p class="text-sm text-aurora-dark/40 mt-1">Crie regras para automatizar seu ambiente</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .btn-primary {
      @apply bg-aurora-primary text-white px-4 py-2 rounded-lg text-sm font-medium
             hover:bg-aurora-primary/90 transition-colors flex items-center
             disabled:opacity-40 disabled:cursor-not-allowed;
    }
  `]
})
export class RulesComponent implements OnInit {
  rules: Rule[] = [];
  lights: Light[] = [];
  loading = false;
  showForm = false;
  editingRuleId: string | null = null;
  newRule: CreateRuleRequest = {
    name: '',
    trigger_type: 'light_low',
    trigger_id: 'esp32-ldr-001',
    trigger_threshold: 30,
    action_type: 'turn_on_light',
    action_id: ''
  };

  constructor(
    private rulesService: RulesService,
    private lightingService: LightingService
  ) {}

  ngOnInit(): void {
    this.loadRules();
    this.loadLights();
  }

  toggleForm(): void {
    this.showForm = !this.showForm;
    if (!this.showForm) {
      this.editingRuleId = null;
      this.resetForm();
    }
    if (this.showForm && this.lights.length === 0) this.loadLights();
  }

  loadLights(): void {
    this.lightingService.listLights().subscribe({
      next: (lights) => this.lights = lights || [],
      error: () => this.lights = []
    });
  }

  loadRules(): void {
    this.loading = true;
    this.rulesService.listRules().subscribe({
      next: (rules) => {
        this.rules = rules || [];
        this.loading = false;
      },
      error: () => {
        this.loading = false;
      }
    });
  }

  canCreate(): boolean {
    if (!this.newRule.name?.trim()) return false;
    if ((this.newRule.action_type === 'turn_on_light' || this.newRule.action_type === 'turn_off_light') && !this.newRule.action_id) return false;
    return true;
  }

  createRule(): void {
    const payload: CreateRuleRequest = {
      ...this.newRule,
      trigger_threshold:
        this.newRule.trigger_type === 'light_low' || this.newRule.trigger_type === 'light_high'
          ? (this.newRule.trigger_threshold ?? 30)
          : undefined
    };
    if (this.editingRuleId) {
      this.rulesService.updateRule(this.editingRuleId, payload).subscribe({
        next: (updated) => {
          this.rules = this.rules.map(r => r.id === updated.id ? updated : r);
          this.showForm = false;
          this.editingRuleId = null;
          this.resetForm();
        }
      });
    } else {
      this.rulesService.createRule(payload).subscribe({
        next: (rule) => {
          this.rules = [rule, ...this.rules];
          this.showForm = false;
          this.resetForm();
        }
      });
    }
  }

  startEdit(rule: Rule): void {
    this.editingRuleId = rule.id;
    this.showForm = true;
    this.newRule = {
      name: rule.name,
      trigger_type: rule.trigger_type,
      trigger_id: rule.trigger_id,
      trigger_threshold: rule.trigger_threshold,
      action_type: rule.action_type,
      action_id: rule.action_id
    };
  }

  private resetForm(): void {
    this.newRule = {
      name: '',
      trigger_type: 'light_low',
      trigger_id: 'esp32-ldr-001',
      trigger_threshold: 30,
      action_type: 'turn_on_light',
      action_id: ''
    };
  }

  deleteRule(ruleId: string): void {
    this.rulesService.deleteRule(ruleId).subscribe({
      next: () => {
        this.rules = this.rules.filter(r => r.id !== ruleId);
      }
    });
  }

  getTriggerLabel(rule: Rule): string {
    const labels: Record<string, string> = {
      motion: 'Presença detectada',
      light_low: 'Luminosidade baixa',
      light_high: 'Luminosidade alta',
      schedule: 'Agendamento'
    };
    const base = labels[rule.trigger_type] || rule.trigger_type;
    if (rule.trigger_type === 'light_low' && rule.trigger_threshold != null && rule.trigger_threshold > 0) {
      return base + ' (abaixo de ' + rule.trigger_threshold + '%)';
    }
    if (rule.trigger_type === 'light_high' && rule.trigger_threshold != null && rule.trigger_threshold > 0) {
      return base + ' (acima de ' + rule.trigger_threshold + '%)';
    }
    return base;
  }

  getActionTargetLabel(rule: Rule): string {
    if (rule.action_type !== 'turn_on_light' && rule.action_type !== 'turn_off_light') return rule.action_id || '';
    const light = this.lights.find(l => l.id === rule.action_id);
    return light ? light.name : (rule.action_id || '');
  }

  getActionLabel(type: string): string {
    const labels: Record<string, string> = {
      turn_on_light: 'Ligar luz',
      turn_off_light: 'Desligar luz',
      trigger_alarm: 'Acionar alarme'
    };
    return labels[type] || type;
  }
}
