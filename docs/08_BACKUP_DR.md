# Backup & Disaster Recovery

- Daily logical backups via `scripts/db/backup.sh`.
- Encrypt backup at rest (gpg/KMS suggested).
- Retention: 30 days hot, 180 days cold.
- Quarterly restore test required.
- Target RPO: 15 min, RTO: 2h.
