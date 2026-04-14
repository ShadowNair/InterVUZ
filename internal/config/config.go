package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort                string
	RootImagePath           string
	PlacesDataPath          string
	ScheduleDataDir         string
	StructureAPIURL         string
	ScheduleGroupAPIBaseURL string
	DatabaseDSN             string
	SyncOnStartup           bool
	SyncInterval            time.Duration
	StructureTargetRootUUID string
}

func Load() Config {
	return Config{
		HTTPPort:                envOrDefault("HTTP_PORT", "8000"),
		RootImagePath:           envOrDefault("ROOT_IMAGE_PATH", "image_floor/Карта МГТУ-1-10-4_page-0001.jpg"),
		PlacesDataPath:          envOrDefault("PLACES_DATA_PATH", "GEO_Json/points.json"),
		ScheduleDataDir:         envOrDefault("SCHEDULE_DATA_DIR", "schedule/lksJSON"),
		StructureAPIURL:         envOrDefault("STRUCTURE_API_URL", "https://lks.bmstu.ru/lks-back/api/v1/structure"),
		ScheduleGroupAPIBaseURL: envOrDefault("SCHEDULE_GROUP_API_BASE_URL", "https://lks.bmstu.ru/lks-back/api/v1/schedules/groups"),
		DatabaseDSN:             os.Getenv("DATABASE_DSN"),
		SyncOnStartup:           envBoolOrDefault("SYNC_ON_STARTUP", false),
		SyncInterval:            envDurationOrDefault("SYNC_INTERVAL", 72*time.Hour),
		StructureTargetRootUUID: envOrDefault("STRUCTURE_TARGET_ROOT_UUID", "8c1b7bb8-e690-11db-89c3-000cf1a7cbf0"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
