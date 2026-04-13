package config

import "os"

type Config struct {
	HTTPPort        string
	RootImagePath   string
	PlacesDataPath  string
	ScheduleDataDir string
	GraphJSONPath   string
	GraphSVGPath    string
}

func Load() Config {
	return Config{
		HTTPPort:        envOrDefault("HTTP_PORT", "8000"),
		RootImagePath:   envOrDefault("ROOT_IMAGE_PATH", "image_floor/Карта МГТУ-1-10-4_page-0001.jpg"),
		PlacesDataPath:  envOrDefault("PLACES_DATA_PATH", "GEO_Json/points.json"),
		ScheduleDataDir: envOrDefault("SCHEDULE_DATA_DIR", "schedule/lksJSON"),
		GraphJSONPath:   envOrDefault("GRAPH_JSON_PATH", "GEO_Json/ulk_first_graph.json"),
		GraphSVGPath:    envOrDefault("GRAPH_SVG_PATH", "GEO_Json/navigation_graph.svg"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
