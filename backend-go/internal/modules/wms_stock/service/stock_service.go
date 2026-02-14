package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"erpwms/backend-go/internal/db/sqlcgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockService struct {
	DB      *pgxpool.Pool
	Queries *sqlcgen.Queries
}

type MoveRequest struct {
	ItemID         string `json:"item_id"`
	Qty            string `json:"qty"`
	FromLocationID string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	ReasonCode     string `json:"reason_code"`
}

type MoveResponse struct {
	MoveID string `json:"move_id"`
	Status string `json:"status"`
}

func (s StockService) MoveStock(ctx context.Context, req MoveRequest, actor uuid.UUID, endpoint, idemKey string) (MoveResponse, error) {
	reqHash, _ := hashReq(req)
	existing, err := s.Queries.GetIdempotency(ctx, sqlcgen.GetIdempotencyParams{Key: idemKey, Endpoint: endpoint})
	if err == nil {
		if existing.RequestHash != reqHash {
			return MoveResponse{}, errors.New("idempotency conflict")
		}
		var r MoveResponse
		_ = json.Unmarshal(existing.ResponseJson, &r)
		return r, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return MoveResponse{}, err
	}

	itemID, err := scanUUID(req.ItemID)
	if err != nil {
		return MoveResponse{}, err
	}
	fromID, err := scanUUID(req.FromLocationID)
	if err != nil {
		return MoveResponse{}, err
	}
	toID, err := scanUUID(req.ToLocationID)
	if err != nil {
		return MoveResponse{}, err
	}
	qty, err := scanNumeric(req.Qty)
	if err != nil {
		return MoveResponse{}, err
	}
	negQty, err := scanNumeric("-" + req.Qty)
	if err != nil {
		return MoveResponse{}, err
	}
	actorID, _ := scanUUID(actor.String())
	requestID, _ := ctx.Value("request_id").(string)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return MoveResponse{}, err
	}
	defer tx.Rollback(ctx)
	q := s.Queries.WithTx(tx)

	move, err := q.InsertStockLedgerMove(ctx, sqlcgen.InsertStockLedgerMoveParams{ItemID: itemID, Qty: qty, FromLocationID: fromID, ToLocationID: toID, ReasonCode: req.ReasonCode, ActorUserID: actorID, RequestID: txt(requestID)})
	if err != nil {
		return MoveResponse{}, err
	}
	if err := q.UpsertStockBalanceDelta(ctx, sqlcgen.UpsertStockBalanceDeltaParams{ItemID: itemID, LocationID: fromID, QtyOnHand: negQty, QtyAllocated: mustNumeric("0")}); err != nil {
		return MoveResponse{}, err
	}
	if err := q.UpsertStockBalanceDelta(ctx, sqlcgen.UpsertStockBalanceDeltaParams{ItemID: itemID, LocationID: toID, QtyOnHand: qty, QtyAllocated: mustNumeric("0")}); err != nil {
		return MoveResponse{}, err
	}

	payload, _ := json.Marshal(map[string]any{"move_id": move.MoveID.String(), "item_id": req.ItemID, "qty": req.Qty})
	if _, err := q.InsertOutboxEvent(ctx, sqlcgen.InsertOutboxEventParams{Topic: "stock.moved", Payload: payload}); err != nil {
		return MoveResponse{}, err
	}
	_ = q.InsertAuditLog(ctx, sqlcgen.InsertAuditLogParams{ActorUserID: actorID, ActorType: "user", Action: "stock.move", Resource: "stock_ledger", ResourceID: txt(move.MoveID.String()), Status: "ok", RequestID: txt(requestID), Metadata: payload})

	resp := MoveResponse{MoveID: move.MoveID.String(), Status: "ok"}
	respJSON, _ := json.Marshal(resp)
	if err := q.InsertIdempotency(ctx, sqlcgen.InsertIdempotencyParams{Key: idemKey, Endpoint: endpoint, ActorUserID: actorID, RequestHash: reqHash, ResponseJson: respJSON}); err != nil {
		return MoveResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MoveResponse{}, err
	}
	return resp, nil
}

func hashReq(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func scanUUID(v string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(v); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid %q", v)
	}
	return u, nil
}

func scanNumeric(v string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(v); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric %q", v)
	}
	return n, nil
}

func mustNumeric(v string) pgtype.Numeric { n, _ := scanNumeric(v); return n }
func txt(v string) pgtype.Text            { return pgtype.Text{String: v, Valid: v != ""} }
