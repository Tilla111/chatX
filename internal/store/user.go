package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64  `json:"id"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) CreateUser(ctx context.Context, user *User) error {

	query := `INSERT INTO users(username, email, password)
	VALUES ($1, $2, $3) RETURNING id, created_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		&user.UserName,
		&user.Email,
		&user.Password,
	).Scan(
		&user.ID,
		&user.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetUser
func (s *UserStore) GetAll(ctx context.Context, currentUserID int, pg *PaginationQuery) ([]User, error) {
	query := `SELECT id, username, email 
              FROM users 
              WHERE ($1 = '' OR username ILIKE '%' || $1 || '%')
              AND id != $2 
              ORDER BY username ASC 
              LIMIT $3 OFFSET $4;`

	rows, err := s.db.QueryContext(ctx, query, pg.Search, currentUserID, pg.Limit, pg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.UserName, &u.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
