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
	DatabaseDSN             string
	AcademicWeek1StartDate  string
	StructureAPIURL         string
	ScheduleGroupAPIBaseURL string
	StructureTargetRootUUID string
	SyncOnStartup           bool
	SyncInterval            time.Duration

	NewsAPIURL       string
	NewsSyncEnabled  bool
	NewsSyncInterval time.Duration
	NewsSyncLimit    int

	AssistantBaseURL         string
	AssistantAPIKey          string
	AssistantModel           string
	AssistantTimeout         time.Duration
	AssistantContextMaxChars int
}

func Load() Config {
	return Config{
		HTTPPort:                getenv("HTTP_PORT", "8000"),
		RootImagePath:           getenv("ROOT_IMAGE_PATH", ""),
		PlacesDataPath:          getenv("PLACES_DATA_PATH", ""),
		ScheduleDataDir:         getenv("SCHEDULE_DATA_DIR", ""),
		DatabaseDSN:             getenv("DATABASE_DSN", ""),
		AcademicWeek1StartDate:  getenv("ACADEMIC_WEEK1_START_DATE", ""),
		StructureAPIURL:         getenv("STRUCTURE_API_URL", "https://lks.bmstu.ru/lks-back/api/v1/structure"),
		ScheduleGroupAPIBaseURL: getenv("SCHEDULE_GROUP_API_BASE_URL", "https://lks.bmstu.ru/lks-back/api/v1/schedules/groups"),
		StructureTargetRootUUID: getenv("STRUCTURE_TARGET_ROOT_UUID", "8c1b7bb8-e690-11db-89c3-000cf1a7cbf0"),
		SyncOnStartup:           getBoolEnv("SYNC_ON_STARTUP", false),
		SyncInterval:            getDurationEnv("SYNC_INTERVAL", 72*time.Hour),

		NewsAPIURL:       getenv("NEWS_API_URL", "https://api.www.bmstu.ru/news"),
		NewsSyncEnabled:  getBoolEnv("NEWS_SYNC_ENABLED", false),
		NewsSyncInterval: getDurationEnv("NEWS_SYNC_INTERVAL", 24*time.Hour),
		NewsSyncLimit:    getIntEnv("NEWS_SYNC_LIMIT", 10),

		AssistantBaseURL:         getenv("ASSISTANT_BASE_URL", "https://polza.ai/api/v1"),
		AssistantAPIKey:          getenv("ASSISTANT_API_KEY", ""),
		AssistantModel:           getenv("ASSISTANT_MODEL", "gpt-4o-mini"),
		AssistantTimeout:         getDurationEnv("ASSISTANT_TIMEOUT", 30*time.Second),
		AssistantContextMaxChars: getIntEnv("ASSISTANT_CONTEXT_MAX_CHARS", 120000),
	}
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getBoolEnv(key string, fallback bool) bool {
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

func getDurationEnv(key string, fallback time.Duration) time.Duration {
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

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
