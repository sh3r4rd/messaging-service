package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository is the interface for the messaging repository.
type Repository interface {
	Ping() error
	CreateMessage(ctx context.Context, msg Message) (*uuid.UUID, error)
	GetConversations(ctx context.Context) ([]Conversation, error)
	GetConversationByID(ctx context.Context, id string) (*Conversation, error)
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
	// Implement the logic to retrieve conversations with communications as participants

	// TODO: Paginate conversations
	const query = `
		SELECT
		c.id AS conversation_id,
		c.created_at,
		COALESCE(
			JSON_AGG(
			JSON_BUILD_OBJECT(
				'id', comm.id,
				'identifier', comm.identifier,
				'type', comm.type
			)
			) FILTER (WHERE comm.id IS NOT NULL), '[]'
		) AS participants
		FROM conversations c
		LEFT JOIN conversation_memberships cm ON cm.conversation_id = c.id
		LEFT JOIN communications comm ON comm.id = cm.communication_id
		GROUP BY c.id, c.created_at
		ORDER BY c.created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var participantsJSON []byte
		if err := rows.Scan(&conv.ID, &conv.CreatedAt, &participantsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(participantsJSON, &conv.Participants); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *PostgresRepository) GetConversationByID(ctx context.Context, id string) (*Conversation, error) {
	const query = `
		SELECT
			c.id AS conversation_id,
			c.created_at,
			COALESCE(
				JSON_AGG(
					JSONB_BUILD_OBJECT(
						'from', m.sender_id,
						'type', m.message_type,
						'body', m.body,
						'attachments', m.attachments,
						'provider_id', m.provider_id,
						'timestamp', m.created_at,
						'id', m.id
					)
					ORDER BY m.created_at
				) FILTER (WHERE m.id IS NOT NULL), '[]'
			) AS messages
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE c.id = $1
		GROUP BY c.id, c.created_at;
	`
	row := r.db.QueryRowContext(ctx, query, id)
	var conv Conversation
	var messagesJSON []byte
	if err := row.Scan(&conv.ID, &conv.CreatedAt, &messagesJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No conversation found with the given ID
		}
		return nil, err
	}
	if err := json.Unmarshal(messagesJSON, &conv.Messages); err != nil {
		return nil, err
	}

	return &conv, nil
}
