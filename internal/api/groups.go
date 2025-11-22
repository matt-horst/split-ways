package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/matt-horst/split-ways/internal/database"
)

func CreateGroup(ctx context.Context, db *sql.DB, tx *sql.Tx, queries *database.Queries, params database.CreateGroupParams) (group database.Group, err error) {
	commit := false
	if tx == nil {
		tx, err = db.Begin()
		if err != nil {
			return
		}
		defer tx.Rollback()

		queries = queries.WithTx(tx)

		commit = true
	}

	group, err = queries.CreateGroup(ctx, params)
	if err != nil {
		return
	}

	_, err = queries.CreateUserGroup(ctx, database.CreateUserGroupParams{
		UserID:  group.Owner,
		GroupID: group.ID,
	})
	if err != nil {
		return
	}

	if commit {
		err = tx.Commit()
	}

	return
}

func IsUserInGroup(ctx context.Context, queries *database.Queries, userID, groupID uuid.UUID) (bool, error) {
	if _, err := queries.GetUserGroup(ctx, database.GetUserGroupParams{
		GroupID: groupID,
		UserID:  userID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
