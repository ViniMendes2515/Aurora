package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"

	"aurora/services/notifications-service/internal/domain"
)

// PostgresTelegramPreferenceRepository implementa TelegramPreferenceRepository com PostgreSQL
type PostgresTelegramPreferenceRepository struct {
	db *sql.DB
}

func NewPostgresTelegramPreferenceRepository(db *sql.DB) *PostgresTelegramPreferenceRepository {
	return &PostgresTelegramPreferenceRepository{db: db}
}

func (r *PostgresTelegramPreferenceRepository) Save(pref *domain.TelegramPreference) error {
	_, err := r.db.Exec(`
		INSERT INTO telegram_preferences (user_id, chat_id, enabled_types, active, linked_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET chat_id = EXCLUDED.chat_id,
		    enabled_types = EXCLUDED.enabled_types,
		    active = EXCLUDED.active,
		    linked_at = EXCLUDED.linked_at
	`, pref.UserID, pref.ChatID, pq.Array(pref.EnabledTypes), pref.Active, pref.LinkedAt)
	return err
}

func (r *PostgresTelegramPreferenceRepository) FindByUserID(userID string) (*domain.TelegramPreference, error) {
	return r.scanPref(r.db.QueryRow(`
		SELECT id, user_id, chat_id, enabled_types, active, linked_at
		FROM telegram_preferences WHERE user_id = $1
	`, userID))
}

func (r *PostgresTelegramPreferenceRepository) FindByChatID(chatID int64) (*domain.TelegramPreference, error) {
	return r.scanPref(r.db.QueryRow(`
		SELECT id, user_id, chat_id, enabled_types, active, linked_at
		FROM telegram_preferences WHERE chat_id = $1
	`, chatID))
}

func (r *PostgresTelegramPreferenceRepository) scanPref(row *sql.Row) (*domain.TelegramPreference, error) {
	var p domain.TelegramPreference
	var types pq.StringArray
	err := row.Scan(&p.ID, &p.UserID, &p.ChatID, &types, &p.Active, &p.LinkedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTelegramNotLinked
	}
	if err != nil {
		return nil, err
	}
	p.EnabledTypes = []string(types)
	return &p, nil
}

func (r *PostgresTelegramPreferenceRepository) UpdateTypes(userID string, types []string) error {
	_, err := r.db.Exec(`
		UPDATE telegram_preferences SET enabled_types = $1 WHERE user_id = $2
	`, pq.Array(types), userID)
	return err
}

func (r *PostgresTelegramPreferenceRepository) SetActive(userID string, active bool) error {
	_, err := r.db.Exec(`
		UPDATE telegram_preferences SET active = $1 WHERE user_id = $2
	`, active, userID)
	return err
}

func (r *PostgresTelegramPreferenceRepository) Delete(userID string) error {
	_, err := r.db.Exec(`DELETE FROM telegram_preferences WHERE user_id = $1`, userID)
	return err
}

func (r *PostgresTelegramPreferenceRepository) SaveLinkToken(token *domain.TelegramLinkToken) error {
	_, err := r.db.Exec(`
		INSERT INTO telegram_link_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id, expires_at = EXCLUDED.expires_at
	`, token.Token, token.UserID, token.ExpiresAt)
	return err
}

func (r *PostgresTelegramPreferenceRepository) FindLinkToken(token string) (*domain.TelegramLinkToken, error) {
	var t domain.TelegramLinkToken
	err := r.db.QueryRow(`
		SELECT token, user_id, expires_at FROM telegram_link_tokens WHERE token = $1
	`, strings.ToUpper(token)).Scan(&t.Token, &t.UserID, &t.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTelegramTokenExpired
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(t.ExpiresAt) {
		_ = r.DeleteLinkToken(token)
		return nil, domain.ErrTelegramTokenExpired
	}
	return &t, nil
}

func (r *PostgresTelegramPreferenceRepository) DeleteLinkToken(token string) error {
	_, err := r.db.Exec(`DELETE FROM telegram_link_tokens WHERE token = $1`, token)
	return err
}
