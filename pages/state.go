package pages

type TrackerConf struct {
	URL        string
	ObjectType string
	BatchSize  int
	Username   string
	Password   string
	JSONFile   string
}

type AppState struct {
	TrackerConf TrackerConf
}

var appState *AppState

// NewAppState create new AppState object initialized
func NewAppState() *AppState {
	appState = &AppState{
		TrackerConf: TrackerConf{
			URL:        "http://localhost.com:9191/queue?source=localhost&destination=eidsr_teis",
			ObjectType: "trackedEntityInstance",
			BatchSize:  15,
			JSONFile:   "",
			Username:   "",
			Password:   "",
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
