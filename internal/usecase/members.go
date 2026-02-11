package service

import (
	"chatX/internal/store"
	"context"
	"database/sql"
	"errors"
)

type MemberSRV struct {
	repo *store.Storage
}

var ErrInvalidMemberChatType = errors.New("members can only be added to group chats")
var ErrMemberAlreadyExists = errors.New("user is already a member of this chat")

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

func (s *MemberSRV) Add(ctx context.Context, actorUserID int64, chatID int, userID int) error {
	chat, err := s.repo.Chatstorage.GetByID(ctx, int64(chatID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.SqlNotfound
		}
		return err
	}

	if chat.ChatType != "group" {
		return ErrInvalidMemberChatType
	}

	role, err := s.repo.MemberStorage.GetRole(ctx, int64(chatID), actorUserID)
	if err != nil {
		if errors.Is(err, store.SqlNotfound) {
			return store.SqlForbidden
		}
		return err
	}
	if role != RoleOwner && role != RoleAdmin {
		return store.SqlForbidden
	}

	exists, err := s.repo.MemberStorage.IsMember(ctx, int64(chatID), int64(userID))
	if err != nil {
		return err
	}
	if exists {
		return ErrMemberAlreadyExists
	}

	return s.repo.MemberStorage.AddMember(ctx, &store.Member{
		ChatID: int64(chatID),
		UserID: int64(userID),
		Rol:    RoleMember,
	})
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
