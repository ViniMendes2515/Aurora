import { Routes } from '@angular/router';
import { authGuard } from '@core/guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full'
  },
  {
    path: 'login',
    loadComponent: () => import('@features/auth/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'register',
    loadComponent: () => import('@features/auth/register/register.component').then(m => m.RegisterComponent)
  },
  {
    path: 'dashboard',
    loadComponent: () => import('@features/dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [authGuard],
    children: [
      {
        path: '',
        loadComponent: () => import('@features/dashboard/home/home.component').then(m => m.HomeComponent)
      },
      {
        path: 'devices',
        loadComponent: () => import('@features/dashboard/devices/devices.component').then(m => m.DevicesComponent)
      },
      {
        path: 'sensors',
        loadComponent: () => import('@features/dashboard/sensors/sensors.component').then(m => m.SensorsComponent)
      },
      {
        path: 'rules',
        loadComponent: () => import('@features/dashboard/rules/rules.component').then(m => m.RulesComponent)
      },
      {
        path: 'notifications',
        loadComponent: () => import('@features/dashboard/notifications/notifications.component').then(m => m.NotificationsComponent)
      }
    ]
  },
  {
    path: '**',
    redirectTo: 'dashboard'
  }
];
