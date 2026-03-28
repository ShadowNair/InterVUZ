package config

import "os"

type Config struct {
	HTTPPort      string
	RootImagePath string
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

	return Config{
		HTTPPort:      port,
		RootImagePath: rootImagePath,
	}
}
