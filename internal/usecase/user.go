package service

import (
	"chatX/internal/store"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
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

type RequestRegister struct {
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,email,max=72"`
	Password string `json:"password" validate:"required"`
}

func (s *UserSrvc) RegisterUser(ctx context.Context, user RequestRegister, exp time.Duration) (string, error) {

	//gentoken
	plaintoken := uuid.New().String()
	hash := sha256.Sum256([]byte(plaintoken))
	token := hex.EncodeToString(hash[:])

	//tranzaction
	err := s.repo.UnitOfWork.Do(ctx, func(ctx context.Context, repos *store.Storage) error {
		//createuser
		u := store.User{
			UserName: user.Username,
			Email:    user.Email,
		}

		password := store.Password{}
		if err := password.Set(user.Password); err != nil {
			return err
		}
		u.Password = password

		if err := repos.UserStore.CreateUser(ctx, &u); err != nil {
			return err
		}

		//create user_nvetations
		if err := repos.UserStore.CreateTokenActivate(ctx, u.ID, token, exp); err != nil {
			return err
		}

		//sendEmail

		return nil
	})
	if err != nil {
		return "", err
	}

	return plaintoken, err
}

func (s *UserSrvc) UserActivate(ctx context.Context, token string) error {

	err := s.repo.UnitOfWork.Do(ctx, func(ctx context.Context, repos *store.Storage) error {

		hash := sha256.Sum256([]byte(token))
		hashedToken := hex.EncodeToString(hash[:])

		user, err := repos.UserStore.GetUserByToken(ctx, hashedToken)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := repos.UserStore.Update(ctx, user); err != nil {
			return err
		}

		if err := repos.UserStore.Clean(ctx, hashedToken); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return err

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
			CreatedAt: u.CreatedAt,
		})
	}

	return users, nil

}
