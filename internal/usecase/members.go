package service

import (
	"chatX/internal/store"
	"context"
)

type MemberSRV struct {
	repo *store.Storage
}

func (s *MemberSRV) GetByChatID(ctx context.Context, chatID int) ([]store.User, error) {

	members, err := s.repo.MemberStorage.GetByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (s *MemberSRV) Delete(ctx context.Context, chatID, userID int) error {
	err := s.repo.MemberStorage.Delete(ctx, chatID, userID)
	if err != nil {
		return err
	}

	return nil
}
