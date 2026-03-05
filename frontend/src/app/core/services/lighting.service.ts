import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@environments/environment';

export interface Light {
  id: string;
  name: string;
  location: string;
  state: 'on' | 'off';
}

@Injectable({ providedIn: 'root' })
export class LightingService {
  private readonly baseUrl = environment.lightingApiUrl;

  constructor(private http: HttpClient) {}

  listLights(): Observable<Light[]> {
    return this.http.get<Light[]>(`${this.baseUrl}/lights`);
  }

  turnOn(lightId: string): Observable<Light> {
    return this.http.post<Light>(`${this.baseUrl}/lights/${lightId}/on`, {});
  }

  turnOff(lightId: string): Observable<Light> {
    return this.http.post<Light>(`${this.baseUrl}/lights/${lightId}/off`, {});
  }

  getStatus(lightId: string): Observable<Light> {
    return this.http.get<Light>(`${this.baseUrl}/lights/${lightId}/status`);
  }
}
