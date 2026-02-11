package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgres(Addr, Host, User, Password, Name string, MaxIdleConns, MaxOpenConns int, MaxIdletime string) (*sql.DB, error) {
	hostPort := Host
	if Addr != "" {
		hostPort = net.JoinHostPort(Host, Addr)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", User, Password, hostPort, Name)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetMaxOpenConns(MaxOpenConns)
	maxTime, err := time.ParseDuration(MaxIdletime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(maxTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
