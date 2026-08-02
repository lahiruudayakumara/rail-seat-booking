package passengerauth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) InsertAccount(ctx context.Context, account Account, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO passenger_accounts(id,full_name,email_normalized,phone_e164,password_hash) VALUES($1,$2,$3,NULLIF($4,''),$5)`, account.ID, account.FullName, account.Email, account.Phone, passwordHash)
	return err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (Account, string, error) {
	var account Account
	var passwordHash string
	err := r.pool.QueryRow(ctx, `SELECT id,full_name,email_normalized,COALESCE(phone_e164,''),created_at,password_hash FROM passenger_accounts WHERE email_normalized=$1`, email).Scan(&account.ID, &account.FullName, &account.Email, &account.Phone, &account.CreatedAt, &passwordHash)
	return account, passwordHash, err
}

func (r *Repository) InsertSession(ctx context.Context, id, accountID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO passenger_sessions(id,account_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, id, accountID, tokenHash, expiresAt)
	return err
}

func (r *Repository) AccountBySession(ctx context.Context, tokenHash string) (Account, error) {
	var account Account
	err := r.pool.QueryRow(ctx, `SELECT a.id,a.full_name,a.email_normalized,COALESCE(a.phone_e164,''),a.created_at FROM passenger_sessions s JOIN passenger_accounts a ON a.id=s.account_id WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at>now()`, tokenHash).Scan(&account.ID, &account.FullName, &account.Email, &account.Phone, &account.CreatedAt)
	if err == nil {
		_, _ = r.pool.Exec(ctx, `UPDATE passenger_sessions SET last_seen_at=now() WHERE token_hash=$1 AND last_seen_at<now()-interval '5 minutes'`, tokenHash)
	}
	return account, err
}

func (r *Repository) RevokeSession(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE passenger_sessions SET revoked_at=now() WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash)
	return err
}
