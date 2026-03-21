package config

import "os"

type Config struct {
	HTTPPort      string
	RootImagePath string
}

func Load() Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	rootImagePath := os.Getenv("ROOT_IMAGE_PATH")
	if rootImagePath == "" {
		rootImagePath = "root-image.png"
	}

	return Config{
		HTTPPort:      port,
		RootImagePath: rootImagePath,
	}
}
