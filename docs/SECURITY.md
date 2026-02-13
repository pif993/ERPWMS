# Security Operations

## Implementato in M1
- Password hashing con Argon2id.
- Field-level encryption AES-256-GCM (`CURRENT` + `PREVIOUS` key, lazy rotate).
- Email hash per lookup (`SEARCH_PEPPER`) e hash audit per IP/UA (`AUDIT_PEPPER`).
- JWT validation con key corrente + precedente.
- Secure headers + CORS restrittivo + request id.
- Ledger/audit append-only via trigger DB.
- Outbox aggiornabile solo su `sent_at`, `attempts`, `last_error`.

## Rotazione chiavi
- JWT: impostare `JWT_SIGNING_KEY_CURRENT` e mantenere la precedente in `JWT_SIGNING_KEY_PREVIOUS` durante la finestra di rollout.
- Field encryption: impostare `FIELD_ENC_MASTER_KEY_CURRENT` + `FIELD_ENC_MASTER_KEY_PREVIOUS` e usare `RotateIfNeeded` in read-path.

## Logging e privacy
- Non loggare Authorization/Cookie/token in chiaro.
- Salvare in audit solo hash IP/UA.

## Container hardening
- immagini con utente non-root.
- in K8s abilitare readOnlyRootFilesystem e drop capabilities.
