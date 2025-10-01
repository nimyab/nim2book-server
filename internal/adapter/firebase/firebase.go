package firebase

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Config struct {
	GoogleCredentials string
}

func New(ctx context.Context, cfg *Config) (*firebase.App, error) {
	opt := option.WithCredentialsJSON([]byte(cfg.GoogleCredentials))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}
	return app, nil
}
