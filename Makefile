.PHONY: run build clean deps help test test-unit test-integration

# Variables
BINARY_NAME=habittracker
MAIN_FILE=main.go

help: ## Mostrar esta ayuda
	@echo "Comandos disponibles:"
	@echo "  make run              - Ejecutar la aplicación"
	@echo "  make build            - Compilar el binario"
	@echo "  make clean            - Limpiar archivos generados"
	@echo "  make deps             - Instalar dependencias"
	@echo "  make test             - Ejecutar tests unitarios"
	@echo "  make test-integration - Ejecutar tests de integración (envía notificaciones reales)"
	@echo "  make check-webhook    - Verificar estado del webhook"

deps: ## Instalar dependencias
	@echo "📦 Instalando dependencias..."
	go mod tidy
	go mod download
	@echo "✅ Dependencias instaladas"

build: deps ## Compilar el binario
	@echo "🔨 Compilando..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Compilación exitosa: $(BINARY_NAME)"

run: deps ## Ejecutar la aplicación
	@echo "🚀 Ejecutando Habit Tracker Bot..."
	go run $(MAIN_FILE)

clean: ## Limpiar archivos generados
	@echo "🧹 Limpiando..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "✅ Limpieza completada"

test: test-unit ## Ejecutar tests unitarios (alias de test-unit)

test-unit: ## Ejecutar tests unitarios
	@echo "🧪 Ejecutando tests unitarios..."
	go test -v ./bot ./config ./habits ./scheduler
	@echo "✅ Tests unitarios completados"

test-integration: ## Ejecutar tests de integración (envía notificaciones reales)
	@echo "🚀 Ejecutando tests de integración..."
	@echo "⚠️  ADVERTENCIA: Estos tests enviarán notificaciones REALES a tu Telegram"
	@echo "⚠️  Asegúrate de haber enviado /start al bot primero"
	@echo ""
	go test -v -run TestEndToEnd
	@echo "✅ Tests de integración completados"

check-webhook: ## Verificar estado del webhook
	@echo "📡 Verificando estado del webhook..."
	go run ./cmd/check-webhook
