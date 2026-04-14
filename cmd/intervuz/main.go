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
	graphdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/graph"
	healthdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/health"
	imagedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/image"
	placedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/place"
	roomdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/room"
	routedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/route"
	scheduledelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/schedule"
	scheduleexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/schedule"
	structureexternal "github.com/GIT_USER_ID/GIT_REPO_ID/internal/external/http/structure"
	mid "github.com/GIT_USER_ID/GIT_REPO_ID/internal/middleware"
	graphrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/graph/memory"
	placerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/place/memory"
	roomrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/room/memory"
	schedulefilerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/schedule/file"
	schedulepgrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/schedule/postgres"
	structurerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/structure/postgres"
	graphusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/graph"
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

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var db *sql.DB
	var err error
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
	roomRepo := roomrepository.New()

	placeUseCase := placeusecase.New(placeRepo)
	graphUseCase := graphusecase.New(graphRepo)
	routeUseCase := routeusecase.New(graphRepo, placeRepo)
	scheduleUseCase := scheduleusecase.New(scheduleRepo)
	roomUseCase := roomusecase.New(roomRepo)

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

	mux := http.NewServeMux()
	mux.Handle("GET /health", healthHandler)
	mux.Handle("GET /image", imageHandler)
	mux.Handle("GET /places", placeListHandler)
	mux.Handle("GET /places/{placeID}", placeGetHandler)
	mux.Handle("GET /graph", graphGetHandler)
	mux.Handle("POST /routes", routeBuildHandler)
	mux.Handle("POST /schedule/import", scheduleImportHandler)
	mux.Handle("GET /users/schedule/groups", groupCatalogHandler)
	mux.Handle("GET /users/schedule/events/{eventID}", eventHandler)
	mux.Handle("GET /users/schedule/{groupID}", groupScheduleHandler)
	mux.Handle("GET /rooms/availability", roomListAvailabilityHandler)

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
	flag.Parse()
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
