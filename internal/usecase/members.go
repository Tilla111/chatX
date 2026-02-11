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

func (s *MemberSRV) IsMember(ctx context.Context, chatID int64, userID int64) (bool, error) {
	return s.repo.MemberStorage.IsMember(ctx, chatID, userID)
}

func (s *MemberSRV) Delete(ctx context.Context, actorUserID int64, chatID int, userID int) error {
	if actorUserID != int64(userID) {
		role, err := s.repo.MemberStorage.GetRole(ctx, int64(chatID), actorUserID)
		if err != nil {
			return err
		}
		if role != RoleOwner && role != RoleAdmin {
			return store.SqlForbidden
		}
	}

	err := s.repo.MemberStorage.Delete(ctx, chatID, userID)
	if err != nil {
		return err
	}

	return nil
}
