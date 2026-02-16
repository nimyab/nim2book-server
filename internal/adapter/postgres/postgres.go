package postgres

import (
	"github.com/nimyab/nim2book-back/ent"
)

type Config struct {
	PostgresURL string
}

type Postgres struct {
	Client *ent.Client
}

func New(config *Config) (*ent.Client, error) {
	client, err := ent.Open("postgres", config.PostgresURL)
	if err != nil {
		return nil, err
	}
	return client, nil
}
