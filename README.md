# GoNauta

Cliente CLI en Go para gestionar conexiones a Nauta (Cuba).

## Características

- 🔐 Almacenamiento cifrado de credenciales localmente
- 🚀 Comandos simples para conectar y desconectar
- ⏱️ Consulta de tiempo restante en tiempo real
- 📊 Información detallada de la cuenta

## Instalación

### Desde releases (recomendado)

Descarga el binario precompilado para tu plataforma desde la [página de releases](https://github.com/yosle/gonauta/releases/latest):

**Windows:**
```powershell
# Descarga gonauta-windows-amd64.exe y renombra a gonauta.exe
# Ejecuta directamente o añade a tu PATH
```

**macOS:**
```bash
# Intel
curl -L -o gonauta https://github.com/yosle/gonauta/releases/latest/download/gonauta-macos-amd64

# Apple Silicon (M1/M2/M3)
curl -L -o gonauta https://github.com/yosle/gonauta/releases/latest/download/gonauta-macos-arm64

chmod +x gonauta
sudo mv gonauta /usr/local/bin/
```

**Linux:**
```bash
# Binario
curl -L -o gonauta https://github.com/yosle/gonauta/releases/latest/download/gonauta-linux-amd64
chmod +x gonauta
sudo mv gonauta /usr/local/bin/

# O instalar paquete .deb (Debian/Ubuntu)
wget https://github.com/yosle/gonauta/releases/latest/download/gonauta-1.0.0-amd64.deb
sudo dpkg -i gonauta-1.0.0-amd64.deb
```

### Desde código fuente

```bash
git clone https://github.com/yosle/gonauta.git
cd gonauta
go build
```

Esto generará el ejecutable `gonauta.exe` (Windows) o `gonauta` (Linux/Mac).

## Uso

### 1. Guardar credenciales

Primero, guarda tus credenciales de Nauta de forma segura:

```bash
go_nauta login
```

Se te pedirá tu usuario (ej: `usuario@nauta.com.cu`) y contraseña. Las credenciales se guardan cifradas en `~/.gonauta/credentials.enc`.

#### Configuración con VPN (Opcional)

Si deseas que GoNauta gestione automáticamente tu conexión VPN, usa el flag `--vpn`:

```bash
go_nauta login --vpn
```

Se te pedirá:
- Usuario y contraseña de Nauta
- **Comando de conexión VPN**: El comando completo para conectar tu VPN (ej: `C:\Program Files\NordVPN\nordvpn.exe -c`)
- **Comando de desconexión VPN**: El comando completo para desconectar tu VPN (ej: `C:\Program Files\NordVPN\nordvpn.exe -d`)

**Ejemplos de comandos VPN con NordVPN:**

**Windows:**
```bash
# Conexión
C:\Program Files\NordVPN\nordvpn.exe -c

# Desconexión
C:\Program Files\NordVPN\nordvpn.exe -d
```

**Linux/macOS:**
```bash
# Conexión
nordvpn connect

# Desconexión
nordvpn disconnect
```

**Comportamiento automático:**
- Al ejecutar `go_nauta connect`, después de conectar a Nauta, se ejecutará automáticamente el comando de conexión VPN
- Al ejecutar `go_nauta logout`, antes de cerrar sesión (con un delay de 2 segundos), se ejecutará automáticamente el comando de desconexión VPN

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
| `login [--vpn]` | Guardar credenciales (usuario y contraseña). Con `--vpn` configura comandos VPN |
| `connect` | Iniciar sesión en Nauta (ejecuta VPN automáticamente si está configurado) |
| `logout` | Cerrar sesión activa (desconecta VPN automáticamente si está configurado) |
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

### Sin VPN

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

### Con VPN (NordVPN)

```bash
# 1. Guardar credenciales con configuración VPN
go_nauta login --vpn
# Usuario: usuario@nauta.com.cu
# Contraseña: ********
# 
# --- Configuración de VPN ---
# Comando de conexión VPN: C:\Program Files\NordVPN\nordvpn.exe -c
# Comando de desconexión VPN: C:\Program Files\NordVPN\nordvpn.exe -d
# ✓ Credenciales guardadas exitosamente
# ✓ Configuración de VPN guardada

# 2. Conectar (automáticamente conecta VPN después)
go_nauta connect
# Conectando a Nauta...
# ✓ Sesión iniciada exitosamente
#   Usuario: usuario@nauta.com.cu
# 
# Conectando VPN...
# ✓ VPN conectado

# 3. Ver tiempo
go_nauta status
# ⏱  Tiempo restante: 02:30:45

# 4. Cerrar sesión (automáticamente desconecta VPN antes)
go_nauta logout
# Desconectando VPN...
# ✓ VPN desconectado
# Cerrando sesión...
# ✓ Sesión cerrada exitosamente
```

## Licencia

MIT
