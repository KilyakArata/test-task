package sqlite

import (
	"context"
	"database/sql"
)

func (s *Storage) CheckToken(token string, ctx context.Context) (role string, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return
			}
			return
		}
		_ = tx.Commit()
	}()

	row := tx.QueryRowContext(ctx, "SELECT Role FROM Users WHERE Token = :token",
		sql.Named("token", token))

	var roleFromDb bool

	err = row.Scan(&roleFromDb)
	if err != nil {
		return "", err
	}

	if !roleFromDb {
		return "User", nil
	}

	return "Admin", nil
}
