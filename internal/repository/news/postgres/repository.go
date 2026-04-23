package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveMany(ctx context.Context, items []domain.NewsRecord) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, item := range items {
		newsID, err := upsertNews(ctx, tx, item)
		if err != nil {
			return fmt.Errorf("upsert news %s: %w", item.Slug, err)
		}

		for _, tag := range item.Tags {
			if err := upsertTag(ctx, tx, tag); err != nil {
				return fmt.Errorf("upsert tag %d for news %s: %w", tag.ID, item.Slug, err)
			}
		}

		if err := replaceNewsTags(ctx, tx, newsID, item.Tags); err != nil {
			return fmt.Errorf("replace news_tags for news %s: %w", item.Slug, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func upsertNews(ctx context.Context, tx *sql.Tx, item domain.NewsRecord) (int64, error) {
	const query = `
		INSERT INTO news (
			slug,
			title,
			preview_text,
			published_date,
			image_preview,
			page_url
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			preview_text = EXCLUDED.preview_text,
			published_date = EXCLUDED.published_date,
			image_preview = EXCLUDED.image_preview,
			page_url = EXCLUDED.page_url
		RETURNING id
	`

	var id int64
	err := tx.QueryRowContext(
		ctx,
		query,
		item.Slug,
		item.Title,
		nullString(item.PreviewText),
		item.PublishedDate,
		nullString(item.ImagePreview),
		item.PageURL,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func upsertTag(ctx context.Context, tx *sql.Tx, tag domain.Tag) error {
	const query = `
		INSERT INTO tags (id, slug, title, color)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			title = EXCLUDED.title,
			color = EXCLUDED.color
	`

	_, err := tx.ExecContext(ctx, query, tag.ID, tag.Slug, tag.Title, tag.Color)
	return err
}

func replaceNewsTags(ctx context.Context, tx *sql.Tx, newsID int64, tags []domain.Tag) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM news_tags WHERE news_id = $1`, newsID); err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	const query = `
		INSERT INTO news_tags (news_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT (news_id, tag_id) DO NOTHING
	`

	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, query, newsID, tag.ID); err != nil {
			return err
		}
	}

	return nil
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}