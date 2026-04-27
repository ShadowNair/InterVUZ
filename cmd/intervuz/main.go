package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/bootstrap"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/config"
	assistantdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/assistant"
	graphdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/graph"
	healthdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/health"
	imagedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/image"
	newsdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/news"
	placedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/place"
	roomdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/room"
	routedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/route"
	scheduledelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/schedule"
	assistantexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/assistant"
	newsexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/news"
	scheduleexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/schedule"
	structureexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/structure"
	mid "github.com/GIT_USER_ID/GIT_REPO_ID/internal/middleware"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/platform/httpjson"
	graphrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/graph/memory"
	newsrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/news/postgres"
	placerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/place/memory"
	roommemoryrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/room/memory"
	roompgrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/room/postgres"
	schedulefilerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/schedule/file"
	schedulepgrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/schedule/postgres"
	structurerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/structure/postgres"
	assistantusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/assistant"
	graphusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/graph"
	newsusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/news"
	placeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/place"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
	routeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/route"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
	schedulesyncusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedulesync"
	structureusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/structure"
)

func main() {
	cfg := config.Load()
	parseFlags(&cfg)
	academicWeek1Start, err := parseAcademicWeek1StartDate(cfg.AcademicWeek1StartDate)
	if err != nil {
		log.Fatalf("bad ACADEMIC_WEEK1_START_DATE: %v", err)
	}

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var db *sql.DB
	if cfg.DatabaseDSN != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(appCtx, 10*time.Second)
		if err := db.PingContext(ctx); err != nil {
			cancel()
			log.Fatalf("ping database: %v", err)
		}
		cancel()

		if cfg.SyncOnStartup {
			if err := runSyncCycle(appCtx, cfg, db); err != nil {
				log.Fatalf("startup sync failed: %v", err)
			}
		} else {
			log.Printf("startup sync skipped")
		}

		startPeriodicSync(appCtx, cfg, db)

		if cfg.NewsSyncEnabled {
			if err := syncNews(appCtx, cfg, db); err != nil {
				log.Fatalf("startup news sync failed: %v", err)
			}
			startPeriodicNewsSync(appCtx, cfg, db)
		} else {
			log.Printf("news sync disabled")
		}
	}

	places, err := bootstrap.LoadPlacesSeed(cfg.PlacesDataPath)
	if err != nil {
		log.Fatalf("load places seed: %v", err)
	}

	var scheduleRepo scheduleusecase.Repository
	if db != nil {
		scheduleRepo = schedulepgrepository.New(db, cfg.StructureTargetRootUUID)
	} else {
		scheduleRepo, err = schedulefilerepository.New(cfg.ScheduleDataDir)
		if err != nil {
			log.Fatalf("init schedule storage: %v", err)
		}
	}

	placeRepo := placerepository.New(places)
	graphRepo := graphrepository.New(places)
	var roomRepo roomusecase.Repository
	if db != nil {
		roomRepo = roompgrepository.New(db, places, academicWeek1Start)
	} else {
		roomRepo = roommemoryrepository.New(places)
	}

	placeUseCase := placeusecase.New(placeRepo)
	graphUseCase := graphusecase.New(graphRepo)
	routeUseCase := routeusecase.New(graphRepo, placeRepo)
	scheduleUseCase := scheduleusecase.New(scheduleRepo)
	roomUseCase := roomusecase.New(roomRepo)
	assistantClient := assistantexternal.New(cfg.AssistantBaseURL, cfg.AssistantAPIKey, cfg.AssistantTimeout, nil)
	assistantUseCase := assistantusecase.New(assistantClient, cfg.AssistantModel, cfg.AssistantContextMaxChars)

	healthHandler := healthdelivery.NewHandler()
	imageHandler := imagedelivery.NewHandler(cfg.RootImagePath)
	placeListHandler := placedelivery.NewListHandler(placeUseCase)
	placeGetHandler := placedelivery.NewGetHandler(placeUseCase)
	graphGetHandler := graphdelivery.NewGetHandler(graphUseCase)
	routeBuildHandler := routedelivery.NewBuildHandler(routeUseCase)
	scheduleImportHandler := scheduledelivery.NewImportHandler(scheduleUseCase)
	groupCatalogHandler := scheduledelivery.NewGroupCatalogHandler(scheduleUseCase)
	groupScheduleHandler := scheduledelivery.NewGroupScheduleHandler(scheduleUseCase)
	eventHandler := scheduledelivery.NewEventHandler(scheduleUseCase)
	roomListAvailabilityHandler := roomdelivery.NewListAvailabilityHandler(roomUseCase)
	roomGetScheduleHandler := roomdelivery.NewGetScheduleHandler(roomUseCase)
	roomCreateBookingHandler := roomdelivery.NewCreateBookingHandler(roomUseCase)
	roomCancelBookingHandler := roomdelivery.NewCancelBookingHandler(roomUseCase)
	var newsListHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpjson.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "news storage is not configured")
	})
	if db != nil {
		newsRepo := newsrepository.New(db)
		newsUseCase := newsusecase.New(nil, newsRepo)
		newsListHandler = newsdelivery.NewListHandler(newsUseCase)
	}
	assistantChatHandler := assistantdelivery.NewChatHandler(assistantUseCase)

	mux := http.NewServeMux()
	mux.Handle("GET /health", healthHandler)
	mux.Handle("GET /image", imageHandler)
	mux.Handle("GET /news", newsListHandler)
	mux.Handle("GET /places", placeListHandler)
	mux.Handle("GET /places/{placeID}", placeGetHandler)
	mux.Handle("GET /graph", graphGetHandler)
	mux.Handle("POST /routes", routeBuildHandler)
	mux.Handle("POST /schedule/import", scheduleImportHandler)
	mux.Handle("GET /users/schedule/groups", groupCatalogHandler)
	mux.Handle("GET /users/schedule/events/{eventID}", eventHandler)
	mux.Handle("GET /users/schedule/{groupID}", groupScheduleHandler)
	mux.Handle("GET /rooms/availability", roomListAvailabilityHandler)
	mux.Handle("GET /rooms/{roomID}/schedule", roomGetScheduleHandler)
	mux.Handle("POST /rooms/{roomID}/bookings", roomCreateBookingHandler)
	mux.Handle("DELETE /rooms/bookings/{bookingID}", roomCancelBookingHandler)
	mux.Handle("POST /assistant/chat", assistantChatHandler)

	corsMiddleware := mid.CORS(nil)
	handler := corsMiddleware(loggingMiddleware(mux))

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-appCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("server started on :%s", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen and serve: %v", err)
	}
}

func parseFlags(cfg *config.Config) {
	flag.BoolVar(&cfg.SyncOnStartup, "sync-on-startup", cfg.SyncOnStartup, "run structure and schedule sync during application startup")
	flag.DurationVar(&cfg.SyncInterval, "sync-interval", cfg.SyncInterval, "background sync interval; set 0 to disable periodic sync")

	flag.BoolVar(&cfg.NewsSyncEnabled, "sync-news", cfg.NewsSyncEnabled, "enable news sync on startup and periodically")
	flag.DurationVar(&cfg.NewsSyncInterval, "news-sync-interval", cfg.NewsSyncInterval, "news sync interval; set 0 to disable periodic news sync")
	flag.IntVar(&cfg.NewsSyncLimit, "news-sync-limit", cfg.NewsSyncLimit, "number of latest news items to sync")
	flag.Parse()
}

func parseAcademicWeek1StartDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("ACADEMIC_WEEK1_START_DATE is required")
	}

	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return time.Time{}, errors.New("must be YYYY-MM-DD")
	}

	return parsed, nil
}

func startPeriodicSync(appCtx context.Context, cfg config.Config, db *sql.DB) {
	if db == nil {
		return
	}

	if cfg.SyncInterval <= 0 {
		log.Printf("periodic sync disabled")
		return
	}

	log.Printf("periodic sync enabled: interval=%s", cfg.SyncInterval)

	go func() {
		ticker := time.NewTicker(cfg.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-appCtx.Done():
				return
			case <-ticker.C:
				if err := runSyncCycle(appCtx, cfg, db); err != nil {
					log.Printf("periodic sync failed: %v", err)
				}
			}
		}
	}()
}

func runSyncCycle(appCtx context.Context, cfg config.Config, db *sql.DB) error {
	if err := syncStructure(appCtx, cfg, db); err != nil {
		return err
	}

	if err := syncSchedules(appCtx, cfg, db); err != nil {
		return err
	}

	return nil
}

func syncStructure(appCtx context.Context, cfg config.Config, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(appCtx, 2*time.Minute)
	defer cancel()

	externalClient := structureexternal.New(cfg.StructureAPIURL, cfg.StructureTargetRootUUID, nil)
	repo := structurerepository.New(db)
	uc := structureusecase.New(externalClient, repo)

	count, err := uc.Sync(ctx)
	if err != nil {
		return err
	}

	log.Printf("structure sync completed: %d units", count)
	return nil
}

func syncSchedules(appCtx context.Context, cfg config.Config, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(appCtx, 10*time.Minute)
	defer cancel()

	source := scheduleexternal.New(cfg.ScheduleGroupAPIBaseURL, nil)
	repo := schedulepgrepository.New(db, cfg.StructureTargetRootUUID)
	uc := schedulesyncusecase.New(repo, source, repo)

	stats, err := uc.SyncAll(ctx)
	if err != nil {
		return err
	}

	log.Printf("schedule sync completed: %d/%d groups, %d events saved", stats.GroupsSynced, stats.GroupsTotal, stats.EventsSaved)
	return nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(startedAt))
	})
}

func syncNews(appCtx context.Context, cfg config.Config, db *sql.DB) error {
	if db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(appCtx, 2*time.Minute)
	defer cancel()

	source := newsexternal.New(cfg.NewsAPIURL, nil)
	repo := newsrepository.New(db)
	uc := newsusecase.New(source, repo)

	if err := uc.Sync(ctx, cfg.NewsSyncLimit); err != nil {
		return err
	}

	log.Printf("news sync completed: latest %d items synced", cfg.NewsSyncLimit)
	return nil
}

func startPeriodicNewsSync(appCtx context.Context, cfg config.Config, db *sql.DB) {
	if db == nil {
		return
	}

	if !cfg.NewsSyncEnabled {
		return
	}

	if cfg.NewsSyncInterval <= 0 {
		log.Printf("periodic news sync disabled")
		return
	}

	log.Printf("periodic news sync enabled: interval=%s limit=%d", cfg.NewsSyncInterval, cfg.NewsSyncLimit)

	go func() {
		ticker := time.NewTicker(cfg.NewsSyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-appCtx.Done():
				return
			case <-ticker.C:
				if err := syncNews(appCtx, cfg, db); err != nil {
					log.Printf("periodic news sync failed: %v", err)
				}
			}
		}
	}()
}
