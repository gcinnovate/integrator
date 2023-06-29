package pages

import (
	"time"
)

// DHIS 2.40 type

// Status represents the status
type Status string

const (
	// StatusActive represents active status
	StatusActive Status = "ACTIVE"
	// StatusCompleted represents completed status
	StatusCompleted Status = "COMPLETED"
	// StatusCancelled represents cancelled status
	StatusCancelled Status = "CANCELLED"
)

// EventStatus represents the status
type EventStatus string

const (
	// EventStatusActive represents active status
	EventStatusActive EventStatus = "ACTIVE"
	// EventStatusCompleted represents completed status
	EventStatusCompleted EventStatus = "COMPLETED"
	// EventStatusVisited represents visited status
	EventStatusVisited EventStatus = "VISITED"
	// EventStatusScheduled ...
	EventStatusScheduled EventStatus = "SCHEDULED"
	// EventStatusOverdue ...
	EventStatusOverdue EventStatus = "OVERDUE"
	// EventStatusSkipped ...
	EventStatusSkipped EventStatus = "SKIPPED"
)

// Scheme represents DE, OU, P, pS, cOC, cO schemes
type Scheme string

const (
	// SchemeUID represents UID scheme
	SchemeUID Scheme = "UID"
	// SchemeCode represents CODE scheme
	SchemeCode Scheme = "CODE"
	// SchemeName represents NAME scheme
	SchemeName Scheme = "NAME"
	// SchemeAttribute represents ATTRIBUTE scheme
	SchemeAttribute Scheme = "ATTRIBUTE"
)

// Async specify how import is done
//type Async bool
//
//const (
//	// AsyncOn represents asynchronous calls to API during import
//	AsyncOn Async = true
//	// AsyncOff ...
//	AsyncOff Async = false
//)

// ReportMode when performing synchronous import. See importSummary
type ReportMode string

const (
	ReportModeFull     ReportMode = "FULL"
	ReportModeErrors   ReportMode = "ERROR"
	ReportModeWarnings ReportMode = "WARNINGS"
)

// ImportMode Indicates the mode of import
type ImportMode string

const (
	ImportModeValidate ImportMode = "VALIDATE"
	ImportModeCommit   ImportMode = "COMMIT"
)

// ImportStrategy indicates the effect the import should have
type ImportStrategy string

const (
	ImportStrategyCreate          ImportStrategy = "CREATE"
	ImportStrategyUpdate          ImportStrategy = "UPDATE"
	ImportStrategyCreateAndUpdate ImportStrategy = "CREATE_AND_UPDATE"
	ImportStrategyDelete          ImportStrategy = "DELETE"
)

// AtomicMode indicates how the import responds to validation errors
type AtomicMode string

const (
	AtomicModeAll    AtomicMode = "ALL"
	AtomicModeObject AtomicMode = "OBJECT"
)

// FlushMode indicates the frequency of flushing -
// how often data is pushed into the database during the import
type FlushMode string

const (
	FlushModeAuto   FlushMode = "AUTO"
	FlushModeObject FlushMode = "OBJECT"
)

// ValidationMode indicates the completeness of the validation step
type ValidationMode string

const (
	ValidationModeFull     ValidationMode = "FULL"
	ValidationModeFailFast ValidationMode = "FAIL_FAST"
	ValidationModeSkip     ValidationMode = "SKIP"
)

// TrackedEntity ....
type TrackedEntity struct {
	TrackedEntity     string `json:"trackedEntity,omitempty"`
	TrackedEntityType string `json:"trackedEntityType"`
	CreatedAt         string `json:"createdAt,omitempty"`
	CreatedAtClient   string `json:"createdAtClient,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
	UpdatedAtClient   string `json:"updatedAtClient,omitempty"`
	OrgUnit           string `json:"orgUnit"`
	Inactive          bool   `json:"inactive,omitempty"`
	Deleted           bool   `json:"deleted,omitempty"`
	Geometry          string `json:"geometry,omitempty"`
	StoredBy          string `json:"storedBy,omitempty"`
	CreatedBy         User   `json:"createdBy,omitempty"`
	UpdatedBy         User   `json:"updatedBy,omitempty"`
}

// EnrollmentV2 .....
type EnrollmentV2 struct {
	Enrollment        string         `json:"enrollment,omitempty"`
	Program           string         `json:"program"`
	TrackedEntity     string         `json:"trackedEntity"`
	TrackedEntityType string         `json:"trackedEntityType,omitempty"`
	Status            Status         `json:"status,omitempty"`
	OrgUnit           string         `json:"orgUnit"`
	OrgUnitName       string         `json:"OrgUnitName,omitempty"`
	EnrolledAt        Date           `json:"enrolledAt"`
	OccurredAt        string         `json:"occurredAt,omitempty"`
	CompletedAt       string         `json:"completedAt,omitempty"`
	CompletedBy       string         `json:"completedBy,omitempty"`
	FollowUp          bool           `json:"followUp,omitempty"`
	Deleted           bool           `json:"deleted,omitempty"`
	Geometry          string         `json:"geometry,omitempty"`
	StoredBy          string         `json:"storedBy,omitempty"`
	CreatedBy         User           `json:"createdBy,omitempty"`
	UpdatedBy         User           `json:"updatedBy,omitempty"`
	Attributes        []AttributeV2  `json:"attributes,omitempty"`
	Events            []EventV2      `json:"events,omitempty"`
	Relationships     []Relationship `json:"relationships,omitempty"`
	Notes             []Note         `json:"notes,omitempty"`
}

// EventV2 ...
type EventV2 struct {
	Event                    string         `json:"event,omitempty"`
	ProgramStage             string         `json:"programStage"`
	Enrollment               string         `json:"enrollment"`
	Program                  string         `json:"program,omitempty"`
	TrackedEntity            string         `json:"trackedEntity,omitempty"`
	Status                   EventStatus    `json:"status,omitempty"`
	EnrollmentStatus         string         `json:"enrollmentStatus,omitempty"`
	OrgUnit                  string         `json:"orgUnit"`
	OrgUnitName              string         `json:"OrgUnitName,omitempty"`
	CreatedAt                string         `json:"createdAt,omitempty"`
	CreatedAtClient          string         `json:"createdAtClient,omitempty"`
	UpdatedAt                string         `json:"updatedAt,omitempty"`
	UpdatedAtClient          string         `json:"updatedAtClient,omitempty"`
	ScheduledAt              string         `json:"scheduledAt,omitempty"`
	OccurredAt               string         `json:"occurredAt"`
	CompletedAt              string         `json:"completedAt,omitempty"`
	CompletedBy              string         `json:"completedBy,omitempty"`
	FollowUp                 bool           `json:"followUp,omitempty"`
	Deleted                  bool           `json:"deleted,omitempty"`
	Geometry                 string         `json:"geometry,omitempty"`
	StoredBy                 string         `json:"storedBy,omitempty"`
	CreatedBy                User           `json:"createdBy,omitempty"`
	UpdatedBy                User           `json:"updatedBy,omitempty"`
	AttributeOptionCombo     string         `json:"attributeOptionCombo,omitempty"`
	AttributeCategoryOptions string         `json:"attributeCategoryOptions"`
	AssignedUser             string         `json:"assignedUser,omitempty"`
	DataValues               []DataValueV2  `json:"dataValues,omitempty"`
	Relationships            []Relationship `json:"relationships,omitempty"`
	Notes                    []Note         `json:"notes,omitempty"`
}

// EventItem for relationships link
type EventItem struct {
	Event string `json:"event"`
}

// TrackedEntityItem ...
type TrackedEntityItem struct {
	TrackedEntity string `json:"trackedEntity"`
}

// EnrollmentItem ...
type EnrollmentItem struct {
	Enrollment string `json:"enrollment"`
}

// FromItem for relationship link
type FromItem struct {
	Event         EventItem         `json:"event,omitempty"`
	Enrollment    EnrollmentItem    `json:"enrollment,omitempty"`
	TrackedEntity TrackedEntityItem `json:"trackedEntity,omitempty"`
}

// ToItem for relationship link
type ToItem struct {
	Event         EventItem         `json:"event,omitempty"`
	Enrollment    EnrollmentItem    `json:"enrollment,omitempty"`
	TrackedEntity TrackedEntityItem `json:"trackedEntity,omitempty"`
}

// Relationship ...
type Relationship struct {
	Relationship     string   `json:"relationship,omitempty"`
	RelationshipType string   `json:"relationshipType"`
	RelationshipName string   `json:"relationshipName,omitempty"`
	CreatedAt        string   `json:"createdAt,omitempty"`
	UpdatedAt        string   `json:"updatedAt,omitempty"`
	Bidirectional    bool     `json:"bidirectional,omitempty"`
	From             FromItem `json:"from"`
	To               ToItem   `json:"to"`
}

// Attributes are actual values describing Tracked Entities
//type Attribute struct {
//
//}

// DataValueV2 struct
type DataValueV2 struct {
	DataElement       string `json:"dataElement"`
	Value             string `json:"value,omitempty"`
	ProvidedElseWhere bool   `json:"providedElseWhere,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
	StoredBy          string `json:"storedBy,omitempty"`
	CreatedBy         string `json:"createdBy,omitempty"`
}

// AttributeV2 .. actual values describing tracked entities
type AttributeV2 struct {
	Attribute   string `json:"attribute"`
	Code        string `json:"code,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	StoredBy    string `json:"storedBy,omitempty"`
	ValueType   string `json:"valueType,omitempty"`
	Value       string `json:"value,omitempty"`
}

// Note for additional info or comment
type Note struct {
	Note      string `json:"note,omitempty"`
	Value     string `json:"value"`
	StoredAt  string `json:"storedAt,omitempty"`
	StoredBy  string `json:"storedBy,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
}

// User ...
type User struct {
	UID       string `json:"uid"`
	Username  string `json:"username"`
	FirstName string `json:"firstName,omitempty"`
	SurName   string `json:"surName,omitempty"`
}

// FlatPayload represents the FLAT payload - most straightforward
// requires UIDs for the connection between objects
type FlatPayload struct {
	TrackedEntities []TrackedEntity `json:"trackedEntities"`
	Enrollments     []EnrollmentV2  `json:"enrollments"`
	Events          []EventV2       `json:"events"`
	Relationships   []Relationship  `json:"relationships"`
}

// NestedPayload represents the NESTED payload - most commonly used
// client does not need to provide UIDs for the connections between objects
type NestedPayload struct {
	TrackedEntities []TrackedEntity `json:"trackedEntities"`
}

// JobResponse ...
type JobResponse struct {
	ID           string `json:"id"`
	ResponseType string `json:"responseType"`
	Location     string `json:"location"`
}

// AsyncResponse ...
type AsyncResponse struct {
	HTTPStatus     string      `json:"httpStatus"`
	HTTPStatusCode string      `json:"httpStatusCode"`
	Status         string      `json:"status"`
	Message        string      `json:"message"`
	Response       JobResponse `json:"response"`
}

// TrackerJobImportResponse ...
type TrackerJobImportResponse struct {
	UID       string    `json:"uid"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	Time      time.Time `json:"time"`
	Message   string    `json:"message"`
	Completed string    `json:"completed"`
	ID        string    `json:"id"`
}

// ImportSummary ...
type ImportSummary struct {
	Status           string      `json:"status"`
	ValidationReport string      `json:"validationReport"`
	Stats            ImportStats `json:"stats"`
	TimingStats      Timers      `json:"timingStats"`
	BundleReport     struct {
		TypeReportMap TypeReportMap `json:"typeReportMap"`
	} `json:"bundleReport"`
	Message any `json:"message"`
}

// Report  for ErrorReport or WarningReport...
type Report struct {
	UID         string `json:"uid"`
	Message     string `json:"message"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Warning     string `json:"warning,omitempty"`
	TrackerType string `json:"trackerType,omitempty"`
}

// ValidationReport ...
type ValidationReport struct {
	ErrorReports   []Report `json:"errorReports"`
	WarningReports []Report `json:"warningReports"`
}

// ImportStats ...provide a quick overview of the import.
type ImportStats struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
	Ignored int `json:"ignored"`
	Total   int `json:"total"`
}

// Timers ...
type Timers struct {
	Preheat     string `json:"preheat"`
	PreProcess  string `json:"preprocess"`
	TotalImport string `json:"totalImport"`
	Validation  string `json:"validation"`
}

// ObjectReport tracker object report
type ObjectReport struct {
	UID          string   `json:"uid"`
	Index        int      `json:"index"`
	TrackerType  string   `json:"trackerType"`
	ErrorReports []Report `json:"errorReport"`
}

// TypeReportMapObject ... for the bundleReport object
type TypeReportMapObject struct {
	TrackerType   string         `json:"trackerType"`
	Stats         ImportStats    `json:"stats"`
	ObjectReports []ObjectReport `json:"objectReports"`
}

// TypeReportMap ...
type TypeReportMap struct {
	TrackedEntityObject TypeReportMapObject `json:"TRACKED_ENTITY"`
	EnrollmentObject    TypeReportMapObject `json:"ENROLLMENT"`
	EventObject         TypeReportMapObject `json:"EVENT"`
}
