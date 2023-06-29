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
		"dispatcher2": {
			"Dispatcher2-Go",
			"Dispatcher2 Management Console",
			adminScreen,
			true,
		},
		"apps": {
			"App Settings",
			"Applications or Endpoints management module",
			adminScreen,
			true,
		},
		"requests": {
			"Requests",
			"Requests management module",
			makeRequestsTable,
			true,
		},
		"config": {
			"Dispatcher2 Configuration",
			"Configuration",
			settings,
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
			"User management",
			adminScreen,
			true,
		},
		"groups": {
			"Group",
			"Group management",
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
		"":            {"tracker", "dispatcher2", "admin"},
		"admin":       {"users", "groups", "permissions"},
		"dispatcher2": {"apps", "requests", "config"},
		"sms":         {"rapidpro", "telegram"},
		"mTracPro":    {"cases", "death", "apt", "tra"},
	}
)
