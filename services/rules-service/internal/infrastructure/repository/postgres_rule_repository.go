package repository

import (
	"database/sql"

	"aurora/services/rules-service/internal/domain"
)

// PostgresRuleRepository implementa domain.RuleRepository
type PostgresRuleRepository struct {
	db *sql.DB
}

// NewPostgresRuleRepository cria uma nova instância do repositório
func NewPostgresRuleRepository(db *sql.DB) *PostgresRuleRepository {
	return &PostgresRuleRepository{db: db}
}

// Save persiste uma regra (insert ou update)
func (r *PostgresRuleRepository) Save(rule *domain.Rule) error {
	query := `
		INSERT INTO rules (id, name, owner_id, trigger_type, trigger_id, trigger_threshold, action_type, action_id, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			trigger_type = EXCLUDED.trigger_type,
			trigger_id = EXCLUDED.trigger_id,
			trigger_threshold = EXCLUDED.trigger_threshold,
			action_type = EXCLUDED.action_type,
			action_id = EXCLUDED.action_id,
			enabled = EXCLUDED.enabled,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query,
		rule.ID, rule.Name, rule.OwnerID, string(rule.TriggerType), rule.TriggerID, rule.TriggerThreshold,
		string(rule.ActionType), rule.ActionID, rule.Enabled, rule.CreatedAt, rule.UpdatedAt,
	)
	return err
}

// FindByID busca uma regra pelo ID
func (r *PostgresRuleRepository) FindByID(id string) (*domain.Rule, error) {
	query := `
		SELECT id, name, owner_id, trigger_type, trigger_id, COALESCE(trigger_threshold, 30), action_type, action_id, enabled, created_at, updated_at
		FROM rules WHERE id = $1
	`
	rule := &domain.Rule{}
	var triggerType, actionType string
	err := r.db.QueryRow(query, id).Scan(
		&rule.ID, &rule.Name, &rule.OwnerID, &triggerType, &rule.TriggerID, &rule.TriggerThreshold,
		&actionType, &rule.ActionID, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	rule.TriggerType = domain.TriggerType(triggerType)
	rule.ActionType = domain.ActionType(actionType)
	return rule, nil
}

// FindByOwnerID busca regras pelo ID do proprietário
func (r *PostgresRuleRepository) FindByOwnerID(ownerID string) ([]*domain.Rule, error) {
	query := `
		SELECT id, name, owner_id, trigger_type, trigger_id, COALESCE(trigger_threshold, 30), action_type, action_id, enabled, created_at, updated_at
		FROM rules WHERE owner_id = $1 OR owner_id = '*'
		ORDER BY created_at DESC
	`
	return r.scanRows(r.db.Query(query, ownerID))
}

// FindByTrigger busca regras por tipo e ID de gatilho
func (r *PostgresRuleRepository) FindByTrigger(triggerType domain.TriggerType, triggerID string) ([]*domain.Rule, error) {
	query := `
		SELECT id, name, owner_id, trigger_type, trigger_id, COALESCE(trigger_threshold, 30), action_type, action_id, enabled, created_at, updated_at
		FROM rules
		WHERE trigger_type = $1 AND (trigger_id = $2 OR trigger_id = '') AND enabled = true
	`
	return r.scanRows(r.db.Query(query, string(triggerType), triggerID))
}

// Delete remove uma regra
func (r *PostgresRuleRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM rules WHERE id = $1", id)
	return err
}

func (r *PostgresRuleRepository) scanRows(rows *sql.Rows, err error) ([]*domain.Rule, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.Rule
	for rows.Next() {
		rule := &domain.Rule{}
		var triggerType, actionType string
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.OwnerID, &triggerType, &rule.TriggerID, &rule.TriggerThreshold,
			&actionType, &rule.ActionID, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rule.TriggerType = domain.TriggerType(triggerType)
		rule.ActionType = domain.ActionType(actionType)
		rules = append(rules, rule)
	}
	return rules, nil
}
