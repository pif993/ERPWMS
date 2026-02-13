# Security Operations

## Controls
- Argon2id for passwords.
- AES-256-GCM envelope encryption for sensitive fields.
- JWT + field encryption key rotation (`current` + `previous`).
- CSRF, CORS allow-list, secure headers, rate limiting.
- Append-only triggers on audit and ledger tables.

## TLS/mTLS
- Public TLS terminated at Caddy.
- Optional internal mTLS (API <-> analytics) via cert path env vars.

## Network policy
K8s baseline:
- deny-all ingress/egress
- allow ingress from ingress controller to API/analytics
- allow API/worker egress to Postgres/Redis/NATS
- allow analytics egress only to Postgres read-only service

## Container hardening
- Run as non-root user.
- Prefer read-only rootfs (enabled in K8s manifests).
- Drop Linux capabilities in compose/k8s where possible.
