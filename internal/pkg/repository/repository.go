package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hatchapp/internal/pkg/apperrors"
	"time"

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
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Ensure rollback on error, unless committed

	var fromID, toID, conversationID, messageID int64

	// 1. Upsert communications
	upsertCommQuery := `
		INSERT INTO communications (identifier, communication_type)
		VALUES ($1, $2)
		ON CONFLICT (identifier) DO NOTHING
	`

	if _, err := tx.ExecContext(ctx, upsertCommQuery, msg.From, msg.CommunicationType); err != nil {
		return nil, apperrors.NewDBError(err, "failed to upsert communication for sender")
	}
	if _, err := tx.ExecContext(ctx, upsertCommQuery, msg.To, msg.CommunicationType); err != nil {
		return nil, apperrors.NewDBError(err, "failed to upsert communication for recipient")
	}

	// 2. Lookup IDs for both communications
	getCommIDQuery := `SELECT id FROM communications WHERE identifier = $1`
	if err := tx.QueryRowContext(ctx, getCommIDQuery, msg.From).Scan(&fromID); err != nil {
		return nil, apperrors.NewDBError(err, "failed to get sender communication ID")
	}
	if err := tx.QueryRowContext(ctx, getCommIDQuery, msg.To).Scan(&toID); err != nil {
		return nil, apperrors.NewDBError(err, "failed to get recipient communication ID")
	}

	// 3. Try to find an existing conversation with just these two participants
	findConversationQuery := `
		SELECT cm.conversation_id
		FROM conversation_memberships cm
		WHERE cm.communication_id IN ($1, $2)
		GROUP BY cm.conversation_id
		HAVING COUNT(*) = 2 AND bool_and(cm.communication_id IN ($1, $2))
		LIMIT 1
	`

	if err := tx.QueryRowContext(ctx, findConversationQuery, fromID, toID).Scan(&conversationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 4. Create new conversation
			createConvQuery := `INSERT INTO conversations (created_at) VALUES (now()) RETURNING id`
			if err := tx.QueryRowContext(ctx, createConvQuery).Scan(&conversationID); err != nil {
				return nil, apperrors.NewDBError(err, "failed to create new conversation")
			}

			// 5. Insert conversation memberships
			insertMembershipQuery := `
				INSERT INTO conversation_memberships (conversation_id, communication_id)
				VALUES ($1, $2), ($1, $3)
				ON CONFLICT DO NOTHING
			`
			if _, err := tx.ExecContext(ctx, insertMembershipQuery, conversationID, fromID, toID); err != nil {
				return nil, apperrors.NewDBError(err, "failed to insert conversation memberships")
			}
		} else {
			return nil, err
		}
	}

	// 6. Insert the message
	insertMessageQuery := `
		INSERT INTO messages (
			conversation_id,
			sender_id,
			provider_id,
			message_type,
			body,
			attachments,
			created_at,
			message_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	if err := tx.QueryRowContext(ctx, insertMessageQuery,
		conversationID,
		fromID,
		msg.ProviderID,
		msg.Type,
		msg.Body,
		pq.Array(msg.Attachments),
		msg.CreatedAt,
		msg.Status,
	).Scan(&messageID); err != nil {
		return nil, apperrors.NewDBError(err, "failed to insert message")
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.NewDBError(err, "failed to commit transaction")
	}

	return &messageID, nil
}

func (r *PostgresRepository) GetConversations(ctx context.Context) ([]Conversation, error) {
	const query = `
		SELECT
		  c.id AS conversation_id,
		  c.created_at,
		  comm.id AS participant_id,
		  comm.identifier,
		  comm.communication_type
		FROM conversations c
		LEFT JOIN conversation_memberships cm ON cm.conversation_id = c.id
		LEFT JOIN communications comm ON comm.id = cm.communication_id
		ORDER BY c.created_at DESC;
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversationMap := make(map[int64]*Conversation)
	orderedIDs := make([]int64, 0)

	for rows.Next() {
		var convID int64
		var createdAt time.Time
		var participantID sql.NullInt64
		var identifier sql.NullString
		var commType sql.NullString

		if err := rows.Scan(&convID, &createdAt, &participantID, &identifier, &commType); err != nil {
			return nil, apperrors.NewDBError(err, "failed to scan conversation row")
		}

		conv, exists := conversationMap[convID]
		if !exists {
			conv = &Conversation{
				ID:           convID,
				CreatedAt:    createdAt.Format(time.RFC3339),
				Participants: []Communication{},
			}
			conversationMap[convID] = conv
			orderedIDs = append(orderedIDs, convID) // record retrieval order
		}

		if participantID.Valid {
			conv.Participants = append(conv.Participants, Communication{
				ID:         participantID.Int64,
				Identifier: identifier.String,
				Type:       commType.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDBError(err, "encountered error while iterating database rows")
	}

	conversations := make([]Conversation, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		conversations = append(conversations, *conversationMap[id])
	}

	return conversations, nil
}

func (r *PostgresRepository) GetConversationByID(ctx context.Context, id string) (*Conversation, error) {
	const query = `
		SELECT
			c.id AS conversation_id,
			c.created_at,
			m.id AS message_id,
			comm.identifier AS sender_identifier,
			m.message_type,
			m.body,
			m.attachments,
			m.provider_id,
			m.created_at AS message_created_at
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		LEFT JOIN communications comm ON comm.id = m.sender_id
		WHERE c.id = $1 AND m.message_status = $2
		ORDER BY m.created_at;
	`

	rows, err := r.db.QueryContext(ctx, query, id, MessageStatusSuccess)
	if err != nil {
		return nil, apperrors.NewDBError(err, fmt.Sprintf("failed to query conversation %s", id))
	}
	defer rows.Close()

	var conv *Conversation
	for rows.Next() {
		var (
			convID      int64
			createdAt   time.Time
			msgID       sql.NullInt64
			from        sql.NullString
			msgType     sql.NullString
			body        sql.NullString
			attachments pq.StringArray
			providerID  sql.NullString
			timestamp   sql.NullTime
		)

		if err := rows.Scan(
			&convID, &createdAt,
			&msgID, &from, &msgType, &body,
			&attachments, &providerID, &timestamp,
		); err != nil {
			return nil, apperrors.NewDBError(err, "failed to scan conversation row")
		}

		if conv == nil {
			conv = &Conversation{
				ID:        convID,
				CreatedAt: createdAt.Format(time.RFC3339),
				Messages:  []Message{},
			}
		}

		if msgID.Valid {
			conv.Messages = append(conv.Messages, Message{
				ID:          msgID.Int64,
				From:        from.String,
				Type:        msgType.String,
				Body:        body.String,
				Attachments: attachments,
				ProviderID:  providerID.String,
				CreatedAt:   timestamp.Time.Format(time.RFC3339),
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDBError(err, "encountered error while iterating database rows")
	}

	if conv == nil {
		return nil, apperrors.DBErrorNotFound
	}

	return conv, nil
}
