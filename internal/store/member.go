package store

import "context"

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
		&member.ChatID,
		&member.UserID,
		&member.Rol,
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

// deleteMember

func (s *MemberStorage) Delete(ctx context.Context, chatID, userID int) error {
	query := `DELETE FROM chat_members m WHERE (m.chat_id = $1 AND (m.user_id = $2)`

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
