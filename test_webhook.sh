#!/bin/bash

# Script para probar la configuración del webhook

echo "🧪 Testing Webhook Configuration..."
echo ""

# Verificar que cloudflared está corriendo
if ! pgrep -f "cloudflared" > /dev/null; then
    echo "❌ Cloudflared tunnel is not running!"
    echo "   Start it with: cloudflared tunnel --url http://localhost:8080"
    exit 1
fi

echo "✅ Cloudflared tunnel is running"
echo ""

# Leer la URL del webhook desde .env
if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    echo "   Copy .env.example to .env and configure it"
    exit 1
fi

WEBHOOK_URL=$(grep "^WEBHOOK_URL=" .env | cut -d'=' -f2)

if [ -z "$WEBHOOK_URL" ]; then
    echo "⚠️  WEBHOOK_URL not set in .env"
    echo "   The bot will run in POLLING mode instead"
    echo ""
else
    echo "🔗 . URL: $WEBHOOK_URL"
    echo ""
fi

# Ejecutar el bot
echo "🚀 Starting bot..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

go run main.go
