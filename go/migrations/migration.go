package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func Up(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}

func Down(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.DownContext(ctx, db, ".")
}
