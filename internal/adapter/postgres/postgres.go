package postgres

import (
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/nimyab/nim2book-back/ent"
)

type Config struct {
	PostgresURL string
	IsDebug     bool
}

type Postgres struct {
	Client *ent.Client
}

func New(config *Config) (*ent.Client, error) {
	db, err := sql.Open("postgres", config.PostgresURL)
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// Создаем драйвер ent
	drv := entsql.OpenDB(dialect.Postgres, db)

	opts := []ent.Option{ent.Driver(drv)}
	if config.IsDebug {
		opts = append(opts, ent.Debug())
	}

	client := ent.NewClient(opts...)
	return client, nil
}
