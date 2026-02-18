package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	UserName  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type Password struct {
	text string
	hash []byte
}

func (p *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = text
	p.hash = hash

	return nil
}

type UserStore struct {
	db DBTX
}

func (s *UserStore) CreateUser(ctx context.Context, user *User) error {

	query := `INSERT INTO users(username, email, password)
	VALUES ($1, $2, $3) RETURNING id, created_at, is_active`

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.UserName,
		user.Email,
		user.Password.hash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
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

func (s *UserStore) CreateTokenActivate(ctx context.Context, userID int64, token string, exp time.Duration) error {
	query := `INSERT INTO user_invitations(token, user_id, expiry) VALUES($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) GetUserByToken(ctx context.Context, token string) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at, u.is_active 
              FROM users AS u 
              JOIN user_invitations AS ui ON ui.user_id = u.id
              WHERE ui.token = $1;`

	var u User
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// QueryRowContext ishlatamiz, chunki bizga 1 ta user kerak
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&u.ID,
		&u.UserName,
		&u.Email,
		&u.CreatedAt,
		&u.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, SqlNotfound // O'zingiz yaratgan xato
		}
		return nil, err
	}

	return &u, nil
}

func (s *UserStore) Update(ctx context.Context, user *User) error {

	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, user.UserName, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return SqlNotfound
	}

	return nil
}

func (s *UserStore) Clean(ctx context.Context, token string) error {
	query := `DELETE FROM User_invitations WHERE token = $1`

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return SqlNotfound
	}

	return nil
}
