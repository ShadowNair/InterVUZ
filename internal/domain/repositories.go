package domain

import "context"

type PlaceRepository interface {
	List(context.Context, PlaceFilter) ([]Place, error)
	GetByID(context.Context, string) (*Place, error)
}

type ScheduleRepository interface {
	Import(context.Context, ScheduleImportRequest) (*ScheduleImportResult, error)
	GetGroupCatalog(context.Context) (*GroupCatalogResponse, error)
	GetGroupSchedule(context.Context, string) (*GroupScheduleResponse, error)
	GetEvent(context.Context, string) (*ScheduleEventResponse, error)
}

type RoomRepository interface {
	ListAvailability(context.Context, RoomAvailabilityFilter) ([]RoomAvailability, error)
}
