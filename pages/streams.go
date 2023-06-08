package pages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"os"
	"time"
)

// TeisPayload to submit to DHIS2
type TeisPayload struct {
	TrackedEntityInstances []TrackedEntityInstance `json:"trackedEntityInstances"`
}

// EnrollmentsPayload to submit to DHIS2
type EnrollmentsPayload struct {
	Enrollments []Enrollment `json:"enrollments"`
}

// EventsPayload to submit to DHIS2
type EventsPayload struct {
	Events []Event `json:"events"`
}

type Date struct {
	time.Time `json:",omitempty"`
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	} else {
		return d.Time.MarshalJSON()
	}
}

func (d *Date) UnmarshalJSON(b []byte) (err error) {
	fmt.Printf("Bytes for Date: %v\n", string(b))
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	date, err := time.Parse(`"2006-01-02"`, string(b))
	if err != nil {
		return err
	}
	d.Time = date
	return
}

// DataValue struct
type DataValue struct {
	DataElement string `json:"dataElement"`
	Value       string `json:"value"`
}

// Attribute struct
type Attribute struct {
	Attribute string      `json:"attribute"`
	Value     interface{} `json:"value"`
}

// TrackedEntityInstance struct
type TrackedEntityInstance struct {
	TrackedEntity     string      `json:"trackedEntityInstance"`
	TrackedEntityType string      `json:"trackedEntityType"`
	OrgUnit           string      `json:"orgUnit"`
	Attributes        []Attribute `json:"attributes"`
}

// Enrollment struct
type Enrollment struct {
	OrgUnit        string `json:"orgUnit"`
	Program        string `json:"program"`
	EnrollmentDate Date   `json:"enrollmentDate"`
	IncidentDate   Date   `json:"incidentDate"`
}

// EnrollmentWithEvent struct
type EnrollmentWithEvent struct {
	OrgUnit        string `json:"orgUnit"`
	Program        string `json:"program"`
	EnrollmentDate Date   `json:"enrollmentDate"`
	IncidentDate   Date   `json:"incidentDate"`
}

// TEIs struct
type TEIs struct {
	TrackedEntityInstances []TrackedEntityInstance `json:"trackedEntityInstances"`
}

// TEIsAndEnrollments struct
type TEIsAndEnrollments struct {
	TrackedEntity string       `json:"trackedEntity"`
	OrgUnit       string       `json:"orgUnit"`
	Attributes    []Attribute  `json:"attributes"`
	Enrollments   []Enrollment `json:"enrollments"`
}

// Coordinate struct
type Coordinate struct {
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}

// Event struct
type Event struct {
	Event                 string `json:"event,omitempty"`
	OrgUnit               string `json:"orgUnit"`
	Program               string `json:"program"`
	TrackedEntityInstance string `json:"trackedEntityInstance"`
	AttributeOptionCombo  string `json:"attributeOptionCombo,omitempty"`
	EventDate             Date   `json:"eventDate"`
	CompleteDate          Date   `json:"completeDate,omitempty"`
	Status                string `json:"status"`
	StoredBy              string `json:"storedBy,omitempty"`
	ProgramStage          string `json:"programStage"`
	Coordinate            `json:"-"`
	DataValues            []DataValue `json:"dataValues"`
}

// TEIAndEnrollmentsAndEvents struct
type TEIAndEnrollmentsAndEvents struct {
	TrackedEntityType string                `json:"trackedEntityType"`
	OrgUnit           string                `json:"orgUnit"`
	Attributes        []Attribute           `json:"attributes"`
	Enrollments       []EnrollmentWithEvent `json:"enrollmenents"`
}

// TrackedEntity streaming

// Entry represents each stream. If the stream fails, an error will be present.
type Entry struct {
	Error error
	Tei   TrackedEntityInstance
}

// Stream helps transmit each streams withing a channel.
type Stream struct {
	stream chan Entry
}

// NewJSONTeiStream returns a new `Stream` type.
func NewJSONTeiStream() Stream {
	return Stream{
		stream: make(chan Entry),
	}
}

// Watch watches JSON streams. Each stream entry will either have an error or a
// User object. Client code does not need to explicitly exit after catching an
// error as the `Start` method will close the channel automatically.
func (s Stream) Watch() <-chan Entry {
	return s.stream
}

// Start starts streaming JSON file line by line. If an error occurs, the channel
// will be closed.
func (s Stream) Start(path string, ftype string, progressFloat binding.ExternalFloat) {
	// Stop streaming channel as soon as nothing left to read in the file.
	defer close(s.stream)

	// Open file to read.
	file, err := os.Open(path)
	if err != nil {
		s.stream <- Entry{Error: fmt.Errorf("open file: %w", err)}
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Get file size for progress calculation
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	decoder := json.NewDecoder(file)

	// Read opening delimiter. `[` or `{`
	if _, err := decoder.Token(); err != nil {
		s.stream <- Entry{Error: fmt.Errorf("decode opening delimiter: %w", err)}
		return
	}

	// Find the start of the array
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			// panic(err)
			s.stream <- Entry{Error: fmt.Errorf("decoding error: %w", err)}
		}

		if token == ftype {
			_, err := decoder.Token()
			if err != nil {
				// panic(err)
				s.stream <- Entry{Error: fmt.Errorf("decoding error: %w", err)}
			}

			break
		}
	}

	// Read file content as long as there is something.
	i := 1
	for decoder.More() {
		var tei TrackedEntityInstance
		if err := decoder.Decode(&tei); err != nil {
			s.stream <- Entry{Error: fmt.Errorf("decode line %d: %w", i, err)}
			return
		}
		s.stream <- Entry{Tei: tei}

		i++
		// Let us how far in the file were
		// Update progress bar based on file position
		filePos, _ := file.Seek(0, 1) // Get current file position
		progress := float64(filePos) / float64(fileSize) * 100
		fmt.Printf("The current Progress is: %v%%\n", progress)
		// progressBar.SetValue(progress)
		_ = progressFloat.Set(progress)
		if int(progress) == 100 {
			time.Sleep(10 * time.Millisecond)
			// progressBar.SetValue(0)
			_ = progressFloat.Set(progress)

		}
	}

	// Read closing delimiter. `]` or `}`
	if _, err := decoder.Token(); err != nil {
		s.stream <- Entry{Error: fmt.Errorf("decode closing delimiter: %w", err)}
		return
	}
}

// Enrollment Streaming

// EnrollmentEntry represents entry on enrollment stream
type EnrollmentEntry struct {
	Error      error
	Enrollment Enrollment
}

// EnrollmentStream helps transmit each enrollment streams within a channel.
type EnrollmentStream struct {
	stream chan EnrollmentEntry
}

// EventEntry represents entry on event stream
type EventEntry struct {
	Error error
	Event Event
}

// EventStream helps transmit each event streams within a channel.
type EventStream struct {
	stream chan EventEntry
}

// NewJSONEnrollmentStream returns a new `EnrollmentStream` type.
func NewJSONEnrollmentStream() EnrollmentStream {
	return EnrollmentStream{
		stream: make(chan EnrollmentEntry),
	}
}

// Watch ....enrollment stream
func (s EnrollmentStream) Watch() <-chan EnrollmentEntry {
	return s.stream
}

// Start starts streaming JSON file line by line. If an error occurs, the channel
// will be closed.
func (s EnrollmentStream) Start(path string, ftype string, progressFloat binding.ExternalFloat) {
	// Stop streaming channel as soon as nothing left to read in the file.
	defer close(s.stream)

	// Open file to read.
	file, err := os.Open(path)
	if err != nil {
		s.stream <- EnrollmentEntry{Error: fmt.Errorf("open file: %w", err)}
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Get file size for progress calculation
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	decoder := json.NewDecoder(file)

	// Read opening delimiter. `[` or `{`
	if _, err := decoder.Token(); err != nil {
		s.stream <- EnrollmentEntry{Error: fmt.Errorf("decode opening delimiter: %w", err)}
		return
	}

	// Find the start of the array
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			// panic(err)
			s.stream <- EnrollmentEntry{Error: fmt.Errorf("decoding error: %w", err)}
		}

		if token == ftype {
			_, err := decoder.Token()
			if err != nil {
				// panic(err)
				s.stream <- EnrollmentEntry{Error: fmt.Errorf("decoding error: %w", err)}
			}

			break
		}
	}

	// Read file content as long as there is something.
	i := 1
	for decoder.More() {
		var enrollment Enrollment
		if err := decoder.Decode(&enrollment); err != nil {
			s.stream <- EnrollmentEntry{Error: fmt.Errorf("decode line %d: %w", i, err)}
			return
		}
		s.stream <- EnrollmentEntry{Enrollment: enrollment}

		i++
		// Let us how far in the file were
		// Update progress bar based on file position
		filePos, _ := file.Seek(0, 1) // Get current file position
		progress := float64(filePos) / float64(fileSize) * 100
		fmt.Printf("The current Progress is: %v%%\n", progress)
		_ = progressFloat.Set(progress)
		if int(progress) == 100 {
			time.Sleep(10 * time.Millisecond)
			_ = progressFloat.Set(0)

		}
	}

	// Read closing delimiter. `]` or `}`
	if _, err := decoder.Token(); err != nil {
		s.stream <- EnrollmentEntry{Error: fmt.Errorf("decode closing delimiter: %w", err)}
		return
	}
}

// Event Streaming

// NewJSONEventStream returns a new `EventStream` type.
func NewJSONEventStream() EventStream {
	return EventStream{
		stream: make(chan EventEntry),
	}
}

// Watch ....event stream
func (s EventStream) Watch() <-chan EventEntry {
	return s.stream
}

// Start starts streaming JSON file line by line. If an error occurs, the channel
// will be closed.
func (s EventStream) Start(path string, ftype string, progressFloat binding.ExternalFloat) {
	// Stop streaming channel as soon as nothing left to read in the file.
	defer close(s.stream)

	// Open file to read.
	file, err := os.Open(path)
	if err != nil {
		s.stream <- EventEntry{Error: fmt.Errorf("open file: %w", err)}
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Get file size for progress calculation
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	decoder := json.NewDecoder(file)

	// Read opening delimiter. `[` or `{`
	if _, err := decoder.Token(); err != nil {
		s.stream <- EventEntry{Error: fmt.Errorf("decode opening delimiter: %w", err)}
		return
	}

	// Find the start of the array
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			// panic(err)
			s.stream <- EventEntry{Error: fmt.Errorf("decoding error: %w", err)}
		}

		if token == ftype {
			_, err := decoder.Token()
			if err != nil {
				// panic(err)
				s.stream <- EventEntry{Error: fmt.Errorf("decoding error: %w", err)}
			}

			break
		}
	}

	// Read file content as long as there is something.
	i := 1
	for decoder.More() {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			s.stream <- EventEntry{Error: fmt.Errorf("decode line %d: %w", i, err)}
			return
		}
		s.stream <- EventEntry{Event: event}

		i++
		// Let us how far in the file were
		// Update progress bar based on file position
		filePos, _ := file.Seek(0, 1) // Get current file position
		progress := float64(filePos) / float64(fileSize) * 100
		fmt.Printf("The current Progress is: %v%%\n", progress)
		_ = progressFloat.Set(progress)
		if int(progress) == 100 {
			time.Sleep(10 * time.Millisecond)
			_ = progressFloat.Set(0)

		}
	}

	// Read closing delimiter. `]` or `}`
	if _, err := decoder.Token(); err != nil {
		s.stream <- EventEntry{Error: fmt.Errorf("decode closing delimiter: %w", err)}
		return
	}
}
