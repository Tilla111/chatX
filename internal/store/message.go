package store

import (
	"context"
	"database/sql"
)

type Message struct {
	ID          int64
	ChatID      int64
	SenderID    int64
	MessageText string
	IsRead      bool
	CreatedAt   string
	UpdatedAt   string
	SenderName  string
	ChatName    string
}

type MessageStorage struct {
	db DBTX
}

// create
func (s *MessageStorage) Create(ctx context.Context, msg *Message) (Message, error) {
	query := `WITH inserted_msg AS (
    INSERT INTO messages (chat_id, sender_id, message_text) 
    VALUES ($1, $2, $3) 
    RETURNING id, chat_id, sender_id, message_text, is_read, created_at, updated_at
)
SELECT 
    m.id, 
    m.chat_id, 
    m.sender_id, 
    m.message_text, 
    m.is_read, 
    m.created_at, 
    m.updated_at,
    u.username AS sender_name,
    CASE 
        WHEN c.chat_type = 'group' THEN gi.group_name
        ELSE (
            SELECT u2.username 
            FROM chat_members cm 
            JOIN users u2 ON cm.user_id = u2.id 
            WHERE cm.chat_id = m.chat_id AND cm.user_id != m.sender_id
            LIMIT 1
        )
    END AS chat_name
FROM inserted_msg m
JOIN users u ON m.sender_id = u.id
JOIN chats c ON m.chat_id = c.id
LEFT JOIN group_info gi ON c.id = gi.chat_id;`

	var result Message
	err := s.db.QueryRowContext(ctx, query, msg.ChatID, msg.SenderID, msg.MessageText).
		Scan(
			&result.ID, &result.ChatID, &result.SenderID, &result.MessageText,
			&result.IsRead, &result.CreatedAt, &result.UpdatedAt,
			&result.SenderName, &result.ChatName,
		)

	if err != nil {
		return Message{}, err
	}

	return result, nil
}

// GetByCha
type MessageDetail struct {
	ID         int64
	Content    string
	SenderID   int64
	SenderName string
	CreatedAt  string
}

func (s *MessageStorage) GetByID(ctx context.Context, id int64) (*Message, error) {
	query := `
        SELECT id, chat_id, sender_id, message_text, is_read, created_at, updated_at 
        FROM messages 
        WHERE id = $1`

	var m Message
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.ChatID, &m.SenderID, &m.MessageText, &m.IsRead, &m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, SqlNotfound
		default:
			return nil, err
		}
	}

	return &m, nil
}

func (s *MessageStorage) GetMessages(ctx context.Context, chatID int64) ([]MessageDetail, error) {
	query := `
        SELECT m.id, m.message_text, m.sender_id, u.username as sender_name, m.created_at
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        WHERE m.chat_id = $1
        ORDER BY m.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageDetail
	for rows.Next() {
		var msg MessageDetail
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.SenderID, &msg.SenderName, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *MessageStorage) MarkAsRead(ctx context.Context, chatID, userID int64) error {
	query := `
        INSERT INTO message_reads (message_id, user_id, read_at)
        SELECT m.id, $2, NOW()
        FROM messages m
        WHERE m.chat_id = $1
          AND m.sender_id <> $2
        ON CONFLICT (message_id, user_id) DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, chatID, userID)
	return err
}

// Update
func (s *MessageStorage) Update(ctx context.Context, msgID, userID int64, newText string) error {
	query := `UPDATE messages SET message_text = $1, updated_at = NOW() 
              WHERE id = $2 AND sender_id = $3`
	result, err := s.db.ExecContext(ctx, query, newText, msgID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return SqlNotfound
	}

	return nil
}

// Delete
func (s *MessageStorage) Delete(ctx context.Context, msgID, userID int64) error {
	query := `DELETE FROM messages WHERE id = $1 AND sender_id = $2`
	result, err := s.db.ExecContext(ctx, query, msgID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return SqlNotfound
	}

	return nil
}
