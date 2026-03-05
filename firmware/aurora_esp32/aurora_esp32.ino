#include <WiFi.h>
#include <WebServer.h>
#include <HTTPClient.h>
#include "config.h"

// ============================================================
// Estado global
// ============================================================
WebServer server(80);

bool ledState[4]  = {false, false, false, false};
bool buzzerState  = false;

const int LED_PINS[4] = {PIN_LED_1, PIN_LED_2, PIN_LED_3, PIN_LED_4};

volatile bool pirTriggered = false;
unsigned long lastPirTrigger = 0;
unsigned long lastLdrRead    = 0;

// ============================================================
// ISR do PIR — chamada quando o pino OUT muda para HIGH
// ============================================================
void IRAM_ATTR onPirDetected() {
  unsigned long now = millis();
  if (now - lastPirTrigger > PIR_COOLDOWN_MS) {
    pirTriggered  = true;
    lastPirTrigger = now;
  }
}

// ============================================================
// Helpers HTTP de saída (ESP32 → microserviços)
// ============================================================
void postDeviceMotion() {
  if (WiFi.status() != WL_CONNECTED) return;

  HTTPClient http;
  String url = String(SENSORS_SERVICE_URL) + "/sensors/device/" + SENSOR_PIR_ID + "/motion";

  http.begin(url);
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  int code = http.POST("{}");
  if (code > 0) {
    Serial.printf("[PIR] Motion event sent → HTTP %d\n", code);
  } else {
    Serial.printf("[PIR] Motion event FAILED: %s\n", http.errorToString(code).c_str());
  }
  http.end();
}

void postLdrReading(int rawValue) {
  if (WiFi.status() != WL_CONNECTED) return;

  float percent = map(rawValue, 0, 4095, 100, 0);  // 0 = escuro, 100 = claro

  HTTPClient http;
  String url = String(SENSORS_SERVICE_URL) + "/sensors/device/" + SENSOR_LDR_ID + "/light";

  http.begin(url);
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  String body = "{\"value\":" + String(percent, 1) + ",\"raw\":" + String(rawValue) + "}";
  int code = http.POST(body);
  if (code > 0) {
    Serial.printf("[LDR] Light level %.1f%% sent → HTTP %d\n", percent, code);
  } else {
    Serial.printf("[LDR] Light level FAILED: %s\n", http.errorToString(code).c_str());
  }
  http.end();
}

// ============================================================
// Handlers HTTP de entrada (microserviços → ESP32)
// ============================================================
void handleHealth() {
  String body = "{\"status\":\"ok\",\"leds\":[";
  for (int i = 0; i < 4; i++) {
    body += ledState[i] ? "true" : "false";
    if (i < 3) body += ",";
  }
  body += "],\"buzzer\":";
  body += buzzerState ? "true" : "false";
  body += ",\"ip\":\"" + WiFi.localIP().toString() + "\"}";
  server.send(200, "application/json", body);
}

// Handlers genéricos por índice (0-based internamente, 1-based na URL)
void handleLedOn(int idx) {
  ledState[idx] = true;
  digitalWrite(LED_PINS[idx], HIGH);
  server.send(200, "application/json", "{\"status\":\"on\"}");
  Serial.printf("[LED %d] Ligado via HTTP\n", idx + 1);
}

void handleLedOff(int idx) {
  ledState[idx] = false;
  digitalWrite(LED_PINS[idx], LOW);
  server.send(200, "application/json", "{\"status\":\"off\"}");
  Serial.printf("[LED %d] Desligado via HTTP\n", idx + 1);
}

void handleLedStatus(int idx) {
  String body = "{\"status\":\"";
  body += ledState[idx] ? "on" : "off";
  body += "\"}";
  server.send(200, "application/json", body);
}

// Wrappers por LED para o WebServer
void handleLed1On()     { handleLedOn(0); }
void handleLed1Off()    { handleLedOff(0); }
void handleLed1Status() { handleLedStatus(0); }
void handleLed2On()     { handleLedOn(1); }
void handleLed2Off()    { handleLedOff(1); }
void handleLed2Status() { handleLedStatus(1); }
void handleLed3On()     { handleLedOn(2); }
void handleLed3Off()    { handleLedOff(2); }
void handleLed3Status() { handleLedStatus(2); }
void handleLed4On()     { handleLedOn(3); }
void handleLed4Off()    { handleLedOff(3); }
void handleLed4Status() { handleLedStatus(3); }

void handleBuzzerOn() {
  buzzerState = true;
  digitalWrite(PIN_BUZZER, HIGH);
  server.send(200, "application/json", "{\"status\":\"on\"}");
  Serial.println("[BUZZER] Acionado via HTTP");
}

void handleBuzzerOff() {
  buzzerState = false;
  digitalWrite(PIN_BUZZER, LOW);
  server.send(200, "application/json", "{\"status\":\"off\"}");
  Serial.println("[BUZZER] Desligado via HTTP");
}

void handleBuzzerStatus() {
  String body = "{\"status\":\"";
  body += buzzerState ? "on" : "off";
  body += "\"}";
  server.send(200, "application/json", body);
}

void handleNotFound() {
  server.send(404, "application/json", "{\"error\":\"not found\"}");
}

// ============================================================
// Registro do ESP32 no lighting-service
// ============================================================
void registerDevice() {
  if (WiFi.status() != WL_CONNECTED) return;

  HTTPClient http;
  String url = String(LIGHTING_SERVICE_URL) + "/devices/register";

  http.begin(url);
  http.addHeader("Content-Type", "application/json");
  http.addHeader("X-Device-Key", DEVICE_API_KEY);

  String ip = WiFi.localIP().toString();
  String body = "{\"device_id\":\"esp32-main\",\"ip\":\"" + ip + "\",\"port\":80}";

  int code = http.POST(body);
  if (code > 0) {
    Serial.printf("[INIT] Dispositivo registrado no lighting-service → HTTP %d\n", code);
  } else {
    Serial.printf("[INIT] Registro no lighting-service falhou: %s\n", http.errorToString(code).c_str());
  }
  http.end();
}

// ============================================================
// Setup
// ============================================================
void setup() {
  Serial.begin(115200);
  delay(500);

  // Configurar pinos
  for (int i = 0; i < 4; i++) {
    pinMode(LED_PINS[i], OUTPUT);
    digitalWrite(LED_PINS[i], LOW);
  }
  pinMode(PIN_BUZZER, OUTPUT);
  pinMode(PIN_PIR,    INPUT);
  // PIN_LDR é analógico — não precisa de pinMode

  digitalWrite(PIN_BUZZER, LOW);

  // Conectar ao Wi-Fi
  Serial.printf("\n[INIT] Conectando em %s", WIFI_SSID);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println();
  Serial.printf("[INIT] Wi-Fi conectado! IP: %s\n", WiFi.localIP().toString().c_str());

  // Configurar interrupção do PIR
  attachInterrupt(digitalPinToInterrupt(PIN_PIR), onPirDetected, RISING);
  Serial.printf("[INIT] PIR ativo no GPIO %d\n", PIN_PIR);

  // Registrar rotas HTTP
  server.on("/health",          HTTP_GET, handleHealth);
  server.on("/led/1/on",        HTTP_GET, handleLed1On);
  server.on("/led/1/off",       HTTP_GET, handleLed1Off);
  server.on("/led/1/status",    HTTP_GET, handleLed1Status);
  server.on("/led/2/on",        HTTP_GET, handleLed2On);
  server.on("/led/2/off",       HTTP_GET, handleLed2Off);
  server.on("/led/2/status",    HTTP_GET, handleLed2Status);
  server.on("/led/3/on",        HTTP_GET, handleLed3On);
  server.on("/led/3/off",       HTTP_GET, handleLed3Off);
  server.on("/led/3/status",    HTTP_GET, handleLed3Status);
  server.on("/led/4/on",        HTTP_GET, handleLed4On);
  server.on("/led/4/off",       HTTP_GET, handleLed4Off);
  server.on("/led/4/status",    HTTP_GET, handleLed4Status);
  server.on("/buzzer/on",       HTTP_GET, handleBuzzerOn);
  server.on("/buzzer/off",      HTTP_GET, handleBuzzerOff);
  server.on("/buzzer/status",   HTTP_GET, handleBuzzerStatus);
  server.onNotFound(handleNotFound);

  server.begin();
  Serial.println("[INIT] Servidor HTTP iniciado na porta 80");

  // Registrar dispositivo nos serviços
  registerDevice();

  Serial.println("[INIT] Aurora ESP32 pronto!");
  Serial.println("  GET /led/1/on      → liga LED 1 (GPIO 23)");
  Serial.println("  GET /led/1/off     → desliga LED 1");
  Serial.println("  GET /led/2/on      → liga LED 2 (GPIO 22)");
  Serial.println("  GET /led/2/off     → desliga LED 2");
  Serial.println("  GET /led/3/on      → liga LED 3 (GPIO 21)");
  Serial.println("  GET /led/3/off     → desliga LED 3");
  Serial.println("  GET /buzzer/on     → liga buzzer");
  Serial.println("  GET /buzzer/off    → desliga buzzer");
  Serial.println("  GET /health        → status do dispositivo");
}

// ============================================================
// Loop principal
// ============================================================
void loop() {
  // Processar requisições HTTP recebidas
  server.handleClient();

  // Processar evento do PIR (fora da ISR para poder usar HTTPClient)
  if (pirTriggered) {
    pirTriggered = false;
    Serial.println("[PIR] Movimento detectado!");
    postDeviceMotion();
  }

  // Leitura periódica do LDR
  unsigned long now = millis();
  if (now - lastLdrRead >= LDR_READ_INTERVAL_MS) {
    lastLdrRead = now;
    int raw = analogRead(PIN_LDR);
    Serial.printf("[LDR] Leitura bruta: %d\n", raw);
    postLdrReading(raw);
  }
}
