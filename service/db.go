package service

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func connectDB() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgresql://user:password@db:5432/exchange_rate")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func migrateDB(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS subscriptions (
      id SERIAL PRIMARY KEY,
      email VARCHAR(255) NOT NULL UNIQUE
    );
  `)
	return err
}
