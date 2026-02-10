package store

import (
	"context"
	"database/sql"
)

type Group struct {
	ID          int64  `json:"id"`
	ChatID      int64  `json:"chat_id"`
	GroupName   string `json:"group_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Groupstorage struct {
	db DBTX
}

func (s *Groupstorage) CreateGroup(ctx context.Context, group *Group) error {
	query := `INSERT INTO group_info (chat_id, group_name, group_description) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := s.db.QueryRowContext(
		ctx,
		query,
		group.ChatID,
		group.GroupName,
		group.Description,
	).Scan(
		&group.ID,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Groupstorage) Update(ctx context.Context, group *Group) (*Group, error) {

	query := `UPDATE group_info 
              SET group_name = $1, group_description = $2, updated_at = NOW() 
              WHERE chat_id = $3 RETURNING id, chat_id, group_name, group_description, updated_at;`

	err := s.db.QueryRowContext(
		ctx,
		query,
		group.GroupName,
		group.Description,
		group.ChatID,
	).Scan(
		&group.ID,
		&group.ChatID,
		&group.GroupName,
		&group.Description,
		&group.UpdatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, SqlNotfound
		default:
			return nil, err
		}
	}

	return group, nil

}
