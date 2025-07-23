package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository is the interface for the messaging repository.
type Repository interface {
	Ping() error
	CreateMessage(ctx context.Context, msg Message) (*uuid.UUID, error)
	GetConversations(ctx context.Context) ([]Conversation, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func (r *PostgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, msg Message) (*uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // safest when concurrent inserts are possible
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() // no‑op if Commit succeeded

	const query = `
		SELECT create_message(
			$1, -- p_from
			$2, -- p_to
			$3, -- p_provider_id
			$4, -- p_message_type
			$5, -- p_communication_type
			$6, -- p_body
			$7, -- p_attachments
			$8  -- p_created_at
		)
	`
	var msgID uuid.UUID
	if err := tx.QueryRowContext(
		ctx,
		query,
		msg.From,
		msg.To,
		msg.ProviderID,
		msg.Type,
		msg.CommunicationType,
		msg.Body,
		pq.Array(msg.Attachments),
		msg.CreatedAt,
	).Scan(&msgID); err != nil {
		return nil, err
	}

	// TODO: add custom error handling
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &msgID, nil
}
func (r *PostgresRepository) GetConversations(ctx context.Context) ([]Conversation, error) {
	// Implement the logic to retrieve conversations from the database.
	// This is just a placeholder implementation.
	fmt.Println("Retrieving conversations...")
	return []Conversation{}, nil
}
