# Habit Tracker Bot 🎯

Bot de Telegram en Go para rastrear hábitos diarios. Recibe notificaciones automáticas todos los días para revisar tus hábitos.

## Características

- ✅ Gestión de hábitos (agregar, listar, eliminar)
- 📅 Notificaciones diarias programadas
- 💾 Persistencia de datos en JSON
- 🤖 Interfaz interactiva con botones inline
- ⏰ Configuración de zona horaria

## Requisitos

- Go 1.21 o superior
- Make
- Token de bot de Telegram (obtenerlo desde [@BotFather](https://t.me/botfather))

## Instalación

1. Clona o descarga este repositorio

2. Copia el archivo de ejemplo de configuración:
```bash
cp .env.example .env
```

3. Edita el archivo `.env` y configura tu token de bot:
```bash
TELEGRAM_BOT_TOKEN=tu_token_aqui
NOTIFICATION_TIME=09:00
TIMEZONE=America/Argentina/Buenos_Aires
```

4. Instala las dependencias:
```bash
make deps
```

## Uso

### Ejecutar el bot

```bash
make run
```

### Compilar el binario

```bash
make build
./habittracker
```

### Limpiar archivos generados

```bash
make clean
```

## Comandos del Bot

Una vez que el bot esté ejecutándose, puedes interactuar con él en Telegram usando estos comandos:

- `/start` - Iniciar el bot y recibir mensaje de bienvenida
- `/help` - Mostrar lista de comandos disponibles
- `/addhabit <nombre>` - Agregar un nuevo hábito
  - Ejemplo: `/addhabit Hacer ejercicio`
- `/listhabits` - Listar todos tus hábitos configurados
- `/deletehabit <id>` - Eliminar un hábito por su ID
  - Ejemplo: `/deletehabit 1`

## Notificaciones Diarias

El bot enviará automáticamente un mensaje todos los días a la hora configurada (por defecto 9:00 AM) con todos tus hábitos. Cada hábito tendrá botones para marcar si lo completaste (✅) o no (❌).

Las respuestas se guardan automáticamente en `data/responses.json`.

## Estructura del Proyecto

```
HabitTracker/
├── bot/              # Cliente de Telegram
│   └── bot.go
├── config/           # Gestión de configuración
│   └── config.go
├── habits/           # Lógica de hábitos
│   └── habit.go
├── scheduler/        # Programador de tareas
│   └── scheduler.go
├── data/             # Almacenamiento de datos
│   ├── habits.json
│   └── responses.json
├── main.go           # Punto de entrada
├── Makefile          # Comandos de gestión
├── .env              # Configuración (no versionado)
└── .env.example      # Plantilla de configuración
```

## Almacenamiento de Datos

Los datos se almacenan en archivos JSON en el directorio `data/`:

- `habits.json` - Lista de hábitos configurados
- `responses.json` - Historial de respuestas diarias

Estos archivos se crean automáticamente la primera vez que ejecutas el bot.

## Testing

El proyecto incluye tests unitarios y de integración.

### Tests Unitarios

Ejecuta los tests unitarios que verifican la funcionalidad del scheduler sin enviar notificaciones:

```bash
make test
# o
make test-unit
```

Estos tests verifican:
- Programación correcta de jobs
- Manejo de zonas horarias
- Validación de formatos de tiempo
- Múltiples jobs programados
- Manejo de errores

### Tests de Integración

⚠️ **ADVERTENCIA**: Los tests de integración envían notificaciones REALES a tu cuenta de Telegram.

Antes de ejecutar estos tests:

1. **Obtén tu Chat ID de Telegram**:
   - Inicia el bot con `make run`
   - Envía `/start` al bot en Telegram
   - Revisa los logs del bot, verás algo como: `User chat ID saved: 123456789`
   - Copia ese número

2. **Configura el Chat ID en `.env`**:
   ```bash
   TELEGRAM_CHAT_ID=123456789
   ```

3. **Ejecuta los tests**:
   ```bash
   make test-integration
   ```

Estos tests incluyen:
- **TestEndToEndNotification**: Programa un job que se ejecuta en 5 segundos y envía una notificación real
- **TestEndToEndWithManualTrigger**: Envía una notificación inmediatamente sin esperar al scheduler

Deberías recibir notificaciones de prueba en tu Telegram con hábitos de ejemplo.

## Desarrollo

### Agregar nuevos comandos

Edita `bot/bot.go` y agrega un nuevo case en el switch de `handleMessage()`.

### Modificar la hora de notificaciones

Edita el archivo `.env` y cambia el valor de `NOTIFICATION_TIME` (formato 24 horas HH:MM).

### Cambiar zona horaria

Edita el archivo `.env` y cambia el valor de `TIMEZONE` usando el formato de la base de datos de zonas horarias de IANA (ej: `America/Argentina/Buenos_Aires`).

## Troubleshooting

### El bot no responde
- Verifica que el token en `.env` sea correcto
- Asegúrate de que el bot esté ejecutándose (`make run`)
- Revisa los logs en la consola

### Las notificaciones no llegan
- Primero envía un mensaje al bot (ej: `/start`) para que registre tu chat ID
- Verifica la configuración de `NOTIFICATION_TIME` en `.env`
- Revisa los logs para ver si hay errores

### Error al compilar
- Asegúrate de tener Go instalado: `go version`
- Ejecuta `make deps` para instalar las dependencias

## Licencia

MIT
