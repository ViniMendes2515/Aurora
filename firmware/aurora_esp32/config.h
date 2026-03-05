#pragma once

#define WIFI_SSID       "DEBORA_MENDES"
#define WIFI_PASSWORD   "1QAZ2wsx12!"

// ============================================================
// Endereço IP do computador onde os serviços estão rodando
// Execute "ip route get 1" ou "hostname -I" para descobrir
// ============================================================
#define SENSORS_SERVICE_URL   "http://192.168.0.160:8081"
#define LIGHTING_SERVICE_URL  "http://192.168.0.160:8082"

#define DEVICE_API_KEY  "aurora-device-key-2024"

#define SENSOR_PIR_ID  "esp32-pir-001"
#define SENSOR_LDR_ID  "esp32-ldr-001"

// ============================================================
// Pinos GPIO
// ============================================================
#define PIN_LED_1   23   // LED 1 — Sala de Estar (resistor em série)
#define PIN_LED_2   22   // LED 2 — Corredor (resistor em série)
#define PIN_LED_3   21   // LED 3 — Garagem (resistor em série)
#define PIN_LED_4   18   // LED 4 — Banheiro (resistor em série)
#define PIN_PIR     14   // Saída OUT do HC-SR501
#define PIN_BUZZER  27   // Pino + do buzzer ativo
#define PIN_LDR     34   // Divisor de tensão com LDR (ADC)

// ============================================================
// Configurações de comportamento
// ============================================================
#define LDR_READ_INTERVAL_MS   2000   // Leitura do LDR a cada 2 segundos
#define PIR_COOLDOWN_MS        3000   // Ignora novos triggers por 3s após detectar
#define BUZZER_DURATION_MS     3000   // Duração do alarme ao detectar movimento
