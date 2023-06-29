package pages

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	filePath   = "No file selected"
	objectType = map[string]string{
		"Tracked Entities": "trackedEntityInstances",
		"Enrollments":      "enrollments",
		"Events":           "events",
		"Tracked Entities + Enrollments + Events": "events+enrollments",
		"Tracker NESTED Payload":                  "nested",
		"Aggregate":                               "dataValues",
	}
)

func makeTrackerTab(w fyne.Window) fyne.CanvasObject {
	f := 0.0
	progressFloat := binding.BindFloat(&f)
	state := GetAppState()
	t := state.TrackerConf
	// Create data bindings for the form fields
	usernameBinding := binding.BindString(&state.TrackerConf.Username)
	authMethodBinding := binding.BindString(&state.TrackerConf.AuthMethod)
	tokenBinding := binding.BindString(&state.TrackerConf.Token)
	queueServerBinding := binding.BindString(&state.TrackerConf.URL)

	progressBar := widget.NewProgressBarWithData(progressFloat)
	progressBar.Min = 0
	progressBar.Max = 100
	// item := container.NewVBox(bar)
	categorySelect := widget.NewSelect([]string{
		"Tracked Entities",
		"Enrollments",
		"Events",
		"Tracked Entities + Enrollments + Events",
		"Tracker NESTED Payload",
		"Aggregate"},
		func(s string) {
			t.ObjectType = s
			UpdateTrackerConf(t)
		})
	categorySelect.PlaceHolder = "Select Tracker Object"
	categorySelect.SetSelected(state.TrackerConf.ObjectType)

	numberPerBatch := newNumEntry()
	numberPerBatch.SetPlaceHolder("Number of items per batch")
	numberPerBatch.OnChanged = func(s string) {
		i, err := strconv.Atoi(s)
		if err == nil {
			t.BatchSize = i
			UpdateTrackerConf(t)

		}
	}
	numberPerBatch.SetText(strconv.Itoa(state.TrackerConf.BatchSize))

	destination := widget.NewSelect([]string{
		"Queuing Server",
		"DHIS2 < 2.40",
		"DHIS2 2.40 and above",
	}, func(s string) {
		t.Destination = s
		UpdateTrackerConf(t)
	})
	destination.PlaceHolder = "Select Destination Type"
	destination.SetSelected(state.TrackerConf.Destination)

	queueServer := widget.NewEntryWithData(queueServerBinding)
	queueServer.SetPlaceHolder("messaging server queue endpoint")
	// queueServer.SetText(state.TrackerConf.URL)
	queueServer.Validator = validation.NewRegexp(
		// `[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`,
		`http(s)?:\/\/(www\.)?[a-zA-Z0-9\-@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`,
		"not a valid url")
	queueServerBinding.AddListener(binding.NewDataListener(func() {
		n, err := queueServerBinding.Get()
		if err == nil {
			state.TrackerConf.URL = n
			UpdateTrackerConf(state.TrackerConf)
		}
	}))

	authMethod := widget.NewSelect([]string{
		"Basic Authentication",
		"Personal Access Token",
		"JSON Web Token",
	}, func(s string) {
		t.AuthMethod = s
		UpdateTrackerConf(t)
		_ = authMethodBinding.Set(s)
	})
	authMethod.PlaceHolder = "Select Authentication Method"
	authMethod.SetSelected(state.TrackerConf.AuthMethod)

	token := widget.NewEntryWithData(tokenBinding)
	token.SetPlaceHolder("Auth Token")
	// token.Validator = validation.NewRegexp(`\w`, "missing username")
	tokenBinding.AddListener(binding.NewDataListener(func() {
		n, _ := tokenBinding.Get()
		state.TrackerConf.Token = n
		UpdateTrackerConf(state.TrackerConf)
	}))

	username := widget.NewEntryWithData(usernameBinding)
	username.SetPlaceHolder("Username")
	usernameBinding.AddListener(binding.NewDataListener(func() {
		n, err := usernameBinding.Get()
		if err == nil {
			state.TrackerConf.Username = n
			UpdateTrackerConf(state.TrackerConf)
		}
	}))

	// Used for status of the batching process
	startTimeBinding := binding.BindString(&state.TrackerConf.Details.StartTime)
	endTimeBinding := binding.BindString(&state.TrackerConf.Details.EndTime)
	objectsBinding := binding.BindString(&state.TrackerConf.Details.Objects)
	batchBinding := binding.BindString(&state.TrackerConf.Details.Batched)
	successBinding := binding.BindString(&state.TrackerConf.Details.Success)
	failedBinding := binding.BindString(&state.TrackerConf.Details.Failed)

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")
	// password.Validator = validation.NewRegexp(`\w{1,}`, "missing password")

	fileLabel := widget.NewLabel(filePath)
	uploadButton := widget.NewButton("Choose File to Upload", func() {
		fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err == nil && file != nil {
				// Set the file label to display the selected file path
				fileLabel.SetText(file.URI().Path())
				filePath = file.URI().Path()
				fmt.Println("Chosen File is:", filePath)
				_ = file.Close()
			}
		}, w)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		fileDialog.Show()
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Tracker Object", Widget: categorySelect, HintText: "Type of tracker object"},
			{Text: "Batch Size", Widget: numberPerBatch, HintText: "Number of items per batch"},
			{Text: "Destination", Widget: destination, HintText: "Destination"},
			{Text: "Queue Sever", Widget: queueServer, HintText: "Queuing endpoint for messaging server"},
		},
		OnCancel: func() {
			fmt.Println("Cancelled")
		},
		OnSubmit: func() {
			fmt.Println("Form submitted")
			//fyne.CurrentApp().SendNotification(&fyne.Notification{
			//	Title:   "Form for: " + numberPerBatch.Text,
			//	Content: "Form Submitted Successfully",
			//})
			// Do something with the form data here
			if (password.Text == "" || username.Text == "") && authMethod.Selected == "Basic Authentication" {
				dialog.ShowInformation("Validation", "Both Username and Password should be provided", w)
				return
			}

			if authMethod.Selected == "Personal Access Token" && token.Text == "" {
				dialog.ShowInformation("Missing Token", "Please provide Personal Access Token", w)
				return
			}
			if authMethod.Selected == "JSON Web Token" && token.Text == "" {
				dialog.ShowInformation("Missing Token", "Please provide JSON Web Token", w)
				return
			}
			if filePath == "" || filePath == "No file selected" {
				dialog.ShowInformation("File Missing", "Please select a file to Process", w)
				return
			}
			destURL := queueServer.Text

			currentTime := time.Now()
			_ = startTimeBinding.Set(fmt.Sprintf("Start Time: %s",
				currentTime.Format("2006-01-02 15:04:05")))
			extraParams := url.Values{
				"year":       {currentTime.Format("2006")},
				"month":      {currentTime.Format("01")},
				"is_qparams": {"false"}, // from dispatcher2 f means POST body isn't query params
			}

			log.Println(
				"The URL is", destURL, "Username: ", username.Text, " password: ",
				password.Text, "file: ", filePath, "Ftype: ", objectType[categorySelect.Selected])
			batchSize, err := strconv.Atoi(numberPerBatch.Text)
			if err != nil {
				batchSize = 10
			}

			switch integrationType := objectType[categorySelect.Selected]; integrationType {
			case "trackedEntityInstances":
				log.Println("Streaming Tracked Entities")
				extraParams.Add("report_type", "teis")
				finalURL, err := addExtraParams(destURL, extraParams)
				if err != nil {
					fmt.Println("Error adding extra parameters:", err)
					return
				}

				stream := NewJSONTeiStream()

				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()
					var (
						objects = 0
						batches = 0
						failed  = 0
						success = 0
					)
					var payLoad []TrackedEntityInstance
					var count = 0
					var chunkSize = batchSize
					for data := range stream.Watch() {
						objects++
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Tei.TrackedEntity, ":", data.Tei.TrackedEntityType)
						payLoad = append(payLoad, data.Tei)
						if count >= chunkSize {
							batches++
							count = 0
							j, err := json.Marshal(payLoad)
							if err == nil {
								log.Println(string(j))
								//var teisPayload = TeisPayload{TrackedEntityInstances: payLoad}
								//// Let's push the payload
								//_, err := postRequest(finalURL, teisPayload, username.Text, password.Text)
								//if err != nil {
								//	log.Println("Error queuing chunk: ", err)
								//}
								e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								if e != nil {
									failed++
								} else {
									success++
								}
								time.Sleep(500 * time.Millisecond)
								payLoad = nil
								_ = batchBinding.Set(fmt.Sprintf("Batches: %d", batches))
								_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
								_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
								_ = endTimeBinding.Set(fmt.Sprintf("End Time: %s",
									time.Now().Format("2006-01-02 15:04:05")))

							}
						}
						count++
						_ = objectsBinding.Set(fmt.Sprintf("Total Tracked Entities: %d", objects))
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						fmt.Println("Working on last Batch")
						e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
						if e != nil {
							failed++
						} else {
							success++
						}
						batches++
						_ = batchBinding.Set(fmt.Sprintf("Batches: %d", batches))
						_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
						_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
						_ = endTimeBinding.Set(fmt.Sprintf("End Time: %s",
							time.Now().Format("2006-01-02 15:04:05")))
					}
				}()
				stream.Start(filePath, objectType[categorySelect.Selected], progressFloat, endTimeBinding)

				// Wait for the streaming task to complete
				wg.Wait()
			case "enrollments":
				log.Println("Streaming Tracked Enrollments")
				extraParams.Add("report_type", "enrollments")
				finalURL, err := addExtraParams(destURL, extraParams)
				if err != nil {
					fmt.Println("Error adding extra parameters:", err)
					return
				}
				stream := NewJSONEnrollmentStream()

				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()

					var (
						batches = 0
						objects = 0
						success = 0
						failed  = 0
					)
					var payLoad []Enrollment
					var count = 0
					var chunkSize = batchSize
					for data := range stream.Watch() {
						objects++
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Enrollment.EnrollmentDate, ":", data.Enrollment.Program)
						payLoad = append(payLoad, data.Enrollment)
						if count >= chunkSize {
							batches++
							count = 0
							j, err := json.Marshal(payLoad)
							if err == nil {
								log.Println(string(j))

								e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								if e != nil {
									failed++
								} else {
									success++
								}
								time.Sleep(500 * time.Millisecond)
								payLoad = nil
								_ = batchBinding.Set(fmt.Sprintf("Batched: %d", batches))
								_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
								_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
								_ = endTimeBinding.Set(fmt.Sprintf("End Time: %s",
									time.Now().Format("2006-01-02 15:04:05")))

							}
						}
						count++
						_ = objectsBinding.Set(fmt.Sprintf("Total Enrollments: %d", objects))
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
						if e != nil {
							failed++
						} else {
							success++
						}
						batches++
						_ = batchBinding.Set(fmt.Sprintf("Batched: %d", batches))
						_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
						_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
						_ = endTimeBinding.Set(fmt.Sprintf("Current Time: %s",
							time.Now().Format("2006-01-02 15:04:05")))
					}
				}()
				stream.Start(filePath, objectType[categorySelect.Selected], progressFloat, endTimeBinding)

				// Wait for the streaming task to complete
				wg.Wait()
			case "events":
				log.Println("Streaming Tracked Events")
				stream := NewJSONEventStream()
				extraParams.Add("report_type", "events")
				finalURL, err := addExtraParams(destURL, extraParams)
				if err != nil {
					fmt.Println("Error adding extra parameters:", err)
					return
				}
				var wg sync.WaitGroup
				wg.Add(1)

				// done := make(chan bool)
				go func() {
					defer wg.Done()

					var (
						objects = 0
						batches = 0
						success = 0
						failed  = 0
						count   = 0
					)
					var payLoad []Event
					var chunkSize = batchSize
					for data := range stream.Watch() {
						objects++
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Event.EventDate, ":", data.Event.Program)
						payLoad = append(payLoad, data.Event)
						if count >= chunkSize {
							batches++
							count = 0
							j, err := json.Marshal(payLoad)
							if err == nil {
								log.Println(string(j))

								e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								if e != nil {
									failed++
								} else {
									success++
								}
								time.Sleep(500 * time.Millisecond)
								payLoad = nil

							}
							_ = batchBinding.Set(fmt.Sprintf("Batches: %d", batches))
							_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
							_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
							_ = endTimeBinding.Set(fmt.Sprintf("Current Time: %s",
								time.Now().Format("2006-01-02 15:04:05")))
						}
						count++
						_ = objectsBinding.Set(fmt.Sprintf("Total Events: %d", objects))
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						e := postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
						if e != nil {
							failed++
						} else {
							success++
						}
						batches++
						_ = batchBinding.Set(fmt.Sprintf("Batched: %d", batches))
						_ = failedBinding.Set(fmt.Sprintf("Failed Batches: %d", failed))
						_ = successBinding.Set(fmt.Sprintf("Successful Batches: %d", success))
						_ = endTimeBinding.Set(fmt.Sprintf("Current Time: %s",
							time.Now().Format("2006-01-02 15:04:05")))
					}
				}()
				stream.Start(filePath, objectType[categorySelect.Selected], progressFloat, endTimeBinding)

				go func() {
					wg.Wait()
					// submitButton.Enable()
				}()
				// Wait for the streaming task to complete
				// wg.Wait()
			default:
				log.Println("Streaming Other Resources")
			}

			dialog.ShowInformation(
				"Success!",
				"Finished Processing.",
				w,
			)

		},
	}
	// form.Append("Auth Method", authMethod)

	form.AppendItem(&widget.FormItem{Text: "Auth Method", Widget: authMethod, HintText: "Authentication Method"})

	form.Append("Username", username)
	form.Append("Password", password)
	form.Append("Token", token)
	form.Append("JSON File", uploadButton)
	form.Append("Selected File", fileLabel)

	// form.Append("JSON File", uploadButton)
	// form.Append("Selected File", fileLabel)

	authMethodBinding.AddListener(binding.NewDataListener(func() {
		am, _ := authMethodBinding.Get()
		fmt.Printf("The auth Method is %s\n", am)
		switch am {
		case "Basic Authentication":
			// username.Validator = validation.NewRegexp(`\w`, "missing username")
			// password.Validator = validation.NewRegexp(`\w`, "missing password")
		case "Personal Access Token":
			// token.Validator = validation.NewRegexp(`\w`, "missing token")

		}

	}))

	startTime := widget.NewLabelWithData(startTimeBinding)
	startTime.TextStyle.Bold = true
	sep := widget.NewSeparator()
	objectsLabel := widget.NewLabelWithData(objectsBinding)
	objectsLabel.TextStyle.Bold = true

	batchesLabel := widget.NewLabelWithData(batchBinding)
	batchesLabel.TextStyle.Bold = true

	successLabel := widget.NewLabelWithData(successBinding)
	successLabel.TextStyle.Bold = true

	failedLabel := widget.NewLabelWithData(failedBinding)
	failedLabel.TextStyle.Bold = true
	sep2 := widget.NewSeparator()
	endTime := widget.NewLabelWithData(endTimeBinding)
	endTime.TextStyle.Bold = true

	z := container.NewVBox(
		progressBar,
		startTime,
		sep,
		objectsLabel,
		batchesLabel,
		successLabel,
		failedLabel,
		sep2,
		endTime)

	c := widget.NewCard("Processing Status", "Show processing details.", z)

	b := container.NewBorder(nil, nil, nil, nil, c)

	return container.NewGridWithColumns(2, form, b)
	// return form
}

type numEntry struct {
	widget.Entry
}

func (n *numEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}

func newNumEntry() *numEntry {
	e := &numEntry{}
	e.ExtendBaseWidget(e)
	e.Validator = validation.NewRegexp(`\d`, "Must contain a number")
	return e
}
