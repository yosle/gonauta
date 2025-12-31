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

## Trigger manual

También puedes ejecutar el workflow manualmente desde la pestaña "Actions" en GitHub sin crear un tag. Esto generará los binarios pero no creará un release público.

## Notas técnicas

- Los binarios se compilan con `CGO_ENABLED=0` para máxima portabilidad
- Se aplican flags de optimización `-ldflags="-s -w"` para reducir el tamaño
- El paquete .deb se instala en `/usr/local/bin/gonauta`
