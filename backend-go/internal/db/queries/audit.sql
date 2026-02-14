-- name: InsertAuditLog :exec
INSERT INTO audit_log (
  actor_user_id, actor_type, action, resource, resource_id, status,
  ip_hash, ua_hash, request_id, trace_id, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);
