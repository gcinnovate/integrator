package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// adminScreen loads a tab panel for admin widgets
func adminScreen(_ fyne.Window) fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Coming soon", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
	)
	return container.NewCenter(content)
}
