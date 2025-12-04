#!/bin/bash

# Aurora Home System - Test Script
# Este script auxilia nos testes do sistema

BASE_URL_AUTH="http://localhost:8080"
BASE_URL_SENSORS="http://localhost:8081"

echo "=== Aurora Home System - Test Script ==="
echo ""

# Registrar usuário
echo "1. Registrando usuário..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL_AUTH/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@aurora.com", "password": "senha123"}')

echo "Response: $REGISTER_RESPONSE"
echo ""

# Login
echo "2. Fazendo login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL_AUTH/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@aurora.com", "password": "senha123"}')

echo "Response: $LOGIN_RESPONSE"

# Extrair token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Erro: Não foi possível obter o token"
  exit 1
fi

echo ""
echo "Token obtido: ${TOKEN:0:50}..."
echo ""

# Testar sensor
echo "3. Registrando movimento no sensor..."
SENSOR_RESPONSE=$(curl -s -X POST "$BASE_URL_SENSORS/sensors/1/motion" \
  -H "Authorization: Bearer $TOKEN")

echo "Response: $SENSOR_RESPONSE"
echo ""

echo "=== Teste concluído ==="
