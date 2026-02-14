package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"erpwms/backend-go/internal/common/auth"
	"erpwms/backend-go/internal/common/crypto"
	"erpwms/backend-go/internal/common/security"
	"erpwms/backend-go/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService struct {
	Queries      *sqlcgen.Queries
	JWT          auth.JWTManager
	SearchPepper string
	AuditPepper  string
	Argon        crypto.Argon2Params
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}

func (s AuthService) Login(ctx context.Context, email, password, ua, ip string) (LoginResult, error) {
	emailHash := security.EmailHash(strings.TrimSpace(strings.ToLower(email)), s.SearchPepper)
	u, err := s.Queries.GetUserByEmailHash(ctx, emailHash)
	if err != nil || !crypto.VerifyPassword(password, u.PasswordHash) {
		s.auditAuth(ctx, nil, "auth.login", "failed", ip, ua, map[string]any{"email_hash": emailHash})
		return LoginResult{}, errors.New("invalid credentials")
	}
	access, err := s.JWT.Issue(u.ID.String(), 15*time.Minute)
	if err != nil {
		return LoginResult{}, err
	}
	refreshRaw, refreshHash, err := s.newRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}
	_, err = s.Queries.CreateRefreshSession(ctx, sqlcgen.CreateRefreshSessionParams{
		UserID:      u.ID,
		RefreshHash: refreshHash,
		UaHash:      txt(security.UAHash(ua, s.AuditPepper)),
		IpHash:      txt(security.IPHash(ip, s.AuditPepper)),
		ExpiresAt:   tstz(time.Now().Add(7 * 24 * time.Hour)),
	})
	if err != nil {
		return LoginResult{}, err
	}
	s.auditAuth(ctx, &u.ID, "auth.login", "ok", ip, ua, nil)
	return LoginResult{AccessToken: access, RefreshToken: refreshRaw, UserID: u.ID.String()}, nil
}

func (s AuthService) Refresh(ctx context.Context, refreshRaw, ua, ip string) (LoginResult, error) {
	hash := security.TokenHash(refreshRaw, s.SearchPepper)
	r, err := s.Queries.GetRefreshSessionByHash(ctx, hash)
	if err != nil || r.RevokedAt.Valid || time.Now().After(r.ExpiresAt.Time) {
		s.auditAuth(ctx, nil, "auth.refresh", "failed", ip, ua, nil)
		return LoginResult{}, errors.New("invalid refresh")
	}
	_ = s.Queries.RevokeRefreshSessionByHash(ctx, hash)
	access, _ := s.JWT.Issue(r.UserID.String(), 15*time.Minute)
	newRaw, newHash, _ := s.newRefreshToken()
	_, err = s.Queries.CreateRefreshSession(ctx, sqlcgen.CreateRefreshSessionParams{
		UserID:      r.UserID,
		RefreshHash: newHash,
		UaHash:      txt(security.UAHash(ua, s.AuditPepper)),
		IpHash:      txt(security.IPHash(ip, s.AuditPepper)),
		ExpiresAt:   tstz(time.Now().Add(7 * 24 * time.Hour)),
	})
	if err != nil {
		return LoginResult{}, err
	}
	s.auditAuth(ctx, &r.UserID, "auth.refresh", "ok", ip, ua, nil)
	return LoginResult{AccessToken: access, RefreshToken: newRaw, UserID: r.UserID.String()}, nil
}

func (s AuthService) Logout(ctx context.Context, refreshRaw string, ua, ip string) error {
	h := security.TokenHash(refreshRaw, s.SearchPepper)
	err := s.Queries.RevokeRefreshSessionByHash(ctx, h)
	s.auditAuth(ctx, nil, "auth.logout", "ok", ip, ua, nil)
	return err
}

func (s AuthService) newRefreshToken() (string, string, error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	return raw, security.TokenHash(raw, s.SearchPepper), nil
}

func (s AuthService) auditAuth(ctx context.Context, userID *pgtype.UUID, action, status, ip, ua string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	requestID, _ := ctx.Value("request_id").(string)
	params := sqlcgen.InsertAuditLogParams{
		ActorType: "user", Action: action, Resource: "auth", Status: status,
		IpHash: txt(security.IPHash(ip, s.AuditPepper)), UaHash: txt(security.UAHash(ua, s.AuditPepper)), RequestID: txt(requestID), Metadata: jsonb(metadata),
	}
	if userID != nil {
		params.ActorUserID = *userID
	}
	_ = s.Queries.InsertAuditLog(ctx, params)
}

func txt(v string) pgtype.Text            { return pgtype.Text{String: v, Valid: v != ""} }
func tstz(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
func jsonb(m map[string]any) []byte {
	b, _ := json.Marshal(m)
	return b
}
