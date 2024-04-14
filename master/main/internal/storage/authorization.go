package sqlite

import (
	"context"
	"database/sql"
)

func (s *Storage) CheckToken(token string, ctx context.Context) (role string, err error) {
	row := s.Db.QueryRowContext(ctx, "SELECT Role FROM Users WHERE Token = :token",
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
