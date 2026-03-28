package config

import "os"

type Config struct {
	HTTPPort           string
	RootImagePath      string
	DatabaseURL        string
	AccessTokenSecret  string
	RefreshTokenSecret string
}

func Load() Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8000"
	}

	rootImagePath := os.Getenv("ROOT_IMAGE_PATH")
	if rootImagePath == "" {
		rootImagePath = "image_floor/Карта МГТУ-1-10-4_page-0001.jpg"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")

	return Config{
		HTTPPort:           port,
		RootImagePath:      rootImagePath,
		DatabaseURL:        databaseURL,
		AccessTokenSecret:  accessTokenSecret,
		RefreshTokenSecret: refreshTokenSecret,
	}
}
