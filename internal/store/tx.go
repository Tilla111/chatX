package store

import (
	"context"
	"database/sql"
)

type SQLUnitOfWork struct {
	db *sql.DB
}

func (u *SQLUnitOfWork) Do(
	ctx context.Context,
	fn func(ctx context.Context, repos *Storage) error,
) error {

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	repos := &Storage{
		UserStore:      &UserStore{tx},
		Chatstorage:    &Chatstorage{tx},
		MemberStorage:  &MemberStorage{tx},
		Groupstorage:   &Groupstorage{tx},
		MessageStorage: &MessageStorage{tx},
	}

	if err := fn(ctx, repos); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
