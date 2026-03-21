package main

import (
	"log"
	"net/http"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/config"
	mid "github.com/GIT_USER_ID/GIT_REPO_ID/internal/middleware"
	healthhandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/health"
	imagehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/image"
	placehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/places"
	roomhandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/rooms"
	routehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/routes"
	schedulehandler "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/handlers/schedule"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/storage"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase"
)

func main() {
	cfg := config.Load()

	placeRepo := storage.NewStubPlaceRepository()
	roomRepo := storage.NewStubRoomRepository()
	scheduleRepo, err := storage.NewStubScheduleRepository("schedule/lksJSON")
	if err != nil {
		log.Fatalf("init schedule storage: %v", err)
	}

	placeUseCase := usecase.NewPlaceUseCase(placeRepo)
	routeUseCase := usecase.NewRouteUseCase(placeRepo)
	scheduleUseCase := usecase.NewScheduleUseCase(scheduleRepo)
	roomUseCase := usecase.NewRoomUseCase(roomRepo)

	placeHTTPHandler := placehandler.NewHandler(placeUseCase)
	routeHTTPHandler := routehandler.NewHandler(routeUseCase)
	scheduleHTTPHandler := schedulehandler.NewHandler(scheduleUseCase)
	roomHTTPHandler := roomhandler.NewHandler(roomUseCase)
	healthHTTPHandler := healthhandler.NewHandler()
	imageHTTPHandler := imagehandler.NewHandler(cfg.RootImagePath)

	mux := http.NewServeMux()
	mux.HandleFunc("/places", placeHTTPHandler.List)
	mux.HandleFunc("/places/", placeHTTPHandler.GetByID)
	mux.HandleFunc("/routes", routeHTTPHandler.Build)
	mux.HandleFunc("/admin/schedule/import", scheduleHTTPHandler.Import)
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


