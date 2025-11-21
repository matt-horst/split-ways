package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/matt-horst/split-ways/internal/database"
)

func Reset(ctx context.Context, db *sql.DB, queries *database.Queries) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("couldn't begin transaction: %v", err)
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	if err := qtx.DeleteAllGroups(ctx); err != nil {
		return fmt.Errorf("couldn't delete all groups: %v", err)
	}

	if err := qtx.DeleteAllUsers(ctx); err != nil {
		return fmt.Errorf("couldn't delete all users: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("couldn't commit transaction: %v", err)
	}

	return nil
}
