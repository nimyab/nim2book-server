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

	MaxRequestCount  int           `env:"MAX_REQUEST_COUNT"`
	WaitMilliseconds time.Duration `env:"WAIT_MILLISECONDS"`

	LibreTranslateURL string `env:"LIBRE_TRANSLATE_URL"`

	WordAlignerURLRest  string `env:"WORD_ALIGNER_URL_REST"`
	WordAlignerAddrGrpc string `env:"WORD_ALIGNER_ADDR_GRPC"`

	RedisURL string `env:"REDIS_URL"`

	PostgresURL string `env:"POSTGRES_URL"`

	MinioRootUser     string `env:"MINIO_ROOT_USER"`
	MinioRootPassword string `env:"MINIO_ROOT_PASSWORD"`
	MinioURL          string `env:"MINIO_URL"`
	MinioBucketName   string `env:"MINIO_BUCKET_NAME"`
	MinioRegion       string `env:"MINIO_REGION"`
	MinioUseSSL       bool   `env:"MINIO_USE_SSL"`

	JWTSecret      string        `env:"JWT_SECRET"`
	JWTAccessTime  time.Duration `env:"JWT_ACCESS_TIME" envDefault:"15"`
	JWTRefreshTime time.Duration `env:"JWT_REFRESH_TIME" envDefault:"30"`

	GoogleClientId    string `env:"GOOGLE_CLIENT_ID"`
	GoogleCredentials string `env:"GOOGLE_CREDENTIALS"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	maxRequestCount, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_COUNT"))
	jwtAccessTime, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TIME"))
	jwtRefreshTime, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TIME"))
	waitMilliseconds, _ := strconv.Atoi(os.Getenv("WAIT_MILLISECONDS"))

	cfg := &Config{
		Env:                 getEnvOrDefault("ENV", EnvDev),
		Port:                getEnvOrDefault("PORT", ":5050"),
		YandexDictionaryKey: os.Getenv("YANDEX_DICTIONARY_KEY"),
		YandexDictionaryURL: os.Getenv("YANDEX_DICTIONARY_URL"),
		MaxRequestCount:     maxRequestCount,
		WaitMilliseconds:    time.Duration(waitMilliseconds) * time.Millisecond,
		LibreTranslateURL:   os.Getenv("LIBRE_TRANSLATE_URL"),
		WordAlignerURLRest:  os.Getenv("WORD_ALIGNER_URL"),
		WordAlignerAddrGrpc: os.Getenv("WORD_ALIGNER_ADDR_GRPC"),
		RedisURL:            os.Getenv("REDIS_URL"),
		PostgresURL:         os.Getenv("POSTGRES_URL"),
		MinioRootUser:       os.Getenv("MINIO_ROOT_USER"),
		MinioRootPassword:   os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioURL:            os.Getenv("MINIO_URL"),
		MinioBucketName:     os.Getenv("MINIO_BUCKET_NAME"),
		MinioRegion:         os.Getenv("MINIO_REGION"),
		MinioUseSSL:         os.Getenv("MINIO_USE_SSL") == "true",
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTAccessTime:       time.Duration(jwtAccessTime) * time.Minute,
		JWTRefreshTime:      time.Duration(jwtRefreshTime) * time.Hour * 24,
		GoogleClientId:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleCredentials:   os.Getenv("GOOGLE_CREDENTIALS"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
