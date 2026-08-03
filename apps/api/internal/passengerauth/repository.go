package passengerauth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) ListTravellers(ctx context.Context, accountID uuid.UUID) ([]Traveller, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,full_name,COALESCE(email_normalized,''),COALESCE(phone_e164,''),created_at,updated_at FROM saved_travellers WHERE account_id=$1 ORDER BY created_at,id`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	travellers := make([]Traveller, 0)
	for rows.Next() {
		var traveller Traveller
		if err = rows.Scan(&traveller.ID, &traveller.FullName, &traveller.Email, &traveller.Phone, &traveller.CreatedAt, &traveller.UpdatedAt); err != nil {
			return nil, err
		}
		travellers = append(travellers, traveller)
	}
	return travellers, rows.Err()
}

func (r *Repository) CountTravellers(ctx context.Context, accountID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM saved_travellers WHERE account_id=$1`, accountID).Scan(&count)
	return count, err
}

func (r *Repository) InsertTraveller(ctx context.Context, accountID uuid.UUID, traveller Traveller) (Traveller, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO saved_travellers(id,account_id,full_name,email_normalized,phone_e164) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,'')) RETURNING created_at,updated_at`, traveller.ID, accountID, traveller.FullName, traveller.Email, traveller.Phone).Scan(&traveller.CreatedAt, &traveller.UpdatedAt)
	return traveller, err
}

func (r *Repository) UpdateTraveller(ctx context.Context, accountID uuid.UUID, traveller Traveller) (Traveller, error) {
	err := r.pool.QueryRow(ctx, `UPDATE saved_travellers SET full_name=$3,email_normalized=NULLIF($4,''),phone_e164=NULLIF($5,''),updated_at=now() WHERE id=$1 AND account_id=$2 RETURNING created_at,updated_at`, traveller.ID, accountID, traveller.FullName, traveller.Email, traveller.Phone).Scan(&traveller.CreatedAt, &traveller.UpdatedAt)
	return traveller, err
}

func (r *Repository) DeleteTraveller(ctx context.Context, accountID, travellerID uuid.UUID) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM saved_travellers WHERE id=$1 AND account_id=$2`, travellerID, accountID)
	if err == nil && command.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *Repository) GetPreferences(ctx context.Context, accountID uuid.UUID) (Preferences, error) {
	var preferences Preferences
	err := r.pool.QueryRow(ctx, `SELECT preferred_coach_class,preferred_seat_type,language,updated_at FROM passenger_preferences WHERE account_id=$1`, accountID).Scan(&preferences.PreferredCoachClass, &preferences.PreferredSeatType, &preferences.Language, &preferences.UpdatedAt)
	return preferences, err
}

func (r *Repository) UpsertPreferences(ctx context.Context, accountID uuid.UUID, preferences Preferences) (Preferences, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO passenger_preferences(account_id,preferred_coach_class,preferred_seat_type,language) VALUES($1,$2,$3,$4) ON CONFLICT(account_id) DO UPDATE SET preferred_coach_class=EXCLUDED.preferred_coach_class,preferred_seat_type=EXCLUDED.preferred_seat_type,language=EXCLUDED.language,updated_at=now() RETURNING updated_at`, accountID, preferences.PreferredCoachClass, preferences.PreferredSeatType, preferences.Language).Scan(&preferences.UpdatedAt)
	return preferences, err
}
