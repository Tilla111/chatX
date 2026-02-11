package store

import (
	"context"
	"database/sql"
)

type Member struct {
	ID       int64  `json:"id"`
	ChatID   int64  `json:"chat_id"`
	UserID   int64  `json:"user_id"`
	Rol      string `json:"rol"`
	JoinedAt string `json:"joined_at"`
}

type MemberStorage struct {
	db DBTX
}

func (s *MemberStorage) AddMember(ctx context.Context, member *Member) error {
	query := `INSERT INTO chat_members(chat_id, user_id, rol) 
	VALUES ($1, $2, $3) RETURNING id, joined_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		member.ChatID,
		member.UserID,
		member.Rol,
	).Scan(
		&member.ID,
		&member.JoinedAt,
	)

	return err

}

// getBYCHatID
func (s *MemberStorage) GetByChatID(ctx context.Context, ChatID int) ([]User, error) {

	query := `
        SELECT 
            u.id, 
            u.username, 
            u.email 
        FROM users u 
        INNER JOIN chat_members c ON u.id = c.user_id 
        WHERE c.chat_id = $1`

	rows, err := s.db.QueryContext(ctx, query, ChatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []User

	for rows.Next() {
		var m User
		if err := rows.Scan(&m.ID, &m.UserName, &m.Email); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *MemberStorage) GetRole(ctx context.Context, chatID int64, userID int64) (string, error) {
	query := `SELECT rol FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	var role string
	err := s.db.QueryRowContext(ctx, query, chatID, userID).Scan(&role)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return "", SqlNotfound
		default:
			return "", err
		}
	}

	return role, nil
}

func (s *MemberStorage) IsMember(ctx context.Context, chatID int64, userID int64) (bool, error) {
	query := `SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	var exists int
	err := s.db.QueryRowContext(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// deleteMember
func (s *MemberStorage) Delete(ctx context.Context, chatID, userID int) error {
	query := `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	rows, err := s.db.ExecContext(
		ctx,
		query,
		chatID,
		userID,
	)
	if err != nil {
		return err
	}

	row, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return SqlNotfound
	}

	return nil
}
