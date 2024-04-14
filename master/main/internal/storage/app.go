package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Query struct {
	TagId     int  `json:"tag_id,omitempty"`
	FeatureId int  `json:"feature_id,omitempty"`
	Revision  bool `json:"use_last_revision,omitempty"`
	Limit     int  `json:"limit,omitempty"`
	Offset    int  `json:"offset,omitempty"`
}
type Banner struct {
	BannerId  int               `json:"banner_id,omitempty"`
	TagIds    []int             `json:"tag_ids"`
	FeatureId int               `json:"feature_id"`
	Content   map[string]string `json:"content"`
	IsActive  bool              `json:"is_active"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}
type BannerUpdate struct {
	BannerId  int
	TagIds    []int             `json:"tag_ids,omitempty"`
	FeatureId int               `json:"feature_id,omitempty"`
	Content   map[string]string `json:"content,omitempty"`
	IsActive  bool              `json:"is_active,omitempty"`
}

func (s *Storage) GetBannerFromStorage(query Query, ctx context.Context) (content string, active bool, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return "", false, err
	}

	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			return
		}
		_ = tx.Commit()
	}()
	row := s.Db.QueryRowContext(ctx, `SELECT b.content, b.is_active FROM banners b
		INNER JOIN banner_tags bt ON b.id = bt.banner_id
		WHERE bt.tag_id = :tagId AND b.feature_id = :featureId`,
		sql.Named("tagId", query.TagId),
		sql.Named("featureId", query.FeatureId))

	var storedData string
	var isActive bool

	err = row.Scan(&storedData, &isActive)
	if err != nil {
		return "", false, err
	}

	return storedData, isActive, nil
}

func (s *Storage) GetAllBannersFromStorage(query Query, ctx context.Context) (banners []Banner, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
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

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`SELECT b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at FROM banners b`)
	if query.TagId != 0 {
		queryBuilder.WriteString(` JOIN banner_tags bt ON b.id = bt.banner_id WHERE bt.tag_id = :tagId`)
	}

	if query.FeatureId != 0 {
		if query.TagId == 0 {
			queryBuilder.WriteString(` WHERE`)
		} else {
			queryBuilder.WriteString(` AND`)
		}
		queryBuilder.WriteString(` b.feature_id = :featureId`)
	}

	if query.Limit != 0 {
		queryBuilder.WriteString(` LIMIT :limit`)
	}

	if query.Offset != 0 {
		queryBuilder.WriteString(` OFFSET :offset`)
	}

	rows, err := s.Db.QueryContext(ctx, queryBuilder.String(),
		sql.Named("featureId", query.FeatureId),
		sql.Named("tagId", query.TagId),
		sql.Named("limit", query.Limit),
		sql.Named("offset", query.Offset))

	var res []Banner
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var banner Banner
		var contentJSON string
		err := rows.Scan(&banner.BannerId, &banner.FeatureId, &contentJSON, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(contentJSON), &banner.Content)
		if err != nil {
			return nil, err
		}

		var tags []int
		rowes, err := s.Db.QueryContext(ctx, "SELECT tag_id FROM banner_tags WHERE banner_id = :bannerId",
			sql.Named("bannerId", banner.BannerId))
		if err != nil {
			return nil, err
		}

		for rowes.Next() {
			var tagID int
			if err := rowes.Scan(&tagID); err != nil {
				return nil, err
			}
			tags = append(tags, tagID)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		rowes.Close()

		banner.TagIds = tags

		res = append(res, banner)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Storage) PostBannerToStorage(banner Banner, ctx context.Context) (id int, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return 0, err
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

	contentJSON, err := json.Marshal(banner.Content)
	if err != nil {
		return 0, err
	}

	result, err := tx.ExecContext(ctx, `INSERT INTO banners (feature_id, content, is_active) 
		VALUES (:featureId, :content, :isActive)`,
		sql.Named("featureId", banner.FeatureId),
		sql.Named("content", string(contentJSON)),
		sql.Named("isActive", banner.IsActive))
	if err != nil {
		return 0, err
	}

	idLast, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, tagID := range banner.TagIds {
		_, err := tx.ExecContext(ctx, `INSERT INTO banner_tags (banner_id, tag_id, feature_id)
			VALUES (:bannerId, :tagId, :featureId)`,
			sql.Named("bannerId", idLast),
			sql.Named("featureId", banner.FeatureId),
			sql.Named("tagId", tagID))
		if err != nil {
			return 0, err
		}
	}

	return int(idLast), nil
}

func (s *Storage) UpdateBannerInStorage(banner BannerUpdate, ctx context.Context) (err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return err
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

	contentJSON, err := json.Marshal(banner.Content)
	if err != nil {
		return err
	}

	var versionCount int
	err = s.Db.QueryRowContext(ctx, `SELECT COUNT(*) FROM banner_versions WHERE banner_id = :bannerId`, sql.Named("bannerId", banner.BannerId)).
		Scan(&versionCount)
	if err != nil {
		return err
	}

	if versionCount == 3 {
		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions WHERE id = 
            	(SELECT id FROM banner_versions WHERE banner_id = :bannerId ORDER BY id ASC LIMIT 1)`,
			sql.Named("bannerId", banner.BannerId))
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions_tags WHERE banner_id =
            	(SELECT id FROM banner_versions WHERE banner_id = :bannerId ORDER BY id ASC LIMIT 1)`,
			sql.Named("bannerId", banner.BannerId))
		if err != nil {
			return err
		}
	}

	var oldBanner Banner
	err = s.Db.QueryRowContext(ctx, `SELECT * FROM banners WHERE id = :bannerId`, sql.Named("bannerId", banner.BannerId)).
		Scan(&oldBanner.BannerId, &oldBanner.FeatureId, &oldBanner.Content, &oldBanner.IsActive, &oldBanner.CreatedAt, &oldBanner.UpdatedAt)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `INSERT INTO banner_versions (banner_id, feature_id, content, is_active, created_at, updated_at) 
		VALUES (:bannerId, :featureId, :content, :isActive, :created_at, :updated_at)`,
		sql.Named("bannerId", banner.BannerId),
		sql.Named("featureId", oldBanner.FeatureId),
		sql.Named("content", oldBanner.Content),
		sql.Named("isActive", oldBanner.IsActive),
		sql.Named("createdAt", oldBanner.CreatedAt),
		sql.Named("updatedAt", oldBanner.UpdatedAt))
	if err != nil {
		return err
	}

	idLast, err := result.LastInsertId()
	if err != nil {
		return err
	}

	rows, err := s.Db.QueryContext(ctx, `SELECT tag_id FROM banner_tags WHERE banner_id = :bannerId`, sql.Named("bannerId", banner.BannerId))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tagID int
		err := rows.Scan(&tagID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO banner_versions_tags (banner_version_id,banner_id, tag_id, feature_id) 
			VALUES (:bannerVersionId, :bannerId, :tagId, :featureId)`,
			sql.Named("bannerVersionId", idLast),
			sql.Named("bannerId", banner.BannerId),
			sql.Named("tagId", tagID),
			sql.Named("featureId", oldBanner.FeatureId))
		if err != nil {
			return err
		}
	}

	_, err = s.Db.ExecContext(ctx, `UPDATE banners
		SET feature_id = COALESCE(NULLIF(:featureId, 0), feature_id),
    		content = COALESCE(NULLIF(:content, '{}'), content),
    		is_active = COALESCE(NULLIF(:bannerId, 0), is_active),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :bannerId`,
		sql.Named("featureId", banner.FeatureId),
		sql.Named("content", contentJSON),
		sql.Named("isActive", banner.IsActive),
		sql.Named("bannerId", banner.BannerId))

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_tags WHERE banner_id = :bannerId`,
		sql.Named("bannerId", banner.BannerId))
	if err != nil {
		return err
	}

	for _, tagID := range banner.TagIds {
		_, err := tx.ExecContext(ctx, `INSERT INTO banner_tags (banner_id, tag_id, feature_id) VALUES (:bannerId, :tagId, :featureId)`,
			sql.Named("bannerId", banner.BannerId),
			sql.Named("featureId", banner.FeatureId),
			sql.Named("tagId", tagID))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) DeleteBannerFromStorage(id int, ctx context.Context) (keys []string, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
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

	deleteKeys := []string{}

	rowas, err := tx.QueryContext(ctx, `SELECT tag_id,feature_id FROM banner_tags WHERE banner_id = :bannerId`, sql.Named("bannerId", id))
	if err != nil {
		return nil, err
	}

	for rowas.Next() {
		var tagID, featureID int
		err := rowas.Scan(&tagID, &featureID)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%d %d", featureID, tagID)
		deleteKeys = append(deleteKeys, key)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions WHERE banner_id = :bannerId`, sql.Named("bannerId", id))
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions_tags WHERE banner_id = :bannerId`, sql.Named("bannerId", id))
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banners WHERE id = :bannerId`, sql.Named("bannerId", id))
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_tags WHERE banner_id = :bannerId`, sql.Named("bannerId", id))
	if err != nil {
		return nil, err
	}

	return deleteKeys, nil
}

func (s *Storage) DeleteBannerFromStorageByFeature(featureId int, ctx context.Context) (keys []string, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
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

	rows, err := tx.QueryContext(ctx, `SELECT id FROM banners WHERE feature_id = :featureId`, sql.Named("featureId", featureId))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var bannerID int
		err := rows.Scan(&bannerID)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions WHERE banner_id = :bannerID`, sql.Named("bannerID", bannerID))
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions_tags WHERE banner_id = :bannerID`, sql.Named("bannerID", bannerID))
		if err != nil {
			return nil, err
		}
	}
	rows.Close()

	deleteKeys := []string{}

	rowys, err := s.Db.QueryContext(ctx, `SELECT tag_id FROM banner_tags WHERE feature_id = :featureId`, sql.Named("featureId", featureId))
	if err != nil {
		return nil, err
	}

	for rowys.Next() {
		var tagID int
		err := rows.Scan(&tagID)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%d %d", featureId, tagID)
		deleteKeys = append(deleteKeys, key)
	}
	rowys.Close()

	_, err = tx.ExecContext(ctx, `DELETE FROM banners WHERE feature_id = :featureId`, sql.Named("featureId", featureId))
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_tags WHERE feature_id = :featureId`, sql.Named("featureId", featureId))
	if err != nil {
		return nil, err
	}

	return deleteKeys, nil
}

func (s *Storage) DeleteBannerFromStorageByTag(tag int, ctx context.Context) (keys []string, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
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

	var deleteKeys []string

	rows, err := tx.QueryContext(ctx, `
		SELECT banners.id
		FROM banners
		JOIN banner_tags ON banners.id = banner_tags.banner_id
		WHERE banner_tags.tag_id = :tagId`, sql.Named("tagId", tag))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var bannerID int
		err := rows.Scan(&bannerID)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions WHERE banner_id = :bannerID`, sql.Named("bannerID", bannerID))
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM banner_versions_tags WHERE banner_id = :bannerID`, sql.Named("bannerID", bannerID))
		if err != nil {
			return nil, err
		}
	}
	rows.Close()

	rowys, err := s.Db.QueryContext(ctx, `SELECT feature_id FROM banner_tags WHERE tag_id = :tagId`, sql.Named("tagId", tag))
	if err != nil {
		return nil, err
	}

	for rowys.Next() {
		var featureID int
		err := rows.Scan(&featureID)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%d %d", featureID, tag)
		deleteKeys = append(deleteKeys, key)
	}
	rowys.Close()

	_, err = tx.ExecContext(ctx, `DELETE FROM banners
		WHERE id IN (
			SELECT banners.id
			FROM banners
			JOIN banner_tags ON banners.id = banner_tags.banner_id
			WHERE banner_tags.tag_id = :tagId);`,
		sql.Named("tagId", tag))
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM banner_tags WHERE tag_id = :tagId`, sql.Named("tagId", tag))
	if err != nil {
		return nil, err
	}

	return deleteKeys, nil
}

func (s *Storage) GetBannerVersionsFromStorage(id int, ctx context.Context) (banners []Banner, err error) {
	tx, err := s.Db.Begin()
	if err != nil {
		return nil, err
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

	rows, err := s.Db.QueryContext(ctx, "SELECT id, feature_id, content, is_active, created_at, updated_at FROM banner_versions WHERE banner_id = :bannerId",
		sql.Named("bannerId", id))

	var res []Banner
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var banner Banner
		var contentJSON string
		err := rows.Scan(&banner.BannerId, &banner.FeatureId, &contentJSON, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(contentJSON), &banner.Content)
		if err != nil {
			return nil, err
		}

		var tags []int
		rowes, err := s.Db.QueryContext(ctx, "SELECT tag_id FROM banner_versions_tags WHERE banner_version_id = :bannerId",
			sql.Named("bannerId", banner.BannerId))
		if err != nil {
			return nil, err
		}

		for rowes.Next() {
			var tagID int
			if err := rowes.Scan(&tagID); err != nil {
				return nil, err
			}
			tags = append(tags, tagID)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		rowes.Close()

		banner.TagIds = tags

		res = append(res, banner)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
