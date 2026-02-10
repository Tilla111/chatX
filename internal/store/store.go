package store

import (
	"context"
	"database/sql"
	"errors"
)

var SqlNotfound = errors.New("Not found")

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Storage struct {
	UnitOfWork interface {
		Do(
			ctx context.Context,
			fn func(ctx context.Context, repos *Storage) error,
		) error
	}
	UserStore interface {
		CreateUser(ctx context.Context, User *User) error
		GetAll(ctx context.Context, currentUserID int, pg *PaginationQuery) ([]User, error)
	}

	Chatstorage interface {
		Createchat(ctx context.Context, chat *Chat) (int64, error)
		GetByID(ctx context.Context, ChatID int64) (*Chat, error)
		GetChatByUserID(ctx context.Context, id int64, searchTerm string) ([]*ChatInfo, error)
		CheckChatP(ctx context.Context, users *Chatcheck) (int64, error)
		Delete(ctx context.Context, chatID int) error
	}

	MemberStorage interface {
		AddMember(ctx context.Context, member *Member) error
		GetByChatID(ctx context.Context, ChatID int) ([]User, error)
		Delete(ctx context.Context, chatID, userID int) error
	}

	Groupstorage interface {
		CreateGroup(ctx context.Context, group *Group) error
		Update(ctx context.Context, group *Group) (*Group, error)
	}

	MessageStorage interface {
		Create(ctx context.Context, msg *Message) (Message, error)
		GetByID(ctx context.Context, id int64) (*Message, error)
		GetMessages(ctx context.Context, chatID int64) ([]MessageDetail, error)
		MarkAsRead(ctx context.Context, chatID, userID int64) error
		Update(ctx context.Context, msgID, userID int64, newText string) error
		Delete(ctx context.Context, msgID, userID int64) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		UnitOfWork:     &SQLUnitOfWork{db},
		UserStore:      &UserStore{db},
		Chatstorage:    &Chatstorage{db},
		MemberStorage:  &MemberStorage{db},
		Groupstorage:   &Groupstorage{db},
		MessageStorage: &MessageStorage{db},
	}
}
