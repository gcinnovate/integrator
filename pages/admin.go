package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// adminScreen loads a tab panel for admin widgets
func adminScreen(_ fyne.Window) fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabelWithStyle(
			fmt.Sprintf("Coming soon %s", fyne.CurrentApp().Preferences().StringWithFallback("Dispatcher2Db", "dispatcher2d")),
			fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
	)
	return container.NewCenter(content)
}
