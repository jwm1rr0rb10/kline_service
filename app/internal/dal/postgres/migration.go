package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jwm1rr0rb10/go-errors"
	"github.com/pressly/goose/v3"

	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
	"github.com/jwm1rr0rb10/kline_service/app/internal/dal/postgres/migrations"
)

func RunMigrations(ctx context.Context, cfg *config.PostgresConfig) error {
	dsn := buildPostgresDSN(cfg)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping database before migrations")
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.EmbedMigrations,
	)
	if err != nil {
		return errors.Wrap(err, "create goose provider")
	}

	if _, err := provider.Up(ctx); err != nil {
		return errors.Wrap(err, "run goose migrations")
	}

	return nil
}

func buildPostgresDSN(cfg *config.PostgresConfig) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   "/" + strings.TrimPrefix(cfg.Database, "/"),
	}

	q := u.Query()
	if cfg.RequireSSL {
		q.Set("sslmode", "require")
	} else {
		q.Set("sslmode", "disable")
	}
	u.RawQuery = q.Encode()

	return u.String()
}
