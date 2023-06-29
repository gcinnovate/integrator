package pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// QueueStatus ....
type QueueStatus string

const (
	QueueStatusReady     QueueStatus = "ready"
	QueueStatusPending   QueueStatus = "pending"
	QueueStatusFailed    QueueStatus = "failed"
	QueueStatusCompleted QueueStatus = "completed"
	QueueStatusExpired   QueueStatus = "expired"
)

// adminScreen loads a tab panel for admin widgets
func settings(_ fyne.Window) fyne.CanvasObject {
	preferences := fyne.CurrentApp().Preferences()
	dbURLEntry := widget.NewEntry()
	dbURLEntry.Text = preferences.StringWithFallback("Dispatcher2Db",
		"postgres://postgres:postgres@localhost/dispatcher2d?sslmode=disable")

	serverPort := widget.NewEntry()
	serverPort.Text = fmt.Sprintf("%d", preferences.IntWithFallback("ServerPort", 9191))
	maxRetries := widget.NewEntry()
	maxRetries.Text = fmt.Sprintf("%d", preferences.IntWithFallback("MaxRetries", 3))
	maxConcurrent := widget.NewEntry()
	maxConcurrent.Text = fmt.Sprintf("%d", preferences.IntWithFallback("MaxConcurrent", 5))
	requestProcessInterval := widget.NewEntry()
	requestProcessInterval.Text = fmt.Sprintf("%d", preferences.IntWithFallback("RequestProcessInterval", 5))
	useGlobalSubmissionPeriod := widget.NewCheck("", func(b bool) {

	})
	useSSL := widget.NewCheck("", func(b bool) {

	})

	defaultStatus := widget.NewSelect([]string{
		string(QueueStatusCompleted),
		string(QueueStatusExpired),
		string(QueueStatusFailed),
		string(QueueStatusPending),
		string(QueueStatusReady),
	}, func(s string) {

	})
	defaultStatus.Selected = preferences.StringWithFallback("DefaultQueueStatus", "ready")
	startOfSubmission := widget.NewSelect([]string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
		"13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, func(s string) {

	})
	startOfSubmission.Selected = fmt.Sprintf(
		"%d", preferences.IntWithFallback("StartOfSubmissionPeriod", 0))
	endOfSubmission := widget.NewSelect([]string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
		"13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23",
	}, func(s string) {

	})
	endOfSubmission.Selected = fmt.Sprintf(
		"%d", preferences.IntWithFallback("EndOfSubmissionPeriod", 23))
	endOfSubmission2 := widget.NewSlider(0, 24)
	endOfSubmission2.Step = 1
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Database URL", Widget: dbURLEntry, HintText: "The database connection URL"},
			{Text: "Server Port", Widget: serverPort, HintText: "Port for backend dispatcher2 server"},
			{Text: "Maximum Concurrent", Widget: maxConcurrent, HintText: "Maximum number of worker go routines"},
			{Text: "Maximum Retries", Widget: maxRetries, HintText: "Maximum retries for failed request"},
			{Text: "Request Process Interval", Widget: requestProcessInterval, HintText: "Interval for processing batched requests"},
			{Text: "Default Queue Status", Widget: defaultStatus, HintText: ""},
			{Text: "Use Global Submission Period", Widget: useGlobalSubmissionPeriod},
			{Text: "Start Submission Period", Widget: startOfSubmission, HintText: "Hour of day to start submission"},
			{Text: "End of Submission Period", Widget: endOfSubmission, HintText: "Hour of day to end submission"},
			{Text: "End of Submission Period", Widget: endOfSubmission2, HintText: ""},
			{Text: "Use SSL", Widget: useSSL, HintText: "Whether to use HTTPS"},
		},
		OnCancel: func() {

		},
		OnSubmit: func() {

		},
	}
	form.SubmitText = "Save"

	content := container.NewVBox(form)
	return content
}

type Item struct {
	ID    int
	Name  string
	Price float64
}

func makeRequestsTable(_ fyne.Window) fyne.CanvasObject {
	items := []Item{
		{ID: 1, Name: "Item 1", Price: 10.5},
		{ID: 2, Name: "Item 2", Price: 20.0},
		{ID: 3, Name: "Item 3", Price: 15.75},
		{ID: 4, Name: "Item 4", Price: 8.9},
		{ID: 5, Name: "Item 5", Price: 12.25},
		{ID: 6, Name: "Item 6", Price: 18.5},
		{ID: 7, Name: "Item 7", Price: 14.0},
		{ID: 8, Name: "Item 8", Price: 7.25},
		{ID: 9, Name: "Item 9", Price: 9.99},
		{ID: 10, Name: "Item 10", Price: 11.5},
		{ID: 11, Name: "Item 11", Price: 16.75},
		{ID: 12, Name: "Item 12", Price: 13.25},
	}
	itemsPerPage := 5

	// Data and widgets for pagination
	currentPage := 0
	// var totalPages int64 = 1
	// Table
	table := widget.NewTable(
		func() (int, int) {
			return itemsPerPage, 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, cellWidget fyne.CanvasObject) {
			itemIndex := currentPage*itemsPerPage + cell.Row
			if itemIndex < len(items) {
				item := items[itemIndex]
				if label, ok := cellWidget.(*widget.Label); ok {
					label.SetText(item.Name)
					// fmt.Println(item.Name)
				}
			} else {
				if label, ok := cellWidget.(*widget.Label); ok {
					label.Text = ""
				}
			}
		},
	)
	table.SetColumnWidth(0, 100)
	table.SetColumnWidth(1, 100)

	// Pagination buttons
	//prevButton := widget.NewButton("Previous", func() {
	//	if currentPage > 0 {
	//		currentPage--
	//		table.Refresh()
	//	}
	//})
	//
	//nextButton := widget.NewButton("Next", func() {
	//	totalPages := (len(items) + itemsPerPage - 1) / itemsPerPage
	//	if currentPage < totalPages-1 {
	//		currentPage++
	//		table.Refresh()
	//	}
	//	fmt.Println("totalPages: ", totalPages, " Current Page: ", currentPage, " Items pp: ", itemsPerPage)
	//})
	totalItems := len(items)
	p := GetPaginator(int64(totalItems), int64(itemsPerPage), 1, true)
	fmt.Printf("%v", p)
	fmt.Printf("16 records, 5 per page: %v", GetPaginator(16, 5, 1, true))
	fmt.Printf("23 records, 5 per page: %v", GetPaginator(23, 5, 2, true))
	fmt.Printf("30 records, 5 per page: %v", GetPaginator(30, 5, 3, true))
	// Pagination container
	paginationContainer := container.NewHBox(
		PaginationWidget(&p, &currentPage, table),
		// prevButton,
		// nextButton,
	)
	// pp := container.NewHBox(PaginationWidget(), )
	// Main container
	//mainContainer := container.NewGridWithColumns(2,
	//	table,
	//	paginationContainer,
	//)
	mainContainer := container.NewBorder(nil, paginationContainer, nil, nil, container.NewMax(table))
	// b := container.NewBorder()
	return mainContainer
}
