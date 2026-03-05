import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface Rule {
  id: string;
  name: string;
  trigger_type: string;
  trigger_id: string;
  trigger_threshold?: number;
  action_type: string;
  action_id: string;
  enabled: boolean;
  created_at: string;
}

export interface CreateRuleRequest {
  name: string;
  trigger_type: string;
  trigger_id: string;
  trigger_threshold?: number;
  action_type: string;
  action_id: string;
}

@Injectable({ providedIn: 'root' })
export class RulesService {
  private readonly baseUrl = environment.rulesApiUrl;

  constructor(private http: HttpClient) {}

  listRules(): Observable<Rule[]> {
    return this.http.get<Rule[]>(`${this.baseUrl}/rules`);
  }

  createRule(req: CreateRuleRequest): Observable<Rule> {
    return this.http.post<Rule>(`${this.baseUrl}/rules`, req);
  }

  updateRule(ruleId: string, req: CreateRuleRequest): Observable<Rule> {
    return this.http.put<Rule>(`${this.baseUrl}/rules/${ruleId}`, req);
  }

  deleteRule(ruleId: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/rules/${ruleId}`);
  }
}
