# API Reference

Base URL: `http://localhost:8080/api/v1`

Interactive docs: `http://localhost:8080/swagger/index.html`

## Authentication

All protected endpoints require:
```
Authorization: Bearer <access_token>
```

---

## Auth

### POST /auth/register
Create a new account.

**Body**
```json
{
  "name": "John Owner",
  "email": "john@example.com",
  "password": "SecurePass1!"
}
```

**Response 201**
```json
{
  "id": "01926b4f-...",
  "name": "John Owner",
  "email": "john@example.com",
  "role": "user",
  "active": true,
  "created_at": "2026-05-19T10:00:00Z"
}
```

---

### POST /auth/login
```json
{ "email": "john@example.com", "password": "SecurePass1!" }
```
**Response 200**
```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_at": "2026-05-20T10:00:00Z"
}
```

---

### POST /auth/refresh
```json
{ "refresh_token": "eyJ..." }
```

---

## Rooms (Owner/Admin)

### GET /rooms
Public. Query params: `page`, `limit`, `status`, `floor`, `search`

**Response 200**
```json
{
  "data": [
    {
      "id": "01926b4f-...",
      "number": "101",
      "floor": 1,
      "type": "studio",
      "size_sqm": 28.5,
      "rent_price": 5000,
      "deposit": 10000,
      "status": "available",
      "amenities": ["wifi", "air_conditioner", "hot_water"],
      "photos": ["https://cdn.example.com/room101.jpg"],
      "description": "Corner studio with city view",
      "owner_id": "01926b4f-..."
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

### GET /rooms/:id
Public.

### POST /rooms
**Owner/Admin only.**
```json
{
  "number": "102",
  "floor": 1,
  "type": "studio",
  "size_sqm": 28.5,
  "rent_price": 5000,
  "deposit": 10000,
  "amenities": ["wifi", "air_conditioner"],
  "photos": [],
  "description": "Studio room"
}
```

Room types: `studio` | `one_bedroom` | `two_bedroom` | `suite` | `loft`

### PUT /rooms/:id
**Owner/Admin only.** Same body as POST (all fields optional).

### DELETE /rooms/:id
**Owner/Admin only.**

---

## Bookings

### POST /bookings
**Tenant/User/Admin.**
```json
{
  "room_id": "01926b4f-...",
  "start_date": "2026-06-01",
  "end_date": "2026-12-31",
  "note": "Requesting early check-in"
}
```

### GET /bookings
**Owner/Admin only.** Query: `page`, `limit`, `status`, `room_id`, `tenant_id`

### GET /bookings/:id
**Owner/Admin only.**

### POST /bookings/:id/approve
**Owner/Admin only.**

### POST /bookings/:id/reject
**Owner/Admin only.**
```json
{ "reason": "Room reserved for renovation" }
```

### POST /bookings/:id/activate
**Owner/Admin only.** Moves booking to active, sets room to occupied.

### POST /bookings/:id/complete
**Owner/Admin only.**

### POST /bookings/:id/cancel
**Tenant/User/Admin.** Tenant can only cancel their own booking.

Booking statuses: `pending` → `approved`/`rejected` → `active` → `completed`/`cancelled`

---

## Payments

### POST /payments
**Tenant/User/Admin.**
```json
{
  "invoice_id": "01926b4f-...",
  "amount": 5000,
  "method": "bank_transfer",
  "gateway": "omise",
  "gateway_ref": "chrg_test_123"
}
```

Payment methods: `cash` | `bank_transfer` | `credit_card` | `qr_code` | `omise` | `stripe`

### GET /payments
**Owner/Admin only.** Query: `page`, `limit`, `status`, `tenant_id`

### GET /payments/:id
**Owner/Admin only.**

### POST /payments/:id/refund
**Owner/Admin only.**

### POST /payments/webhook/:gateway
**Public** (secured by gateway signature). Processes payment gateway callbacks.

```json
{
  "event": "payment.success",
  "reference": "chrg_test_123",
  "amount": 5000,
  "currency": "THB",
  "status": "successful",
  "metadata": {}
}
```

---

## Invoices

### POST /invoices
**Owner/Admin only.**
```json
{
  "tenant_id": "01926b4f-...",
  "booking_id": "01926b4f-...",
  "due_date": "2026-06-30",
  "items": [
    { "description": "Monthly Rent - June 2026", "quantity": 1, "unit_price": 5000 },
    { "description": "Water Bill", "quantity": 15, "unit_price": 18 }
  ]
}
```

### GET /invoices
**Owner/Admin only.** Query: `page`, `limit`, `status`, `tenant_id`

### GET /invoices/:id

### POST /invoices/:id/send
**Owner/Admin only.** Sends invoice to tenant (email notification).

### POST /invoices/:id/pay
**Owner/Admin only.** Manually mark as paid.

### POST /invoices/:id/cancel
**Owner/Admin only.**

### GET /invoices/:id/download
Returns print-ready HTML invoice (use browser Print → Save as PDF).

### GET /tenants/:id/invoices
**Tenant/User/Admin.** Tenant's own invoices.

Invoice statuses: `draft` → `sent` → `paid`/`overdue`/`cancelled`

---

## Maintenance

### POST /maintenance
**Tenant/User/Admin.**
```json
{
  "room_id": "01926b4f-...",
  "title": "Broken AC",
  "description": "Air conditioner not cooling",
  "priority": "high"
}
```

Priorities: `low` | `medium` | `high` | `urgent`

### GET /maintenance
**Owner/Admin only.** Query: `page`, `limit`, `status`, `priority`, `room_id`

### GET /maintenance/:id

### POST /maintenance/:id/start
**Owner/Admin only.**

### POST /maintenance/:id/resolve
**Owner/Admin only.**
```json
{ "resolution_note": "Replaced refrigerant, tested OK" }
```

### POST /maintenance/:id/close
**Tenant/User/Admin.** Tenant confirms issue resolved.

Statuses: `open` → `in_progress` → `resolved` → `closed`

---

## Notices

### GET /notices
Public. Query: `page`, `limit`, `type`, `active_only`

Notice types: `general` | `maintenance` | `event` | `emergency` | `billing`

### GET /notices/:id
Public.

### POST /notices
**Owner/Admin only.**
```json
{
  "title": "Water Supply Maintenance",
  "body": "Water will be cut from 09:00-12:00 on June 5.",
  "type": "maintenance",
  "pinned": false,
  "expires_at": "2026-06-06T00:00:00Z"
}
```

### PUT /notices/:id
**Owner/Admin only.**

### DELETE /notices/:id
**Owner/Admin only.**

---

## Messages

### POST /messages
**Any authenticated user.**
```json
{
  "recipient_id": "01926b4f-...",
  "content": "Hello, I have a question about my invoice."
}
```

### GET /messages/threads
**Any authenticated user.** Lists all threads the caller participates in.

### GET /messages/threads/:threadId
**Any authenticated user.** Lists messages in thread. Query: `page`, `limit`

### POST /messages/threads/:threadId/read
**Any authenticated user.** Marks all messages in thread as read.

---

## Reports (Owner/Admin)

### GET /owners/:id/dashboard
**Owner/Admin only.**

**Response 200**
```json
{
  "owner_id": "01926b4f-...",
  "total_rooms": 20,
  "occupied_rooms": 15,
  "available_rooms": 4,
  "maintenance_rooms": 1,
  "pending_bookings": 3,
  "open_maintenance_tickets": 2,
  "unpaid_invoices": 5,
  "monthly_income": 75000,
  "occupancy_rate": 75.0,
  "generated_at": "2026-05-19T10:00:00Z"
}
```

### GET /reports/income
**Owner/Admin only.** Query: `from` (YYYY-MM-DD), `to` (YYYY-MM-DD)

**Response 200**
```json
{
  "from": "2026-05-01T00:00:00Z",
  "to": "2026-05-31T23:59:59Z",
  "total_income": 150000,
  "total_paid_invoices": 28,
  "total_pending_invoices": 5,
  "monthly_breakdown": [
    { "month": "2026-05", "income": 150000, "invoice_count": 28 }
  ]
}
```

### GET /tenants/:id/history
**Tenant/User/Admin.** Tenant's payment and invoice history.

---

## Users

### GET /users
**Owner/Admin only.** Query: `page`, `limit`, `role`, `active`, `search`

### GET /users/:id
**Any authenticated user.**

### PUT /users/:id
**Any authenticated user** (own profile only; admin can update any).
```json
{ "name": "Updated Name" }
```

### DELETE /users/:id
**Any authenticated user** (own account; admin can delete any).

---

## Health

### GET /health
Liveness check (always 200 if process running).

### GET /ready
Readiness check (pings MongoDB; 503 if unavailable).

### GET /version
```json
{
  "version": "v1.2.3",
  "commit": "abc1234",
  "build_time": "2026-05-19T10:00:00Z"
}
```

---

## Error Responses

All errors follow this shape:

```json
{
  "code": "NOT_FOUND",
  "message": "room not found",
  "details": [],
  "request_id": "01926b4f-..."
}
```

| HTTP | Code | Meaning |
|---|---|---|
| 400 | `BAD_REQUEST` | Invalid request body or parameters |
| 400 | `VALIDATION_FAILED` | Field validation errors (details array populated) |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 403 | `FORBIDDEN` | Authenticated but insufficient role |
| 404 | `NOT_FOUND` | Resource does not exist |
| 409 | `CONFLICT` | Duplicate resource or state conflict |
| 429 | `TOO_MANY_REQUESTS` | Rate limit exceeded |
| 500 | `INTERNAL_ERROR` | Server error (details hidden in production) |

---

## Pagination

All list endpoints accept:
- `page` (default: 1)
- `limit` (default: 20, max: 100)

Response includes `total`, `page`, `limit` alongside `data` array.
