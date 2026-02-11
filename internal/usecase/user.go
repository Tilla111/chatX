package service

import (
	"chatX/internal/store"
	"context"
)

type UserSrvc struct {
	repo *store.Storage
}

type User struct {
	ID        int64  `json:"id"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

func (s *UserSrvc) CreateUser(ctx context.Context, user *store.User) error {
	return s.repo.UserStore.CreateUser(ctx, user)
}

func (s *UserSrvc) GetUsers(ctx context.Context, userID int, pg *store.PaginationQuery) ([]User, error) {

	var users []User

	chat, err := s.repo.UserStore.GetAll(ctx, userID, pg)
	if err != nil {
		return nil, err
	}

	for _, u := range chat {
		users = append(users, User{
			ID:        u.ID,
			UserName:  u.UserName,
			Email:     u.Email,
			Password:  string(u.Password),
			CreatedAt: u.CreatedAt,
		})
	}

	return users, nil

}
