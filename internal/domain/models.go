package domain

import "time"

type PlaceFilter struct {
	Type     string
	Building string
	Floor    *int
	Search   string
}

type RoomAvailabilityFilter struct {
	StartsAt  time.Time
	EndsAt    time.Time
	Building  string
	Capacity  *int
	Equipment []string
}

type Coordinates struct {
	Building string  `json:"building"`
	Floor    int     `json:"floor"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type Place struct {
	ID           string      `json:"id"`
	ExternalUUID string      `json:"externalUuid,omitempty"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description,omitempty"`
	Coordinates  Coordinates `json:"coordinates"`
	Tags         []string    `json:"tags,omitempty"`
	IsAccessible bool        `json:"isAccessible"`
}

type PlacesResponse struct {
	Items []Place `json:"items"`
	Total int     `json:"total"`
}

type RouteRequest struct {
	FromPlaceID    string `json:"fromPlaceId"`
	ToPlaceID      string `json:"toPlaceId"`
	AccessibleOnly bool   `json:"accessibleOnly"`
}

type RouteStep struct {
	Order       int         `json:"order"`
	Instruction string      `json:"instruction"`
	PlaceID     string      `json:"placeId,omitempty"`
	Coordinates Coordinates `json:"coordinates"`
}

type Route struct {
	DistanceMeters           int         `json:"distanceMeters"`
	EstimatedDurationMinutes int         `json:"estimatedDurationMinutes"`
	Steps                    []RouteStep `json:"steps"`
}

type ScheduleImportRequest struct {
	UserID     string              `json:"userId"`
	SourceType string              `json:"sourceType"`
	SourceURL  string              `json:"sourceUrl,omitempty"`
	FileName   string              `json:"fileName,omitempty"`
	Events     []ScheduleEventItem `json:"events,omitempty"`
}

type ScheduleImportResult struct {
	ImportID       string `json:"importId"`
	Status         string `json:"status"`
	ImportedEvents int    `json:"importedEvents,omitempty"`
}

type GroupCatalogResponse struct {
	Data GroupCatalogNode `json:"data"`
}

type GroupCatalogNode struct {
	Abbr       string             `json:"abbr"`
	Name       string             `json:"name,omitempty"`
	UUID       string             `json:"uuid,omitempty"`
	Course     int                `json:"course,omitempty"`
	Semester   int                `json:"semester,omitempty"`
	NodeType   string             `json:"nodeType,omitempty"`
	ParentUUID string             `json:"parentUuid,omitempty"`
	Children   []GroupCatalogNode `json:"children,omitempty"`
}

type GroupScheduleResponse struct {
	Data GroupSchedule `json:"data"`
}

type GroupSchedule struct {
	Link     string              `json:"link"`
	Type     string              `json:"type"`
	UUID     string              `json:"uuid"`
	Title    string              `json:"title"`
	Schedule []ScheduleEventItem `json:"schedule"`
}

type ScheduleEventResponse struct {
	Data ScheduleEventItem `json:"data"`
}

type ScheduleEventItem struct {
	Day              int                `json:"day"`
	Time             int                `json:"time"`
	Week             string             `json:"week"`
	Groups           []ScheduleGroupRef `json:"groups"`
	Stream           ScheduleStream     `json:"stream"`
	EndTime          string             `json:"endTime"`
	Teachers         []Teacher          `json:"teachers"`
	Audiences        []Audience         `json:"audiences"`
	StartTime        string             `json:"startTime"`
	Discipline       Discipline         `json:"discipline"`
	Permission       string             `json:"permission"`
	EndTimeMinNum    int                `json:"endTimeMinNum"`
	EndTimeHourNum   int                `json:"endTimeHourNum"`
	StartTimeMinNum  int                `json:"startTimeMinNum"`
	StartTimeHourNum int                `json:"startTimeHourNum"`
}

type ScheduleGroupRef struct {
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	DepartmentUID string `json:"department_uid,omitempty"`
}

type ScheduleStream struct {
	Name   string                `json:"name"`
	Groups []ScheduleStreamGroup `json:"groups"`
}

type ScheduleStreamGroup struct {
	Sub1      int    `json:"sub1"`
	Sub2      int    `json:"sub2"`
	GroupUUID string `json:"groupUuid"`
}

type Teacher struct {
	UUID       string `json:"uuid"`
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName,omitempty"`
}

type Audience struct {
	Name               string `json:"name"`
	UUID               string `json:"uuid"`
	Building           string `json:"building"`
	RootUUID           string `json:"root_uuid"`
	DepartmentUID      string `json:"department_uid,omitempty"`
	BuildingIDBlock    int    `json:"building_id_block"`
	BuildingIDBuilding int    `json:"building_id_building"`
}

type Discipline struct {
	Abbr      string `json:"abbr"`
	ActType   string `json:"actType"`
	FullName  string `json:"fullName"`
	ShortName string `json:"shortName"`
}

type RoomAvailability struct {
	RoomID        string   `json:"roomId"`
	Name          string   `json:"name"`
	Building      string   `json:"building"`
	Floor         int      `json:"floor"`
	Capacity      int      `json:"capacity"`
	Equipment     []string `json:"equipment,omitempty"`
	AvailableFrom string   `json:"availableFrom"`
	AvailableTo   string   `json:"availableTo"`
}

type RoomAvailabilityResponse struct {
	Items []RoomAvailability `json:"items"`
}

type RoomInfo struct {
	RoomID    string   `json:"roomId"`
	Name      string   `json:"name"`
	Building  string   `json:"building"`
	Floor     int      `json:"floor"`
	Equipment []string `json:"equipment,omitempty"`
}

type RoomScheduleItem struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Title         string   `json:"title"`
	StartsAt      string   `json:"startsAt"`
	EndsAt        string   `json:"endsAt"`
	StartTime     string   `json:"startTime"`
	EndTime       string   `json:"endTime"`
	Week          string   `json:"week,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	Teachers      []string `json:"teachers,omitempty"`
	BookerName    string   `json:"bookerName,omitempty"`
	BookerContact string   `json:"bookerContact,omitempty"`
}

type RoomScheduleResponse struct {
	Room  RoomInfo           `json:"room"`
	Date  string             `json:"date"`
	Items []RoomScheduleItem `json:"items"`
}

type RoomBookingRequest struct {
	RoomID        string    `json:"roomId"`
	StartsAt      time.Time `json:"startsAt"`
	EndsAt        time.Time `json:"endsAt"`
	BookerName    string    `json:"bookerName"`
	BookerContact string    `json:"bookerContact"`
}

type CreateRoomBookingRequest struct {
	StartsAt      string `json:"startsAt"`
	BookerName    string `json:"bookerName"`
	BookerContact string `json:"bookerContact"`
}

type RoomBooking struct {
	ID            string `json:"id"`
	RoomID        string `json:"roomId"`
	RoomName      string `json:"roomName"`
	StartsAt      string `json:"startsAt"`
	EndsAt        string `json:"endsAt"`
	BookerName    string `json:"bookerName"`
	BookerContact string `json:"bookerContact"`
}

type RoomBookingsResponse struct {
	Date  string        `json:"date"`
	Items []RoomBooking `json:"items"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
