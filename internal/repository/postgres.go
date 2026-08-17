package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mojtabatorabi/whatsapp-bot/internal/model"
)

type PostgresMessageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMessageRepository(
	db *pgxpool.Pool,
) *PostgresMessageRepository {

	return &PostgresMessageRepository{
		db: db,
	}
}

func (r *PostgresMessageRepository) Save(
	ctx context.Context,
	userID string,
	message model.Message,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO messages
		(user_id, role, content)
		VALUES ($1, $2, $3)
		`,
		userID,
		message.Role,
		message.Content,
	)

	return err
}

func (r *PostgresMessageRepository) GetMessages(
	ctx context.Context,
	userID string,
) ([]model.Message, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT role, content
		FROM messages
		WHERE user_id = $1
		ORDER BY created_at ASC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []model.Message

	for rows.Next() {

		var message model.Message

		err := rows.Scan(
			&message.Role,
			&message.Content,
		)

		if err != nil {
			return nil, err
		}

		messages = append(
			messages,
			message,
		)
	}

	return messages, rows.Err()
}

func (r *PostgresMessageRepository) Clear(
	ctx context.Context,
	userID string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		DELETE FROM messages
		WHERE user_id = $1
		`,
		userID,
	)

	return err
}
