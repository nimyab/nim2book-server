package config

import (
	"os"
	"strconv"
	"time"

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

	MaxRequestCount int `env:"MAX_REQUEST_COUNT" envDefault:"100"`

	LibreTranslateURL string `env:"LIBRE_TRANSLATE_URL"`

	WordAlignerURL string `env:"WORD_ALIGNER_URL"`

	RedisURL string `env:"REDIS_URL"`

	PostgresURL string `env:"POSTGRES_URL"`

	S3RootUser     string `env:"S3_ROOT_USER"`
	S3RootPassword string `env:"S3_ROOT_PASSWORD"`
	S3URL          string `env:"S3_URL"`
	S3BucketName   string `env:"S3_BUCKET_NAME"`
	S3Region       string `env:"S3_REGION"`

	JWTSecret      string        `env:"JWT_SECRET"`
	JWTAccessTime  time.Duration `env:"JWT_ACCESS_TIME" envDefault:"15"`
	JWTRefreshTime time.Duration `env:"JWT_REFRESH_TIME" envDefault:"30"`
}

var appConfig *Config

func init() {
	maxRequestCount, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_COUNT"))
	jwtAccessTime, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TIME"))
	jwtRefreshTime, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TIME"))

	appConfig = &Config{
		Env:                 os.Getenv("ENV"),
		Port:                os.Getenv("PORT"),
		YandexDictionaryKey: os.Getenv("YANDEX_DICTIONARY_KEY"),
		YandexDictionaryURL: os.Getenv("YANDEX_DICTIONARY_URL"),
		MaxRequestCount:     maxRequestCount,
		LibreTranslateURL:   os.Getenv("LIBRE_TRANSLATE_URL"),
		WordAlignerURL:      os.Getenv("WORD_ALIGNER_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		PostgresURL:         os.Getenv("POSTGRES_URL"),
		S3RootUser:          os.Getenv("S3_ROOT_USER"),
		S3RootPassword:      os.Getenv("S3_ROOT_PASSWORD"),
		S3URL:               os.Getenv("S3_URL"),
		S3BucketName:        os.Getenv("S3_BUCKET_NAME"),
		S3Region:            os.Getenv("S3_REGION"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTAccessTime:       time.Duration(jwtAccessTime) * time.Minute,
		JWTRefreshTime:      time.Duration(jwtRefreshTime) * time.Hour * 24,
	}
}

func GetConfig() *Config {
	return appConfig
}
