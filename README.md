# Secure Payment Service

Este es un servicio backend para gestionar transferencias y consulta de saldos, construido en Go, Gin y GORM. Expone una API RESTful e incluye un endpoint de métricas para Prometheus.

## Configuración del Servicio
El servicio carga su configuración a través de variables de entorno definidas en internal/config/config.go. 

Las variables clave son:
- DATABASE_URL: La cadena de conexión a la base de datos PostgreSQL.
- ADDRESS: La dirección y puerto en los que el servidor escuchará (ej. :8080).

Para tests unitarios, el servicio utiliza SQLite en memoria por defecto, lo que hace los tests rápidos y autónomos.

Para desarrollo local, el servicio utiliza PostgreSQL. La configuración de la base de datos y el puerto para el entorno de Docker Compose ya están definidos directamente en el docker-compose.yml para el servicio app.

## Pasos para ejecutar el proyecto:

1. Clonar el repositorio
```
git clone git@github.com:javiacuna/secure-payment-service.git
cd secure-payment-service
```

2. Ejecutar el Servicio con Docker Compose
```
docker-compose up --build
```

## Endpoints de la API:
El servicio expone los siguientes endpoints (asumiendo que se ejecuta en http://localhost:8080):

- POST /transfer: Crea una nueva transferencia.

```
curl --location 'http://localhost:8080/api/v1/transfer' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer TOKEN' \
--data '{
    "source_account_id": "acc-001",
    "destination_account_id": "acc-002",
    "amount": 100.50,
    "currency": "USD"
}'
```

- GET /transfer/:id: Obtiene detalles de una transferencia.

```
curl --location 'http://localhost:8080/api/v1/transfer/7538b6f4-dfed-40e0-b08f-931feaf1ae3b' \
--header 'Authorization: Bearer TOKEN'
```

- GET /account/:id/balance: Consulta el saldo de una cuenta.

```
curl --location 'http://localhost:8080/api/v1/account/acc-002/balance' \
--header 'Authorization: Bearer TOKEN'
```

- POST /webhook: Actualiza el estado de una transferencia (vía webhook).

```
curl --location 'http://localhost:8080/api/v1/webhook' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer TOKEN' \
--data '{
    "transfer_id": "7538b6f4-dfed-40e0-b08f-931feaf1ae3b",
    "status": "COMPLETED"
}'
```

- GET /metrics: Expone métricas para Prometheus.

```
curl --location 'http://localhost:8080/metrics'
```

## 🔐 Autenticación (JWT)

Este servicio requiere autenticación mediante tokens JWT para acceder a sus endpoints seguros.

Para generar un token JWT válido para pruebas, utiliza el proyecto dedicado:
[**Generador de Token JWT**](https://github.com/javiacuna/jwt-token-generator)

Una vez generado, incluir el token en las solicitudes HTTP usando el encabezado `Authorization` con el prefijo `Bearer`:
`Authorization: Bearer <TOKEN_GENERADO>`
