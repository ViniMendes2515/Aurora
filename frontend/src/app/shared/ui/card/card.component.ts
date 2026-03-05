import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './card.component.html',
  styleUrl: './card.component.css'
})
export class CardComponent {
  @Input() title: string = '';
  @Input() value: string | number = '';
  @Input() icon: string = 'pi-chart-bar';
  @Input() variant: 'primary' | 'secondary' | 'base' | 'dark' = 'primary';
}
