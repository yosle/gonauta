# GoNauta

Cliente CLI en Go para gestionar conexiones a Nauta (Cuba).

## Características

- 🔐 Almacenamiento cifrado de credenciales localmente
- 🚀 Comandos simples para conectar y desconectar
- ⏱️ Consulta de tiempo restante en tiempo real
- 📊 Información detallada de la cuenta

## Instalación

```bash
go build
```

Esto generará el ejecutable `go_nauta.exe` (Windows) o `go_nauta` (Linux/Mac).

## Uso

### 1. Guardar credenciales

Primero, guarda tus credenciales de Nauta de forma segura:

```bash
go_nauta login
```

Se te pedirá tu usuario (ej: `usuario@nauta.com.cu`) y contraseña. Las credenciales se guardan cifradas en `~/.gonauta/credentials.enc`.

### 2. Conectar a Nauta

Para iniciar sesión en Nauta:

```bash
go_nauta connect
```

### 3. Ver tiempo restante

Consulta cuánto tiempo te queda en la sesión activa:

```bash
go_nauta status
```

### 4. Ver información completa

Obtén información detallada de tu cuenta (créditos, fecha de expiración, etc.):

```bash
go_nauta info
```

### 5. Cerrar sesión

Cuando termines, cierra la sesión:

```bash
go_nauta logout
```

## Comandos disponibles

| Comando | Descripción |
|---------|-------------|
| `login` | Guardar credenciales (usuario y contraseña) |
| `connect` | Iniciar sesión en Nauta |
| `logout` | Cerrar sesión activa |
| `status` | Ver tiempo restante de la sesión activa |
| `info` | Ver información completa del usuario |
| `help` | Mostrar ayuda |

## Seguridad

- Las credenciales se almacenan cifradas usando AES-256-GCM
- La clave de cifrado se deriva del hostname de la máquina
- Los archivos de configuración se guardan con permisos restrictivos (0600)
- La sesión activa se guarda localmente para permitir comandos rápidos

## Estructura de archivos

```
~/.gonauta/
├── credentials.enc  # Credenciales cifradas
└── session.json     # Sesión activa (temporal)
```

## Dependencias

- `github.com/PuerkitoBio/goquery` - Parsing HTML
- `golang.org/x/term` - Lectura segura de contraseñas
- `golang.org/x/net` - Networking

## Desarrollo

### Estructura del proyecto

- `main.go` - CLI y comandos principales
- `nauta.go` - Cliente y lógica de Nauta
- `config.go` - Gestión de credenciales cifradas
- `session_store.go` - Gestión de sesiones activas

### Compilar

```bash
go build
```

### Ejecutar sin compilar

```bash
go run . <comando>
```

## Ejemplo de uso completo

```bash
# 1. Guardar credenciales
go_nauta login
# Usuario: usuario@nauta.com.cu
# Contraseña: ********

# 2. Conectar
go_nauta connect
# ✓ Sesión iniciada exitosamente

# 3. Ver tiempo
go_nauta status
# ⏱  Tiempo restante: 02:30:45

# 4. Ver información
go_nauta info
# === Información del Usuario ===
# Estado: activo
# Créditos: 125.50 CUP
# ...

# 5. Cerrar sesión
go_nauta logout
# ✓ Sesión cerrada exitosamente
```

## Licencia

MIT
