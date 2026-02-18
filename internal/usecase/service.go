package service

import (
	"chatX/internal/store"
	"context"
	"time"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type Services struct {
	UserSrvc interface {
		RegisterUser(ctx context.Context, user RequestRegister, exp time.Duration) (string, error)
		UserActivate(ctx context.Context, token string) error
		GetUsers(ctx context.Context, userID int, pg *store.PaginationQuery) ([]User, error)
	}

	ChatSRVC interface {
		CreatePrivateChat(ctx context.Context, senderID int64, receiverID int64) (int64, error)
		CreateGroupChat(ctx context.Context, group *Group) (int64, error)
		GetUserChats(ctx context.Context, userID int64, searchTerm string) ([]*ChatInfo, error)
		Updatechat(ctx context.Context, group *Chatgroup) (*store.Group, error)
		DeleteChat(ctx context.Context, chatID int) error
	}

	MemberSRV interface {
		GetByChatID(ctx context.Context, chatID int) ([]store.User, error)
		IsMember(ctx context.Context, chatID int64, userID int64) (bool, error)
		Add(ctx context.Context, actorUserID int64, chatID int, userID int) error
		Delete(ctx context.Context, actorUserID int64, chatID int, userID int) error
	}

	MessageSRV interface {
		Create(ctx context.Context, msg Message) (*Message, error)
		GetByID(ctx context.Context, id int64) (*Message, error)
		GetByChatID(ctx context.Context, chatID int64) ([]MessageDetail, error)
		MarkChatAsRead(ctx context.Context, chatID, userID int64) error
		UpdateMessage(ctx context.Context, msgID, userID int64, newText string) error
		DeleteMessage(ctx context.Context, msgID, userID int64) error
	}
}

func NewServices(repo *store.Storage) *Services {
	return &Services{
		UserSrvc:   &UserSrvc{repo},
		ChatSRVC:   &ChatSRVC{repo},
		MemberSRV:  &MemberSRV{repo},
		MessageSRV: &MessageSRV{repo},
	}
}
