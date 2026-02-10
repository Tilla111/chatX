package store

import (
	"context"
	"database/sql"
	"time"
)

type Chat struct {
	ID        int64  `json:"id"`
	ChatType  string `json:"chat_type"`
	CreatedAt string `json:"created_at"`
}

type ChatInfo struct {
	ChatID        int64  `json:"chat_id"`
	ChatType      string `json:"chat_type"`
	ChatName      string `json:"chat_name"`
	UserRole      string `json:"user_role"`
	JoinedAt      string `json:"joined_at"`
	LastMessage   string `json:"last_message"`
	LastMessageAt string `json:"last_message_at"`
	UnreadCount   int    `json:"unread_count"`
}

type Chatcheck struct {
	UserID   int64
	MemberID int64
}

type Chatstorage struct {
	db DBTX
}

func (s *Chatstorage) Createchat(ctx context.Context, chat *Chat) (int64, error) {
	query := `INSERT INTO chats (chat_type) VALUES ($1) RETURNING id, created_at`
	err := s.db.QueryRowContext(
		ctx,
		query,
		&chat.ChatType,
	).Scan(
		&chat.ID,
		&chat.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return chat.ID, nil

}

func (s *Chatstorage) GetByID(ctx context.Context, ChatID int64) (*Chat, error) {
	query := `SELECT id, chat_type, created_at FROM chats WHERE id = $1`

	chat := &Chat{}

	err := s.db.QueryRowContext(
		ctx,
		query,
		ChatID,
	).Scan(
		&chat.ID,
		&chat.ChatType,
		&chat.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *Chatstorage) GetChatByUserID(ctx context.Context, id int64, searchTerm string) ([]*ChatInfo, error) {
	query := `
    SELECT 
        c.id AS chat_id,
        c.chat_type,
        CASE 
            WHEN c.chat_type = 'group' THEN COALESCE(gi.group_name, 'Group')
            ELSE (
                SELECT u.username FROM chat_members cm2 
                JOIN users u ON u.id = cm2.user_id 
                WHERE cm2.chat_id = c.id AND cm2.user_id != $1 LIMIT 1
            )
        END AS chat_name,
        cm.rol AS user_role,
        cm.joined_at,
        COALESCE(m.message_text, '') AS last_message,
        m.created_at AS last_message_at,
        -- O'qilmagan xabarlar sonini hisoblash
        (SELECT COUNT(*) FROM messages m2 
         WHERE m2.chat_id = c.id 
           AND m2.sender_id != $1 
           AND m2.is_read = FALSE) AS unread_count
    FROM chat_members cm
    JOIN chats c ON cm.chat_id = c.id
    LEFT JOIN group_info gi ON c.id = gi.chat_id
    LEFT JOIN (
        SELECT DISTINCT ON (chat_id) chat_id, message_text, created_at
        FROM messages
        ORDER BY chat_id, created_at DESC
    ) m ON m.chat_id = c.id
    WHERE cm.user_id = $1 
      AND (
          $2 = '' 
          OR gi.group_name ILIKE '%' || $2 || '%'
          OR EXISTS (
              SELECT 1 FROM chat_members cm3
              JOIN users u2 ON u2.id = cm3.user_id
              WHERE cm3.chat_id = c.id AND cm3.user_id != $1 AND u2.username ILIKE '%' || $2 || '%'
          )
      )
    ORDER BY last_message_at DESC NULLS LAST, cm.joined_at DESC;`

	rows, err := s.db.QueryContext(ctx, query, id, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*ChatInfo
	for rows.Next() {
		var c ChatInfo
		var lastMsgAt *time.Time

		err := rows.Scan(
			&c.ChatID,
			&c.ChatType,
			&c.ChatName,
			&c.UserRole,
			&c.JoinedAt,
			&c.LastMessage,
			&lastMsgAt,
			&c.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		if lastMsgAt != nil {
			c.LastMessageAt = lastMsgAt.Format("2006-01-02 15:04:05")
		}

		chats = append(chats, &c)
	}
	return chats, nil
}

func (s *Chatstorage) CheckChatP(ctx context.Context, users *Chatcheck) (int64, error) {
	query := `SELECT c.id 
    FROM chats c
    JOIN chat_members cm1 ON c.id = cm1.chat_id
    JOIN chat_members cm2 ON c.id = cm2.chat_id
    WHERE c.chat_type = 'private'
       AND cm1.user_id = $1  
       AND cm2.user_id = $2 
    LIMIT 1;`

	var ChatID int64
	err := s.db.QueryRowContext(
		ctx,
		query,
		users.UserID,
		users.MemberID,
	).Scan(
		&ChatID,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return 0, SqlNotfound
		default:
			return 0, err
		}
	}

	return ChatID, nil
}

func (s *Chatstorage) Delete(ctx context.Context, chatID int) error {
	query := `DELETE FROM chats WHERE id = $1`

	rows, err := s.db.ExecContext(
		ctx,
		query,
		chatID,
	)
	if err != nil {
		return err
	}

	id, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if id == 0 {
		return SqlNotfound
	}

	return nil
}
