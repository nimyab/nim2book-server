package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env  string `env:"ENV" envDefault:"dev"`
	Port string `env:"PORT" envDefault:":5050"`

	YandexDictionaryKey string `env:"YANDEX_DICTIONARY_KEY"`
	YandexDictionaryURL string `env:"YANDEX_DICTIONARY_URL"`

	LibreTranslateURL string `env:"LIBRE_TRANSLATE_URL"`

	WordAlignerURL string `env:"WORD_ALIGNER_URL"`

	RedisURL string `env:"REDIS_URL"`

	PostgresURL string `env:"POSTGRES_URL"`

	S3RootUser     string `env:"S3_ROOT_USER"`
	S3RootPassword string `env:"S3_ROOT_PASSWORD"`
	S3URL          string `env:"S3_URL"`
	S3BucketName   string `env:"S3_BUCKET_NAME"`
	S3Region       string `env:"S3_REGION"`

	JWTSecret    string `env:"JWT_SECRET"`
	JWTExpiresIn string `env:"JWT_EXPIRES_IN"`
}

var appConfig *Config

func init() {
	appConfig = &Config{
		Env:                 os.Getenv("ENV"),
		Port:                os.Getenv("PORT"),
		YandexDictionaryKey: os.Getenv("YANDEX_DICTIONARY_KEY"),
		YandexDictionaryURL: os.Getenv("YANDEX_DICTIONARY_URL"),
		LibreTranslateURL:   os.Getenv("LIBRE_TRANSLATE_URL"),
		WordAlignerURL:      os.Getenv("WORD_ALIGNER_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		PostgresURL:         os.Getenv("POSTGRES_URL"),
		S3RootUser:          os.Getenv("S3_ROOT_USER"),
		S3RootPassword:      os.Getenv("S3_ROOT_PASSWORD"),
		S3URL:               os.Getenv("S3_URL"),
		S3BucketName:        os.Getenv("S3_BUCKET_NAME"),
		S3Region:            os.Getenv("S3_REGION"),
		JWTExpiresIn:        os.Getenv("JWT_EXPIRING_IN"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
	}
}

func GetConfig() *Config {
	return appConfig
}
