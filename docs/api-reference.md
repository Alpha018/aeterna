# Aeterna API Reference

**Base URL:** `http://localhost:3000`  
**Framework:** Go + Fiber v2  
**Content-Type:** `application/json` (excepto uploads: `multipart/form-data`)

---

## Versiones y autenticación

### `v1` — `/api`

Auth por cookie de sesión `aeterna_session` (HTTPOnly, SameSite=Strict).  
Middleware: `MasterAuth`.

### `v2` — `/api/v2`

Orientado a clientes móviles (Android/iOS).  
Auth por `Authorization: Bearer <access_token>`.  
Middleware: `MasterAuthV2` — si no hay Bearer usa cookie como fallback.

> **Regla de rutas protegidas:** todas las rutas protegidas de `/api/...` existen también en `/api/v2/...` con el mismo contrato de request/response. Solo difiere el mecanismo de autenticación.

---

## Formato de errores

```json
{
  "error": "mensaje legible",
  "code":  "error_code",
  "detail": "solo en modo desarrollo (opcional)"
}
```

Códigos de error comunes: `bad_request`, `not_found`, `unauthorized`, `internal_error`, `rate_limited`, `sse_limit_exceeded`.

---

## Setup y Auth

### Estado de setup

```
GET /api/setup/status
GET /api/v2/setup/status
```

**Response `200`:**
```json
{
  "configured": true,
  "allow_registration": false
}
```

---

### Setup inicial (primer usuario)

```
POST /api/setup
POST /api/v2/setup
```

Solo disponible cuando `configured = false`.

**Body:**
```json
{
  "email": "admin@example.com",
  "password": "StrongPass123!",
  "owner_email": "admin@example.com"
}
```

**Response `200` (v1):**
```json
{
  "success": true,
  "recovery_key": "RK-XXXXX-XXXXX-XXXXX-XXXXX"
}
```

**Response `200` (v2):**
```json
{
  "success": true,
  "user_id": "uuid",
  "token_type": "Bearer",
  "access_token": "enc:...",
  "expires_at": "2026-05-15T18:25:43Z",
  "recovery_key": "RK-XXXXX-XXXXX-XXXXX-XXXXX"
}
```

---

### Registro de usuario adicional

```
POST /api/auth/register
POST /api/v2/auth/register
```

Solo disponible si `allow_registration = true`.

**Body:** mismo que setup.

**Response `200` (v1):** `{ "success": true, "recovery_key": "RK-..." }`  
**Response `200` (v2):** mismo formato Bearer + `recovery_key`.

---

### Login

```
POST /api/auth/login
POST /api/v2/auth/login
```

**Body:**
```json
{
  "email": "user@example.com",
  "password": "StrongPass123!"
}
```

**Response `200` (v1):**
```json
{ "success": true }
```

**Response `200` (v2):**
```json
{
  "success": true,
  "user_id": "uuid",
  "token_type": "Bearer",
  "access_token": "enc:...",
  "expires_at": "2026-05-15T18:25:43Z"
}
```

---

### Verify (alias legacy)

```
POST /api/auth/verify
```

Alias de `POST /api/auth/login` (solo v1, requiere `email`).

---

### Reset de contraseña con recovery key

```
POST /api/auth/reset-password
POST /api/v2/auth/reset-password
```

**Body:**
```json
{
  "email": "user@example.com",
  "recovery_key": "RK-XXXXX-XXXXX-XXXXX-XXXXX",
  "new_password": "NewStrongPass123!"
}
```

**Response `200` (v1):** `{ "success": true, "recovery_key": "RK-YYYYY-..." }`  
**Response `200` (v2):** mismo formato Bearer + `recovery_key` nuevo.

---

### Estado de sesión

```
GET /api/auth/session
GET /api/v2/auth/session
```

**Response `200`:**
```json
{ "authorized": true, "user_id": "uuid" }
```

v2 acepta Bearer o cookie.

---

### Logout

```
POST /api/auth/logout
POST /api/v2/auth/logout
```

**Response `200`:** `{ "success": true }`

Borra la cookie de sesión. En v2 el token Bearer no se invalida en servidor (TTL expira solo).

---

## Mensajes (dead man's switch)

### Ver mensaje público

```
GET /api/messages/:id
GET /api/v2/messages/:id
```

Endpoint público (sin auth). `content` solo se expone si el switch ya fue disparado.

**Response `200`:**
```json
{
  "content": "Contenido del mensaje (vacío si no está triggered)",
  "status": "active | triggered",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Crear mensaje

```
POST /api/messages
POST /api/v2/messages
```

**Body:**
```json
{
  "content": "Mi mensaje",
  "recipient_emails": ["dest1@example.com", "dest2@example.com"],
  "recipient_email": "legacy@example.com",
  "trigger_duration": 43200,
  "reminders": [1440, 60]
}
```

- `trigger_duration`: minutos hasta el trigger. Rango: 1–525600 (1 min – 1 año).
- `reminders`: array de enteros (minutos antes del trigger). Puede ser vacío.
- `recipient_emails` tiene prioridad. `recipient_email` es un alias legacy de un solo destinatario.

**Response `200`:**
```json
{
  "id": "uuid-del-mensaje",
  "message": "Dead man's switch activated!"
}
```

---

### Listar mensajes

```
GET /api/messages
GET /api/v2/messages
```

**Response `200`:** array de objetos Message.

```json
[
  {
    "id": "uuid",
    "content": "Contenido descifrado",
    "recipient_email": "dest@example.com",
    "trigger_duration": 43200,
    "last_seen": "2024-01-01T00:00:00Z",
    "status": "active",
    "triggered_at": null,
    "reminders": [
      { "id": 1, "message_id": "uuid", "minutes_before": 1440, "sent": false }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "attachment_count": 2
  }
]
```

Sin mensajes → `[]`.

---

### Actualizar mensaje

```
PUT /api/messages/:id
PUT /api/v2/messages/:id
```

**Body:** mismo que crear.

**Response `200`:**
```json
{
  "success": true,
  "message": { ... }
}
```

El campo `message` es el objeto Message actualizado (mismo shape que en List).

---

### Eliminar mensaje

```
DELETE /api/messages/:id
DELETE /api/v2/messages/:id
```

**Response `200`:**
```json
{ "success": true, "message": "Message deleted successfully" }
```

---

### Heartbeat autenticado

```
POST /api/heartbeat
POST /api/v2/heartbeat
```

**Body:**
```json
{ "id": "uuid-del-mensaje" }
```

**Response `200`:**
```json
{
  "status": "alive",
  "last_seen": "2024-01-01T12:00:00Z"
}
```

---

### Heartbeat rápido (sin auth, HTML)

```
GET  /api/quick-heartbeat/:token
POST /api/quick-heartbeat/:token
```

No existe en v2. Devuelve `text/html`.

- `GET`: muestra página con botón "Send Heartbeat".
- `POST`: ejecuta heartbeat masivo para todos los switches del usuario del token y devuelve página de confirmación.

---

### Obtener token de heartbeat rápido

```
GET /api/heartbeat-token
GET /api/v2/heartbeat-token
```

**Response `200`:**
```json
{ "token": "abc123xyz..." }
```

---

## Adjuntos de mensaje

**Límites:**
- Máximo 5 archivos por mensaje.
- Máximo 10 MB por archivo.
- Máximo 25 MB acumulado por mensaje.

### Subir adjunto

```
POST /api/messages/:id/attachments
POST /api/v2/messages/:id/attachments
```

`multipart/form-data`, campo `file`.

**Response `200`:**
```json
{
  "success": true,
  "attachment": {
    "id": "uuid",
    "message_id": "uuid",
    "filename": "documento.pdf",
    "size": 102400,
    "mime_type": "application/pdf",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### Listar adjuntos

```
GET /api/messages/:id/attachments
GET /api/v2/messages/:id/attachments
```

**Response `200`:** array de objetos Attachment.

```json
[
  {
    "id": "uuid",
    "message_id": "uuid",
    "filename": "documento.pdf",
    "size": 102400,
    "mime_type": "application/pdf",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

Sin adjuntos → `[]`.

---

### Eliminar adjunto

```
DELETE /api/messages/:id/attachments/:attachmentId
DELETE /api/v2/messages/:id/attachments/:attachmentId
```

**Response `200`:**
```json
{ "success": true, "message": "Attachment deleted successfully" }
```

---

## Farewell Letters

Una farewell letter es un correo diferido que se envía después de que el switch se dispara.  
Ruta base: `/api/messages/:id/farewell-letters`

> **Nota:** el path es `/farewell-letters` (con guion y plural completo), no `/farewells`.

### Listar farewell letters

```
GET /api/messages/:id/farewell-letters
GET /api/v2/messages/:id/farewell-letters
```

**Response `200`:** array de objetos FarewellLetter.

```json
[
  {
    "id": "uuid",
    "message_id": "uuid",
    "recipient_email": "recipient@example.com",
    "subject": "Carta personal",
    "content": "Contenido descifrado",
    "delay_minutes": 1440,
    "status": "pending",
    "sent_at": null,
    "attachment_count": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

Sin cartas → `[]`.  
Si el `:id` del mensaje no existe → `404 { "error": "Message not found", "code": "not_found" }`.

---

### Crear farewell letter

```
POST /api/messages/:id/farewell-letters
POST /api/v2/messages/:id/farewell-letters
```

**Body:**
```json
{
  "recipient_email": "recipient@example.com",
  "subject": "Carta personal",
  "content": "Contenido de la carta...",
  "delay_minutes": 1440
}
```

- `delay_minutes`: minutos después del trigger para enviar. Mínimo `0` (envío inmediato). Máximo `525600` (1 año).
- No se puede crear ni editar una letter con `status = "sent"`.

**Response `201`:** objeto FarewellLetter (mismo shape que en el listado).

---

### Actualizar farewell letter

```
PUT /api/messages/:id/farewell-letters/:letterId
PUT /api/v2/messages/:id/farewell-letters/:letterId
```

**Body:** mismo que crear.

**Response `200`:** objeto FarewellLetter actualizado.

---

### Eliminar farewell letter

```
DELETE /api/messages/:id/farewell-letters/:letterId
DELETE /api/v2/messages/:id/farewell-letters/:letterId
```

**Response `200`:**
```json
{ "success": true, "message": "Farewell letter deleted" }
```

---

## Adjuntos de farewell letters

**Límites:**
- Máximo 10 archivos por letter.
- Máximo 20 MB por archivo.
- Máximo 50 MB acumulado por letter.

### Subir adjunto

```
POST /api/messages/:id/farewell-letters/:letterId/attachments
POST /api/v2/messages/:id/farewell-letters/:letterId/attachments
```

`multipart/form-data`, campo `file`.

**Response `200`:**
```json
{
  "success": true,
  "attachment": {
    "id": "uuid",
    "letter_id": "uuid",
    "filename": "foto.jpg",
    "size": 204800,
    "mime_type": "image/jpeg",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### Listar adjuntos

```
GET /api/messages/:id/farewell-letters/:letterId/attachments
GET /api/v2/messages/:id/farewell-letters/:letterId/attachments
```

**Response `200`:** array de objetos FarewellAttachment (mismo shape que arriba, sin `success`).

Sin adjuntos → `[]`.

---

### Eliminar adjunto

```
DELETE /api/messages/:id/farewell-letters/:letterId/attachments/:attachmentId
DELETE /api/v2/messages/:id/farewell-letters/:letterId/attachments/:attachmentId
```

**Response `200`:**
```json
{ "success": true, "message": "Farewell attachment deleted" }
```

---

## Webhooks

### Listar webhooks

```
GET /api/webhooks
GET /api/v2/webhooks
```

**Response `200`:** array de objetos Webhook. `secret` siempre vacío en respuestas.

```json
[
  {
    "id": 1,
    "url": "https://example.com/webhook",
    "secret": "",
    "enabled": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

Sin webhooks → `[]`.

---

### Crear webhook

```
POST /api/webhooks
POST /api/v2/webhooks
```

**Body:**
```json
{
  "url": "https://example.com/webhook",
  "secret": "mi-secreto-hmac",
  "enabled": true
}
```

**Response `200`:** objeto Webhook creado (`secret` = `""`).

---

### Actualizar webhook

```
PUT /api/webhooks/:id
PUT /api/v2/webhooks/:id
```

**Body:** mismo que crear. Si `secret` está vacío, conserva el secreto anterior.

**Response `200`:** objeto Webhook actualizado (`secret` = `""`).

---

### Eliminar webhook

```
DELETE /api/webhooks/:id
DELETE /api/v2/webhooks/:id
```

**Response `200`:** `{ "success": true }`

---

## Configuración (Settings)

### Obtener settings

```
GET /api/settings
GET /api/v2/settings
```

**Response `200`:**
```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": "587",
  "smtp_user": "user@gmail.com",
  "smtp_from": "user@gmail.com",
  "smtp_from_name": "Mi Nombre",
  "webhook_url": "https://example.com/hook",
  "webhook_enabled": false,
  "owner_email": "admin@example.com",
  "allow_registration": false,
  "can_manage_registration": true
}
```

Campos no expuestos: `smtp_pass`, `webhook_secret`, `heartbeat_token`.

---

### Guardar settings

```
POST /api/settings
POST /api/v2/settings
```

**Body:**
```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": "587",
  "smtp_user": "user@gmail.com",
  "smtp_pass": "app-password",
  "smtp_from": "user@gmail.com",
  "smtp_from_name": "Mi Nombre",
  "webhook_url": "https://example.com/hook",
  "webhook_secret": "secreto",
  "webhook_enabled": true,
  "owner_email": "admin@example.com",
  "allow_registration": false
}
```

`allow_registration` solo puede ser modificado por el usuario primario.

**Response `200`:** `{ "success": true }`

---

### Probar SMTP

```
POST /api/settings/test
POST /api/v2/settings/test
```

**Body:** mismo que guardar settings (no persiste, solo prueba la conexión).

**Response `200`:**
```json
{ "success": true, "message": "Connection successful" }
```

---

## Eventos en tiempo real (SSE)

```
GET /api/events
GET /api/v2/events
```

Stream de eventos Server-Sent Events (SSE). Requiere auth.

**Query params:**
- `client_id` (opcional): identificador del cliente. Si se omite, el servidor genera uno.

**Cabeceras de respuesta:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**Formato de cada evento SSE:**
```
event: <tipo>
data: {"type":"<tipo>","code":"<codigo>","at":"2024-01-01T00:00:00Z","data":{"resource":"","entity_id":"","reason":""},"resource":"","entity_id":"","reason":""}
```

**Contrato recomendado para clientes nuevos (Android/Web):**
- Usar `type` para suscribirse (`addEventListener("messages.changed", ...)`).
- Usar `code` para lógica de notificación (campanita, badges, reglas de UX).
- Usar `data` para detalles del evento.
- Mantener compatibilidad leyendo `resource/entity_id/reason` top-level si `data` no existe.

**Tipos de eventos:**

| Tipo | Cuándo se emite |
|---|---|
| `ready` | Al conectar (inmediato) |
| `ping` | Cada 20 segundos (keepalive) |
| `messages.changed` | Al crear, editar, eliminar o hacer heartbeat de un switch |
| `attachments.changed` | Al subir o eliminar un adjunto de mensaje |
| `farewells.changed` | Al crear, editar, eliminar una farewell letter o sus adjuntos |
| `settings.changed` | Al guardar settings |
| `webhooks.changed` | Al crear, editar o eliminar un webhook |

**Códigos de evento (`code`):**
- Stream: `stream.ready`, `stream.ping`
- Message: `message.created`, `message.updated`, `message.deleted`, `message.heartbeat`, `message.bulk_heartbeat`, `message.attachment_uploaded`, `message.attachment_deleted`, `message.farewell_created`, `message.farewell_updated`, `message.farewell_deleted`
- Attachment: `attachment.uploaded`, `attachment.deleted`
- Farewell: `farewell.created`, `farewell.updated`, `farewell.deleted`, `farewell_attachment.uploaded`, `farewell_attachment.deleted`
- Settings/Webhook: `settings.saved`, `webhook.created`, `webhook.updated`, `webhook.deleted`

**Mapa canónico `type -> codes` (recomendado para Web + Android):**
```json
{
  "ready": [
    "stream.ready"
  ],
  "ping": [
    "stream.ping"
  ],
  "messages.changed": [
    "message.created",
    "message.updated",
    "message.deleted",
    "message.heartbeat",
    "message.bulk_heartbeat",
    "message.attachment_uploaded",
    "message.attachment_deleted",
    "message.farewell_created",
    "message.farewell_updated",
    "message.farewell_deleted"
  ],
  "attachments.changed": [
    "attachment.uploaded",
    "attachment.deleted"
  ],
  "farewells.changed": [
    "farewell.created",
    "farewell.updated",
    "farewell.deleted",
    "farewell_attachment.uploaded",
    "farewell_attachment.deleted"
  ],
  "settings.changed": [
    "settings.saved"
  ],
  "webhooks.changed": [
    "webhook.created",
    "webhook.updated",
    "webhook.deleted"
  ]
}
```

**Campos de `data` (payload normalizado):**

| Campo | Tipo | Presencia | Descripción |
|---|---|---|---|
| `resource` | `string` | opcional | Clase de entidad afectada (`message`, `attachment`, `farewell`, `farewell_attachment`, `settings`, `webhook`) |
| `entity_id` | `string` | opcional | ID de la entidad afectada cuando aplica |
| `reason` | `string` | opcional | Motivo de mutación (ej: `created`, `updated`, `heartbeat`) |

Notas de compatibilidad:
- `data.resource`, `data.entity_id` y `data.reason` se reflejan también en top-level como `resource`, `entity_id`, `reason`.
- Clientes nuevos deben priorizar `data`; top-level queda como fallback legacy.

**Matriz de payload `data` por `code`:**

| `code` | `data.resource` | `data.entity_id` | `data.reason` |
|---|---|---|---|
| `stream.ready` | — | — | `connected` |
| `stream.ping` | — | — | — |
| `message.created` | `message` | `<message_id>` | `created` |
| `message.updated` | `message` | `<message_id>` | `updated` |
| `message.deleted` | `message` | `<message_id>` | `deleted` |
| `message.heartbeat` | `message` | `<message_id>` | `heartbeat` |
| `message.bulk_heartbeat` | `message` | — | `bulk_heartbeat` |
| `message.attachment_uploaded` | `message` | `<message_id>` | `attachment_uploaded` |
| `message.attachment_deleted` | `message` | — | `attachment_deleted` |
| `message.farewell_created` | `message` | `<message_id>` | `farewell_created` |
| `message.farewell_updated` | `message` | `<message_id>` | `farewell_updated` |
| `message.farewell_deleted` | `message` | `<message_id>` | `farewell_deleted` |
| `attachment.uploaded` | `attachment` | `<attachment_id>` | `uploaded` |
| `attachment.deleted` | `attachment` | `<attachment_id>` | `deleted` |
| `farewell.created` | `farewell` | `<farewell_id>` | `created` |
| `farewell.updated` | `farewell` | `<farewell_id>` | `updated` |
| `farewell.deleted` | `farewell` | `<farewell_id>` | `deleted` |
| `farewell_attachment.uploaded` | `farewell_attachment` | `<attachment_id>` | `attachment_uploaded` |
| `farewell_attachment.deleted` | `farewell_attachment` | `<attachment_id>` | `attachment_deleted` |
| `settings.saved` | `settings` | — | `saved` |
| `webhook.created` | `webhook` | `<webhook_id>` | `created` |
| `webhook.updated` | `webhook` | `<webhook_id>` | `updated` |
| `webhook.deleted` | `webhook` | `<webhook_id>` | `deleted` |

**Error `429`** (demasiadas conexiones SSE abiertas del mismo usuario):
```json
{ "error": "...", "code": "sse_limit_exceeded" }
```

---

## Usuarios (solo administrador primario)

### Listar usuarios

```
GET /api/users
GET /api/v2/users
```

**Response `200`:**
```json
[
  {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "is_primary": true
  }
]
```

---

### Eliminar usuario

```
DELETE /api/users/:id
DELETE /api/v2/users/:id
```

No se puede eliminar al usuario primario ni a uno mismo.

**Response `200`:** `{ "success": true }`

---

## Seguridad y middleware

### `MasterAuth` (v1)

- Valida cookie `aeterna_session`.
- Enforce de origin allowlist (`ALLOWED_ORIGINS`) en todas las rutas protegidas.
- Sin origin en producción → `403`.

### `MasterAuthV2` (v2)

- Acepta `Authorization: Bearer <token>`.
- Si no hay Bearer, fallback a cookie con las mismas validaciones de `MasterAuth`.

### `AuthRateLimiter`

Aplicado a `register`, `login`, `verify`, `reset-password`.

- Ventana de intentos: 5 minutos.
- Máximo 5 intentos fallidos antes del primer bloqueo.
- Lock progresivo con backoff exponencial: 1 min → 2 min → 4 min → 8 min → máx 15 min.
- Éxito reinicia el contador.
- Respuesta `429` incluye `retry_after_secs`.

### Rate limiter global

- 120 requests por minuto por IP (todas las rutas).

### `SecurityHeaders`

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), camera=(), microphone=(), payment=()`
- `Content-Security-Policy: default-src 'self'; img-src 'self' data:`
- `Strict-Transport-Security` (solo producción)
- `Cache-Control: no-cache, no-store, must-revalidate` en rutas `/api*`

### CORS

- `AllowMethods: GET, POST, PUT, DELETE, OPTIONS`
- `AllowHeaders: Origin, Content-Type, Accept, Authorization`
- `AllowCredentials: true`
- `AllowOrigins`: valor de `ALLOWED_ORIGINS` (no puede ser `*` en producción salvo `PROXY_MODE=simple`)

---

## Resumen de rutas

### Públicas (v1)

| Método | Ruta |
|---|---|
| GET | `/api/setup/status` |
| POST | `/api/setup` |
| POST | `/api/auth/register` |
| POST | `/api/auth/login` |
| POST | `/api/auth/verify` |
| POST | `/api/auth/reset-password` |
| GET | `/api/auth/session` |
| POST | `/api/auth/logout` |
| GET | `/api/messages/:id` |
| GET | `/api/quick-heartbeat/:token` |
| POST | `/api/quick-heartbeat/:token` |

### Públicas (v2)

Igual que v1 excepto: no existe `/api/v2/auth/verify` ni `/api/v2/quick-heartbeat/:token`.

### Protegidas (iguales en v1 y v2)

| Método | Ruta |
|---|---|
| POST | `/messages` |
| GET | `/messages` |
| PUT | `/messages/:id` |
| DELETE | `/messages/:id` |
| POST | `/heartbeat` |
| GET | `/heartbeat-token` |
| POST | `/messages/:id/attachments` |
| GET | `/messages/:id/attachments` |
| DELETE | `/messages/:id/attachments/:attachmentId` |
| GET | `/messages/:id/farewell-letters` |
| POST | `/messages/:id/farewell-letters` |
| PUT | `/messages/:id/farewell-letters/:letterId` |
| DELETE | `/messages/:id/farewell-letters/:letterId` |
| POST | `/messages/:id/farewell-letters/:letterId/attachments` |
| GET | `/messages/:id/farewell-letters/:letterId/attachments` |
| DELETE | `/messages/:id/farewell-letters/:letterId/attachments/:attachmentId` |
| GET | `/webhooks` |
| POST | `/webhooks` |
| PUT | `/webhooks/:id` |
| DELETE | `/webhooks/:id` |
| GET | `/settings` |
| POST | `/settings` |
| POST | `/settings/test` |
| GET | `/events` |
| GET | `/users` |
| DELETE | `/users/:id` |
