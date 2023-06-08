package pages

import (
	"fyne.io/fyne/v2"
)

// Page defines the data structure for Pages app
type Page struct {
	Title, Intro string
	View         func(w fyne.Window) fyne.CanvasObject
	SupportWeb   bool
}

var (
	Pages = map[string]Page{
		"tracker": {
			"Tracker Integrator",
			"Integration using JSON file with tracker objects",
			makeTrackerTab,
			true,
		},
		"cases": {
			"Cases",
			"Cases Report",
			makeTrackerTab,
			true,
		},
		"admin": {
			"Admin",
			"Administrative management console",
			adminScreen,
			true,
		},
		"users": {
			"Users",
			"Users management module",
			adminScreen,
			true,
		},
		"groups": {
			"Groups",
			"Groups management module",
			adminScreen,
			true,
		},
		"permissions": {
			"Permissions",
			"The user permissions module",
			adminScreen,
			true,
		},
	}

	PageIndex = map[string][]string{
		"":              {"tracker", "admin"},
		"admin":         {"users", "groups", "permissions"},
		"sms":           {"rapidpro", "telegram"},
		"mTracPro":      {"cases", "death", "apt", "tra"},
		"familyconnect": {"register"},
	}
)
