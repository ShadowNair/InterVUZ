package domain

import "encoding/json"

type StructureResponse struct {
	Data StructureNode `json:"data"`
}

type StructureNode struct {
	Abbr       string          `json:"abbr"`
	Name       string          `json:"name,omitempty"`
	UUID       string          `json:"uuid,omitempty"`
	Course     int             `json:"course,omitempty"`
	Semester   int             `json:"semester,omitempty"`
	NodeType   string          `json:"nodeType,omitempty"`
	ParentUUID string          `json:"parentUuid,omitempty"`
	Children   []StructureNode `json:"children,omitempty"`
}

type StructureUnit struct {
	ExternalUUID     string
	ParentExternalID *string

	Code     string
	Name     string
	NodeType string

	Course   *int
	Semester *int

	FacultyExternalUUID    *string
	FacultyCode            *string
	FacultyName            *string
	DepartmentExternalUUID *string
	DepartmentCode         *string
	DepartmentName         *string
	CourseNodeExternalUUID *string
	CourseNodeName         *string

	RawPayload json.RawMessage
}