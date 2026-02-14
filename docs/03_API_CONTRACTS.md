# API Contracts

## Auth
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`

## WMS
- `GET /api/stock/balances`
- `POST /api/stock/moves` (requires `Idempotency-Key`)

## Orders
- `POST /api/orders`
- `POST /api/orders/{id}/allocate`

## Health
- `GET /health`
