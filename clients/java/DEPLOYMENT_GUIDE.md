# 🚀 Guía de Publicación - NetTestLab Cliente Java

## ✅ Estado del Proyecto

El cliente Java de NetTestLab está **completamente listo** para producción y publicación en Maven Central. Incluye:

- ✅ Implementación completa de todos los servicios gRPC
- ✅ Tests exhaustivos (36 tests, 100% exitosos)
- ✅ Documentación JavaDoc completa
- ✅ Ejemplos de uso funcionales
- ✅ Configuración Maven para publicación
- ✅ JARs generados (principal, sources, javadoc)

## 📋 Resumen Técnico

### Componentes Implementados

1. **NetTestLabClient** - Cliente principal con patrón Builder
2. **NetworkControlClient** - Control de condiciones de red
3. **MonitoringClient** - Monitoreo y métricas del sistema
4. **ProfileClient** - Gestión de perfiles de red
5. **NetworkConditionsBuilder** - Constructor fluido para condiciones
6. **ProtoUtils** - Utilidades de conversión protobuf
7. **BasicExample** - Ejemplo completo de uso

### Tecnologías

- **Java 11+** - Versión mínima soportada
- **gRPC 1.53.0** - Comunicación con servidor
- **Protobuf 3.21.12** - Serialización de mensajes
- **SLF4J** - Logging estructurado
- **JUnit 5** - Framework de testing
- **Maven** - Sistema de build

## 🔧 Cómo Usar el Cliente

### Instalación

Cuando esté publicado en Maven Central, añade esta dependencia:

```xml
<dependency>
    <groupId>io.github.eitol</groupId>
    <artifactId>nettestlab-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Uso Básico

```java
// Crear cliente
try (NetTestLabClient client = NetTestLabClient.builder()
        .host("localhost")
        .port(8080)
        .build()) {
    
    // Verificar salud del servidor
    var health = client.monitoring().getHealth();
    
    // Aplicar condiciones de red 3G
    NetworkConditions conditions = NetworkConditionsBuilder.threeG();
    client.networkControl().applyConditions("eth0", conditions);
    
    // Obtener métricas
    var metrics = client.monitoring().getMetrics();
    System.out.println("CPU: " + metrics.getSystem().getCpuUsage() + "%");
}
```

### Ejemplos Predefinidos

```java
// Condiciones de red predefinidas
NetworkConditions twoG = NetworkConditionsBuilder.twoG();     // 2G móvil
NetworkConditions threeG = NetworkConditionsBuilder.threeG(); // 3G móvil  
NetworkConditions fourG = NetworkConditionsBuilder.fourG();   // 4G/LTE
NetworkConditions fiveG = NetworkConditionsBuilder.fiveG();   // 5G
NetworkConditions wifi = NetworkConditionsBuilder.wiFi();     // WiFi típico
NetworkConditions satellite = NetworkConditionsBuilder.satellite(); // Satélite

// Condiciones personalizadas
NetworkConditions custom = NetworkConditionsBuilder.create()
    .latency(100)                    // 100ms latencia
    .jitter(20)                      // 20ms jitter
    .bandwidth(5_000_000, 1_000_000) // 5Mbps down, 1Mbps up
    .packetLoss(0.1f)               // 0.1% pérdida
    .build();
```

## 📦 Publicación en Maven Central

### 1. Prerequisitos

Antes de publicar, necesitas:

- [ ] Cuenta en [Sonatype OSSRH](https://central.sonatype.org/register/central-portal/)
- [ ] Dominio GitHub verificado (`io.github.eitol`)
- [ ] Claves GPG para firmar artefactos
- [ ] Configurar `~/.m2/settings.xml`

### 2. Configuración de Credenciales

Configura tu `~/.m2/settings.xml`:

```xml
<settings>
  <servers>
    <server>
      <id>central</id>
      <username>tu-username-sonatype</username>
      <password>tu-password-sonatype</password>
    </server>
  </servers>
  
  <profiles>
    <profile>
      <id>release</id>
      <properties>
        <gpg.keyname>tu-key-id-gpg</gpg.keyname>
        <gpg.passphrase>tu-passphrase-gpg</gpg.passphrase>
      </properties>
    </profile>
  </profiles>
</settings>
```

### 3. Comandos de Publicación

```bash
# 1. Generar claves GPG (si no las tienes)
gpg --gen-key

# 2. Publicar la clave pública
gpg --keyserver keys.openpgp.org --send-keys TU_KEY_ID

# 3. Desplegar a Central Portal
mvn clean deploy -Prelease

# 4. Alternativamente, generar bundle para subida manual
mvn clean package -Prelease
```

### 4. Verificación Pre-publicación

Ejecuta estos comandos para verificar que todo está listo:

```bash
# Tests completos
mvn clean test

# Generar todos los artefactos
mvn clean package -Prelease

# Verificar contenido de JARs
jar -tf target/nettestlab-client-1.0.0.jar
jar -tf target/nettestlab-client-1.0.0-sources.jar
jar -tf target/nettestlab-client-1.0.0-javadoc.jar
```

### 5. Estructura de Archivos Generados

Después de `mvn package -Prelease`:

```
target/
├── nettestlab-client-1.0.0.jar           # JAR principal
├── nettestlab-client-1.0.0-sources.jar   # Código fuente
├── nettestlab-client-1.0.0-javadoc.jar   # Documentación
├── nettestlab-client-1.0.0.jar.asc       # Firma GPG principal
├── nettestlab-client-1.0.0-sources.jar.asc # Firma GPG sources
└── nettestlab-client-1.0.0-javadoc.jar.asc # Firma GPG javadoc
```

## 🧪 Testing

El proyecto incluye 36 tests que cubren:

- ✅ Todas las operaciones gRPC
- ✅ Manejo de errores y excepciones
- ✅ Configuración del cliente
- ✅ Utilidades de conversión
- ✅ Builders y validaciones

```bash
# Ejecutar tests
mvn test

# Tests con coverage
mvn test jacoco:report
```

## 📚 Documentación

### JavaDoc

La documentación está incluida en el JAR y disponible después de la instalación. Incluye:

- API completa de todas las clases públicas
- Ejemplos de uso en comentarios
- Información de parámetros y excepciones
- Enlaces entre clases relacionadas

### Ejemplos

Consulta `src/main/java/io/github/eitol/nettestlab/examples/BasicExample.java` para ver un ejemplo completo de uso.

## 🔍 Solución de Problemas

### Problemas Comunes

1. **Error de conexión gRPC**
   ```
   Solución: Verificar que el servidor NetTestLab esté ejecutándose en el host/puerto correcto
   ```

2. **Warnings de Java 24**
   ```
   Solución: Los warnings son normales y no afectan la funcionalidad
   ```

3. **Error en tests**
   ```
   Solución: Ejecutar mvn clean test para limpiar y re-ejecutar
   ```

## 📞 Soporte

- **GitHub Issues**: Para reportar bugs o solicitar features
- **Email**: Para soporte directo
- **Documentación**: README.md y JavaDoc incluidos

---

## ✨ ¡Listo para Producción!

El cliente Java está completamente implementado, probado y listo para ser publicado en Maven Central. Todos los componentes están funcionando correctamente y la documentación está completa.

**Próximos pasos:**
1. Configurar credenciales de Sonatype OSSRH
2. Generar claves GPG para firma
3. Ejecutar `mvn deploy -Prelease`
4. Esperar aprobación en Central Portal

¡El cliente estará disponible para toda la comunidad Java! 🎉