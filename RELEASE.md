# Guía de Releases

Este proyecto utiliza GitHub Actions para crear releases automáticos multiplataforma.



```bash
# Crear tag localmente
git tag -a v1.0.0 -m "Release v1.0.0"

# Pushear el tag a GitHub
git push origin v1.0.0
```

## Versionado

Este proyecto sigue [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0): Cambios incompatibles en la API
- **MINOR** (v0.1.0): Nueva funcionalidad compatible con versiones anteriores
- **PATCH** (v0.0.1): Corrección de bugs compatible con versiones anteriores

## Plataformas soportadas

El workflow genera binarios e instaladores para:

### Windows
- `gonauta-X.X.X-windows-amd64.zip` - Windows 64-bit (Intel/AMD)

### Linux

**Paquetes:**
- `gonauta-X.X.X-amd64.deb` - Debian/Ubuntu (amd64)
- `gonauta-X.X.X-arm64.deb` - Debian/Ubuntu (ARM64)
- `gonauta-X.X.X-x86_64.rpm` - Fedora/RHEL/CentOS (amd64)
- `gonauta-X.X.X-arm64.rpm` - Fedora/RHEL/CentOS (ARM64)
- `gonauta-X.X.X-amd64.apk` - Alpine Linux (amd64)
- `gonauta-X.X.X-arm64.apk` - Alpine Linux (ARM64)

**Archivos tar.gz:**
- `gonauta-X.X.X-linux-amd64.tar.gz` - Linux 64-bit (Intel/AMD)
- `gonauta-X.X.X-linux-arm64.tar.gz` - Linux ARM64

## Instalación desde releases

### Windows
1. Descarga `gonauta-X.X.X-windows-amd64.zip` desde la página de releases
2. Extrae el archivo ZIP
3. Mueve `gonauta.exe` a una carpeta en tu PATH o ejecuta directamente

### Linux (binario)
```bash
# Descargar el binario
curl -L -o gonauta.tar.gz https://github.com/yosle/gonauta/releases/latest/download/gonauta-1.0.0-linux-amd64.tar.gz

# Extraer
tar -xzf gonauta.tar.gz

# Dar permisos de ejecución
chmod +x gonauta

# Mover a PATH (opcional)
sudo mv gonauta /usr/local/bin/
```

### Debian/Ubuntu (.deb)
```bash
wget https://github.com/yosle/gonauta/releases/latest/download/gonauta-1.0.0-amd64.deb
sudo dpkg -i gonauta-1.0.0-amd64.deb
```

### Fedora/RHEL/CentOS (.rpm)
```bash
wget https://github.com/yosle/gonauta/releases/latest/download/gonauta-1.0.0-x86_64.rpm
sudo rpm -i gonauta-1.0.0-x86_64.rpm
```

### Alpine Linux (.apk)
```bash
wget https://github.com/yosle/gonauta/releases/latest/download/gonauta-1.0.0-x86_64.apk
sudo apk add --allow-untrusted gonauta-1.0.0-x86_64.apk
```

## Trigger manual

También puedes ejecutar el workflow manualmente desde la pestaña "Actions" en GitHub sin crear un tag. Esto generará los binarios pero no creará un release público.

## Notas técnicas

- Los binarios se compilan con `CGO_ENABLED=0` para máxima portabilidad
- Se aplican flags de optimización `-ldflags="-s -w"` para reducir el tamaño
- El paquete .deb se instala en `/usr/local/bin/gonauta`
