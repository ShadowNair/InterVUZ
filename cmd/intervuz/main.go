package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/bootstrap"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/config"
	graphdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/graph"
	healthdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/health"
	imagedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/image"
	placedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/place"
	roomdelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/room"
	routedelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/route"
	scheduledelivery "github.com/GIT_USER_ID/GIT_REPO_ID/internal/delivery/http/schedule"
	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
	mid "github.com/GIT_USER_ID/GIT_REPO_ID/internal/middleware"
	graphrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/graph/memory"
	staticgraphrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/graph/static"
	svgrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/graph/svg"
	placerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/place/memory"
	roomrepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/room/memory"
	schedulerepository "github.com/GIT_USER_ID/GIT_REPO_ID/internal/repository/schedule/file"
	graphusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/graph"
	placeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/place"
	roomusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/room"
	routeusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/route"
	scheduleusecase "github.com/GIT_USER_ID/GIT_REPO_ID/internal/usecase/schedule"
)

func main() {
	cfg := config.Load()

	places, err := bootstrap.LoadPlacesSeed(cfg.PlacesDataPath)
	if err != nil {
		log.Fatalf("load places seed: %v", err)
	}

	scheduleRepo, err := schedulerepository.New(cfg.ScheduleDataDir)
	if err != nil {
		log.Fatalf("init schedule storage: %v", err)
	}

	placeRepo := placerepository.New(places)
	roomRepo := roomrepository.New()

	var graphRepo interface {
		Get(context.Context) (domain.NavigationGraph, error)
	}

	graphFromJSON, err := bootstrap.LoadNavigationGraphFromJSON(cfg.GraphJSONPath)
	if err == nil {
		graphRepo = staticgraphrepository.New(graphFromJSON)
	} else {
		log.Printf("load graph from json failed, continue with svg fallback: %v", err)
		svgGraphRepo, svgErr := svgrepository.New(cfg.GraphSVGPath, places)
		if svgErr != nil {
			log.Printf("load graph from svg failed, fallback to generated graph: %v", svgErr)
			graphRepo = graphrepository.New(places)
		} else {
			graphRepo = svgGraphRepo
		}
	}

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

	log.Printf("server started on :%s", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(startedAt))
	})
}
