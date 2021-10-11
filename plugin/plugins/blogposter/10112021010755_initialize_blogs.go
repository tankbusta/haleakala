package blogposter

import (
	"context"
	"time"

	"github.com/tankbusta/haleakala/database"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

func init() {
	database.Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		log.Warn().Msg("Blog migration starting...")
		type BlogPost struct {
			URL        string
			Title      string
			TimeToRead int
			PostedOn   time.Time
			IndexedOn  time.Time
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err := tx.NewCreateTable().
			Model((*BlogPost)(nil)).
			IfNotExists().
			Exec(ctx); err != nil {
			return err
		}

		node, err := getBlogContent()
		if err != nil {
			return err
		}

		blogs, err := GetMandiantBlogs(node)
		if err != nil {
			return err
		}

		log.Warn().Msgf("Discovered %d blog posts, storing into database", len(blogs))
		res, err := tx.NewInsert().Model(&blogs).Exec(ctx)
		if err != nil {
			return err
		}

		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}

		log.Warn().Msgf("Indexed %d blog posts", ra)
		return tx.Commit()
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewDropTable().Model((*BlogPost)(nil)).Exec(ctx)
		return err
	})
}
