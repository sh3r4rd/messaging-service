package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hatchapp/internal/pkg/apperrors"

	"github.com/labstack/gommon/log"
	"github.com/lib/pq"
)

// Repository is the interface for the messaging repository.
type Repository interface {
	Ping() error
	CreateMessage(ctx context.Context, msg Message) (*int64, error)
	GetConversations(ctx context.Context) ([]Conversation, error)
	GetConversationByID(ctx context.Context, id string) (*Conversation, error)
	Close() error
	GetDriver() *sql.DB
}

type PostgresRepository struct {
	db *sql.DB
}

func (r *PostgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *PostgresRepository) Close() error {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Errorf("failed to close database connection: %v", err)
			return apperrors.NewDBError(err, "failed to close database connection")
		}
		log.Info("Database connection closed successfully")
		return nil
	}
	log.Warn("Attempted to close a nil database connection, ignoring.")
	return nil
}

func (r *PostgresRepository) GetDriver() *sql.DB {
	if r.db == nil {
		log.Error("Attempted to get a nil database connection, returning nil.")
		return nil
	}
	return r.db
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, msg Message) (*int64, error) {
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
	var msgID int64
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
		return nil, apperrors.NewDBError(err, "failed to create message")
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("failed to commit transaction: %v", err)
		return nil, apperrors.NewDBError(err, "failed to commit transaction")
	}
	return &msgID, nil
}
func (r *PostgresRepository) GetConversations(ctx context.Context) ([]Conversation, error) {
	const query = `
		SELECT
		c.id AS conversation_id,
		c.created_at,
		COALESCE(
			JSONB_AGG(
				JSONB_BUILD_OBJECT(
					'id', comm.id,
					'identifier', comm.identifier,
					'type', comm.communication_type
				)
			) FILTER (WHERE comm.id IS NOT NULL), '[]'
		) AS participants
		FROM conversations c
		LEFT JOIN conversation_memberships cm ON cm.conversation_id = c.id
		LEFT JOIN communications comm ON comm.id = cm.communication_id
		GROUP BY c.id
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
			return nil, apperrors.NewDBError(err, "failed to get conversations")
		}

		if err := json.Unmarshal(participantsJSON, &conv.Participants); err != nil {
			return nil, apperrors.NewDBError(err, "failed to unmarshal participants for conversation")
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
				JSONB_AGG(
					JSONB_BUILD_OBJECT(
						'id', m.id,
						'from', comm.identifier,
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
		LEFT JOIN communications comm ON comm.id = m.sender_id
		WHERE c.id = $1
		GROUP BY c.id;
	`

	var conv Conversation
	var messagesJSON []byte
	row := r.db.QueryRowContext(ctx, query, id)

	if err := row.Scan(&conv.ID, &conv.CreatedAt, &messagesJSON); err != nil {
		if err == sql.ErrNoRows {
			log.Info("no conversation found with id:", id)
			return nil, apperrors.DBErrorNotFound
		}

		return nil, apperrors.NewDBError(err, fmt.Sprintf("failed to get conversation with id: %s", id))
	}

	if err := json.Unmarshal(messagesJSON, &conv.Messages); err != nil {
		return nil, apperrors.NewDBError(err, "failed to unmarshal conversation messages")
	}

	return &conv, nil
}
