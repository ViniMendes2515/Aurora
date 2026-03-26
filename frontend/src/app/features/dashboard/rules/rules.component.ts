import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RulesService, Rule, CreateRuleRequest } from '@core/services/rules.service';
import { LightingService, Light } from '@core/services/lighting.service';

@Component({
  selector: 'app-rules',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './rules.component.html'
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

  toggleRule(rule: Rule): void {
    this.rulesService.toggleRule(rule.id).subscribe({
      next: (updated) => {
        this.rules = this.rules.map(r => r.id === updated.id ? updated : r);
      }
    });
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
