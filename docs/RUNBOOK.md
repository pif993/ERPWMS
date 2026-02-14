# RUNBOOK

Placeholder operational document for ERPWMS.

## Key points
- Deny by default access and least privilege.
- Append-only controls for ledger/audit/event streams.
- No secrets in Git; use env/secrets manager.
- DB migrations must use a 4-digit numeric prefix: `0001_...`, `0002_...` (do not use 5-digit variants).
