package main

import (
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/config"
	authlogin "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/login"
	authlogout "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/logout"
	authme "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/me"
	authrefresh "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/refresh"
	authregister "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/auth/register"
	healthhandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/health"
	imagehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/image"
	placehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/places"
	roomhandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/rooms"
	routehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/routes"
	schedulehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/schedule"
	mid "github.com/GIT_USER_ID/GIT_REPO_ID/internal/middleware"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/storage"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/storage/postgres"
	profile_repo "github.com/GIT_USER_ID/GIT_REPO_ID/internal/storage/postgres/profile-repo"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
	profileusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/profile"
)

func main() {
	cfg := config.Load()

	postgresDB, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("init postgres storage: %v", err)
	}

	placeRepo := storage.NewStubPlaceRepository()
	roomRepo := storage.NewStubRoomRepository()
	scheduleRepo, err := storage.NewStubScheduleRepository(filepath.Join("schedule", "lksJSON"))
	if err != nil {
		log.Fatalf("init schedule storage: %v", err)
	}
	authRepo := profile_repo.New(postgresDB)

	placeUseCase := usecase.NewPlaceUseCase(placeRepo)
	routeUseCase := usecase.NewRouteUseCase(placeRepo)
	scheduleUseCase := usecase.NewScheduleUseCase(scheduleRepo)
	roomUseCase := usecase.NewRoomUseCase(roomRepo)
	authUseCase := profileusecase.New(authRepo, profileusecase.Config{
		AccessTokenSecret:  cfg.AccessTokenSecret,
		RefreshTokenSecret: cfg.RefreshTokenSecret,
	})

	placeHTTPHandler := placehandler.NewHandler(placeUseCase)
	routeHTTPHandler := routehandler.NewHandler(routeUseCase)
	scheduleHTTPHandler := schedulehandler.NewHandler(scheduleUseCase)
	roomHTTPHandler := roomhandler.NewHandler(roomUseCase)
	authRegisterHTTPHandler := authregister.NewHandler(authUseCase)
	authLoginHTTPHandler := authlogin.NewHandler(authUseCase)
	authRefreshHTTPHandler := authrefresh.NewHandler(authUseCase)
	authLogoutHTTPHandler := authlogout.NewHandler(authUseCase)
	authMeHTTPHandler := authme.NewHandler(authUseCase)
	healthHTTPHandler := healthhandler.NewHandler()
	imageHTTPHandler := imagehandler.NewHandler(cfg.RootImagePath)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authRegisterHTTPHandler.Handle)
	mux.HandleFunc("/auth/login", authLoginHTTPHandler.Handle)
	mux.HandleFunc("/auth/refresh", authRefreshHTTPHandler.Handle)
	mux.HandleFunc("/auth/logout", authLogoutHTTPHandler.Handle)
	mux.HandleFunc("/auth/me", authMeHTTPHandler.Handle)
	mux.HandleFunc("/places", placeHTTPHandler.List)
	mux.HandleFunc("/places/", placeHTTPHandler.GetByID)
	mux.HandleFunc("/routes", routeHTTPHandler.Build)
	mux.HandleFunc("/schedule/import", scheduleHTTPHandler.Import)
	mux.HandleFunc("/users/schedule/groups", scheduleHTTPHandler.GetGroups)
	mux.HandleFunc("/users/schedule/events/", scheduleHTTPHandler.GetEvent)
	mux.HandleFunc("/users/schedule/", scheduleHTTPHandler.GetGroupSchedule)
	mux.HandleFunc("/rooms/availability", roomHTTPHandler.ListAvailability)
	mux.HandleFunc("/health", healthHTTPHandler.Handle)
	mux.HandleFunc("/image", imageHTTPHandler.Serve)

	corsMiddleware := mid.CORS(nil)
	handlerWithCORS := corsMiddleware(mux)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handlerWithCORS,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server started on :%s", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}
