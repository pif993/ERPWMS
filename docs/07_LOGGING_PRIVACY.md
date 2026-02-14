# Logging & Privacy

- JSON structured logs with request_id and trace_id.
- Never log plaintext emails, phones, tokens, passwords.
- Use redaction helpers (`internal/common/pii`).
- `LOG_REDACTION=true` in production.
