package sqlite

import (
	"context"
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite"
)

type Storage struct {
	Db *sql.DB
}

func New(storagePath string, log *slog.Logger, ctx context.Context) (*Storage, error) {
	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		log.Error("failed to open storage:", err)
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS banners (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feature_id INTEGER,
		content TEXT,
		is_active BOOLEAN,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Error("Ошибка во время создания таблицы banners:", err)
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS banner_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		banner_id INTEGER,
		tag_id INTEGER,
		feature_id INTEGER,
		FOREIGN KEY(banner_id) REFERENCES banners(id),
        UNIQUE(tag_id, feature_id)
	)`)
	if err != nil {
		log.Error("Ошибка во время создания таблицы banner_tags: ", err)
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS banner_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		banner_id INTEGER,
		feature_id INTEGER,
		content TEXT,
		is_active BOOLEAN,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(banner_id) REFERENCES banners(id)
	)`)
	if err != nil {
		log.Error("Ошибка во время создания таблицы banners:", err)
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS banner_versions_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		banner_version_id INTEGER,
		banner_id INTEGER,
		tag_id INTEGER,
		feature_id INTEGER,
		FOREIGN KEY(banner_version_id) REFERENCES banner_versions(id)
	)`)
	if err != nil {
		log.Error("Ошибка во время создания таблицы banner_tags: ", err)
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS Users (
        Role BOOLEAN,
        Token TEXT PRIMARY KEY
    )`)
	if err != nil {
		log.Error("Ошибка во время создания таблицы Users: ", err)
		return nil, err
	}

	rows, err := db.QueryContext(ctx, "SELECT COUNT(*) FROM Users")
	if err != nil {
		log.Error("Ошибка во время создания таблицы Users:", err)
		return nil, err
	}

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Ошибка во время создания таблицы Users:", err)
			return nil, err
		}
	}
	rows.Close()

	if count == 0 {
		_, err = db.ExecContext(ctx, `INSERT INTO Users (Role, Token) VALUES ('true', 'c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f')`)
		if err != nil {
			log.Error("Ошибка во время добавления роли Admin в таблицу Users: ", err)
			return nil, err
		}

		_, err = db.ExecContext(ctx, `INSERT INTO Users (Role, Token) VALUES ('false', 'b512d97e7cbf97c273e4db073bbb547aa65a84589227f8f3d9e4a72b9372a24d')`)
		if err != nil {
			log.Error("Ошибка во время добавления роли User в таблицу Users: ", err)
			return nil, err
		}
	}

	return &Storage{Db: db}, nil
}
