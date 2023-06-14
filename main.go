package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gcinnovate/integrator/pages"
	"log"
	"net/url"
)

const preferenceCurrentPage = "currentPage"

var topWindow fyne.Window

func main() {
	a := app.NewWithID("com.gcinnovate.integrator")
	logLifecycle(a)
	w := a.NewWindow("Integrator")
	topWindow = w

	appState := pages.NewAppState()
	pages.UpdateTrackerConf(appState.TrackerConf)

	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewMax()
	title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = fyne.TextWrapWord
	setPage := func(t pages.Page) {
		if fyne.CurrentDevice().IsMobile() {
			child := a.NewWindow(t.Title)
			topWindow = child
			child.SetContent(t.View(topWindow))
			child.Show()
			child.SetOnClosed(func() {
				topWindow = w
			})
			return
		}

		title.SetText(t.Title)
		intro.SetText(t.Intro)

		content.Objects = []fyne.CanvasObject{t.View(w)}
		content.Refresh()
	}

	page := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)
	if fyne.CurrentDevice().IsMobile() {
		w.SetContent(makeNav(setPage, false))
	} else {
		split := container.NewHSplit(makeNav(setPage, true), page)
		split.Offset = 0.12
		w.SetContent(split)
	}

	w.Resize(fyne.NewSize(1200, 700))
	w.CenterOnScreen()
	w.ShowAndRun()
}

func logLifecycle(a fyne.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	a.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func makeMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {

	//openSettings := func() {
	//	w := a.NewWindow("Fyne Settings")
	//	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	//	w.Resize(fyne.NewSize(480, 480))
	//	w.Show()
	//}
	//settingsItem := fyne.NewMenuItem("Settings", openSettings)
	//settingsShortcut := &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierShortcutDefault}
	//settingsItem.Shortcut = settingsShortcut
	//w.Canvas().AddShortcut(settingsShortcut, func(shortcut fyne.Shortcut) {
	//	openSettings()
	//})

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://wiki.hispuganda.org/en/iwizard/about")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://fyne.io/sponsor/")
			_ = a.OpenURL(u)
		}))

	main := fyne.NewMainMenu(
		helpMenu,
	)
	return main
}

func unsupportedPage(p pages.Page) bool {
	return !p.SupportWeb && fyne.CurrentDevice().IsBrowser()
}

func makeNav(setPage func(page pages.Page), loadPrevious bool) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return pages.PageIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := pages.PageIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			p, ok := pages.Pages[uid]
			if !ok {
				fyne.LogError("Missing page panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(p.Title)
			if unsupportedPage(p) {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{Italic: true}
			} else {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{}
			}
		},
		OnSelected: func(uid string) {
			if p, ok := pages.Pages[uid]; ok {
				if unsupportedPage(p) {
					return
				}
				a.Preferences().SetString(preferenceCurrentPage, uid)
				setPage(p)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentPage, "welcome")
		tree.Select(currentPref)
	}

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	switch sh := s.(type) {
	case *fyne.ShortcutCopy:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutCut:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutPaste:
		sh.Clipboard = w.Clipboard()
	}
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}
