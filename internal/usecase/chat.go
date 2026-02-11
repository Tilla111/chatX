package service

import (
	"chatX/internal/store"
	"context"
	"database/sql"
	"errors"
)

type ChatSRVC struct {
	repo *store.Storage
}

type ChatReq struct {
	UserID   int64
	MemberID int64
	ChatType string
}

func (s *ChatSRVC) CreatePrivateChat(ctx context.Context, senderID int64, receiverID int64) (int64, error) {

	if senderID == receiverID {
		return 0, sql.ErrNoRows
	}

	req := store.Chatcheck{
		UserID:   senderID,
		MemberID: receiverID,
	}

	existingChatID, err := s.repo.Chatstorage.CheckChatP(ctx, &req)
	if err == nil && existingChatID != 0 {
		return existingChatID, nil
	}
	if err != nil && !errors.Is(err, store.SqlNotfound) {
		return 0, err
	}

	var newChatID int64
	err = s.repo.UnitOfWork.Do(ctx, func(ctx context.Context, repos *store.Storage) error {

		chat := &ChatReq{ChatType: "private"}
		req := &store.Chat{ChatType: chat.ChatType}
		ID, err := repos.Chatstorage.Createchat(ctx, req)
		if err != nil {
			return err
		}

		newChatID = ID

		for _, uid := range []int64{senderID, receiverID} {
			err := repos.MemberStorage.AddMember(ctx, &store.Member{
				ChatID: newChatID,
				UserID: uid,
				Rol:    RoleMember,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return newChatID, err
}

type Group struct {
	SenderID    int64
	ReceiverID  []int64
	Name        string
	Description string
}

func (s *ChatSRVC) CreateGroupChat(ctx context.Context, group *Group) (int64, error) {
	var newChatID int64

	if len(group.ReceiverID) == 0 {
		return 0, sql.ErrNoRows
	}
	for _, id := range group.ReceiverID {
		if id == group.SenderID {
			return 0, sql.ErrNoRows
		}
	}

	err := s.repo.UnitOfWork.Do(ctx, func(ctx context.Context, repos *store.Storage) error {

		chat := &ChatReq{ChatType: "group"}
		req := &store.Chat{ChatType: chat.ChatType}
		ID, err := repos.Chatstorage.Createchat(ctx, req)
		if err != nil {
			return err
		}

		newChatID = ID

		// add sender as admin
		if err := repos.MemberStorage.AddMember(ctx, &store.Member{
			ChatID: newChatID,
			UserID: group.SenderID,
			Rol:    RoleOwner,
		}); err != nil {
			return err
		}

		// add receivers
		for _, uid := range group.ReceiverID {
			if err := repos.MemberStorage.AddMember(ctx, &store.Member{
				ChatID: newChatID,
				UserID: uid,
				Rol:    RoleMember,
			}); err != nil {
				return err
			}
		}

		if err := repos.Groupstorage.CreateGroup(ctx, &store.Group{
			ChatID:      newChatID,
			GroupName:   group.Name,
			Description: group.Description,
		}); err != nil {
			return err
		}

		return nil
	})

	return newChatID, err
}

type ChatInfo struct {
	ChatID        int64  `json:"chat_id"`
	ChatType      string `json:"chat_type"`
	ChatName      string `json:"chat_name"`
	UserRole      string `json:"user_role"`
	JoinedAt      string `json:"joined_at"`
	LastMessage   string `json:"last_message"`
	LastMessageAt string `json:"last_message_at"`
	UnreadCount   int    `json:"unread_count"` // Yangi qo'shildi
}

func (s *ChatSRVC) GetUserChats(ctx context.Context, userID int64, searchTerm string) ([]*ChatInfo, error) {
	chats, err := s.repo.Chatstorage.GetChatByUserID(ctx, userID, searchTerm)
	if err != nil {
		return nil, err
	}

	if chats == nil {
		return []*ChatInfo{}, nil
	}

	result := make([]*ChatInfo, len(chats))
	for i, chat := range chats {
		result[i] = (*ChatInfo)(chat)
	}
	return result, nil
}

type Chatgroup struct {
	ChatID      int
	GroupName   string
	Description string
}

func (s *ChatSRVC) Updatechat(ctx context.Context, group *Chatgroup) (*store.Group, error) {

	chat, err := s.repo.Groupstorage.Update(ctx, &store.Group{
		ChatID:      int64(group.ChatID),
		GroupName:   group.GroupName,
		Description: group.Description,
	})
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *ChatSRVC) DeleteChat(ctx context.Context, chatID int) error {

	if err := s.repo.Chatstorage.Delete(ctx, chatID); err != nil {
		return err
	}

	return nil
}
