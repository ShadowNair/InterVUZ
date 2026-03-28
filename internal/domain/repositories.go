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

type AuthRepository interface {
	CreateProfile(context.Context, Profile) (*Profile, error)
	GetProfileByEmail(context.Context, string) (*Profile, error)
	GetProfileByID(context.Context, string) (*Profile, error)
	CreateRefreshToken(context.Context, RefreshToken) (*RefreshToken, error)
	GetRefreshToken(context.Context, string) (*RefreshToken, error)
	RevokeRefreshToken(context.Context, string) error
	DeleteExpiredRefreshTokens(context.Context) error
}
