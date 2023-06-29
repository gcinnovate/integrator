package pages

type BatchDetails struct {
	Objects   string
	Batched   string
	Success   string
	Failed    string
	StartTime string
	EndTime   string
}

type TrackerConf struct {
	URL         string
	ObjectType  string
	BatchSize   int
	Destination string
	AuthMethod  string
	Username    string
	Password    string
	Token       string
	JSONFile    string
	Details     BatchDetails
}

type AppState struct {
	TrackerConf TrackerConf
}

var appState *AppState

// NewAppState create new AppState object initialized
func NewAppState() *AppState {
	appState = &AppState{
		TrackerConf: TrackerConf{
			URL:         "http://localhost.com:9191/api/queue?source=localhost&destination=dhis2",
			ObjectType:  "Tracked Entities",
			BatchSize:   15,
			Destination: "Queuing Server",
			JSONFile:    "",
			AuthMethod:  "Basic Authentication",
			Username:    "",
			Password:    "",
			Token:       "",
			Details: BatchDetails{
				"Total Objects: 0",
				"Batches: 0",
				"Successful Batches: 0",
				"Failed Batches: 0",
				"Start Time: ",
				"End Time: ",
			},
		},
	}
	return appState
}

// GetAppState returns a reference to the app state
func GetAppState() *AppState {
	return appState
}

func UpdateTrackerConf(t TrackerConf) {
	appState.TrackerConf = t
}
