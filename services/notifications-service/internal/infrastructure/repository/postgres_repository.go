package repository

import (
	"database/sql"

	"aurora/services/notifications-service/internal/domain"
)

// PostgresNotificationRepository implementa NotificationRepository usando PostgreSQL
type PostgresNotificationRepository struct {
	db *sql.DB
}

// NewPostgresNotificationRepository cria uma nova instancia do repositorio
func NewPostgresNotificationRepository(db *sql.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

// Save persiste uma nova notificacao no banco
func (r *PostgresNotificationRepository) Save(n *domain.Notification) error {
	_, err := r.db.Exec(
		`INSERT INTO notifications (id, user_id, type, title, message, sensor_id, location, read, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		n.ID, n.UserID, string(n.Type), n.Title, n.Message,
		n.SensorID, n.Location, n.Read, n.CreatedAt,
	)
	return err
}

// FindByUserID retorna notificacoes proprias do usuario e notificacoes broadcast (user_id="*")
func (r *PostgresNotificationRepository) FindByUserID(userID string) ([]*domain.Notification, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, type, title, message, sensor_id, location, read, created_at
		 FROM notifications
		 WHERE user_id = $1 OR user_id = '*'
		 ORDER BY created_at DESC
		 LIMIT 100`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var notifType string
		if err := rows.Scan(
			&n.ID, &n.UserID, &notifType, &n.Title, &n.Message,
			&n.SensorID, &n.Location, &n.Read, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		n.Type = domain.NotificationType(notifType)
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// FindByID busca uma notificacao pelo seu ID
func (r *PostgresNotificationRepository) FindByID(id string) (*domain.Notification, error) {
	var n domain.Notification
	var notifType string
	err := r.db.QueryRow(
		`SELECT id, user_id, type, title, message, sensor_id, location, read, created_at
		 FROM notifications WHERE id = $1`,
		id,
	).Scan(
		&n.ID, &n.UserID, &notifType, &n.Title, &n.Message,
		&n.SensorID, &n.Location, &n.Read, &n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotificationNotFound
	}
	if err != nil {
		return nil, err
	}
	n.Type = domain.NotificationType(notifType)
	return &n, nil
}

// MarkAsRead atualiza o campo read=true para a notificacao com o ID fornecido
func (r *PostgresNotificationRepository) MarkAsRead(id string) error {
	res, err := r.db.Exec(`UPDATE notifications SET read = TRUE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.ErrNotificationNotFound
	}
	return nil
}
