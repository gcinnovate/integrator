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
	filePath = "No file selected"
	m        = map[string]string{
		"Tracked Entities": "trackedEntityInstances",
		"Enrollments":      "enrollments",
		"Events":           "events",
		"Tracked Entities + Enrollments + Events": "events+enrollments",
		"Aggregate": "dataValues",
	}
)

func makeTrackerTab(w fyne.Window) fyne.CanvasObject {
	f := 0.0
	progressFloat := binding.BindFloat(&f)
	state := GetAppState()
	t := state.TrackerConf
	// Create data bindings for the form fields
	usernameBinding := binding.BindString(&state.TrackerConf.Username)
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

	username := widget.NewEntryWithData(usernameBinding)
	username.SetPlaceHolder("Username")
	username.Validator = validation.NewRegexp(`\w`, "missing username")
	usernameBinding.AddListener(binding.NewDataListener(func() {
		n, err := usernameBinding.Get()
		if err == nil {
			state.TrackerConf.Username = n
			UpdateTrackerConf(state.TrackerConf)
		}
	}))

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")
	password.Validator = validation.NewRegexp(`\w`, "missing password")

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
			destURL := queueServer.Text

			currentTime := time.Now()
			extraParams := url.Values{
				"year":       {currentTime.Format("2006")},
				"month":      {currentTime.Format("01")},
				"is_qparams": {"false"}, // from dispatcher2 f means POST body isn't query params
			}

			log.Println(
				"The URL is", destURL, "Username: ", username.Text, " password: ",
				password.Text, "file: ", filePath, "Ftype: ", m[categorySelect.Selected])
			batchSize, err := strconv.Atoi(numberPerBatch.Text)
			if err != nil {
				batchSize = 10
			}

			switch integrationType := m[categorySelect.Selected]; integrationType {
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

					var payLoad []TrackedEntityInstance
					var count = 0
					var chunkSize = batchSize
					for data := range stream.Watch() {
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Tei.TrackedEntity, ":", data.Tei.TrackedEntityType)
						payLoad = append(payLoad, data.Tei)
						if count > chunkSize {
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
								_ = postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								time.Sleep(500 * time.Millisecond)
								payLoad = nil

							}
						}
						count++
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						_ = postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
					}
				}()
				stream.Start(filePath, m[categorySelect.Selected], progressFloat)

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

					var payLoad []Enrollment
					var count = 0
					var chunkSize = batchSize
					for data := range stream.Watch() {
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Enrollment.EnrollmentDate, ":", data.Enrollment.Program)
						payLoad = append(payLoad, data.Enrollment)
						if count > chunkSize {
							count = 0
							j, err := json.Marshal(payLoad)
							if err == nil {
								log.Println(string(j))

								postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								time.Sleep(500 * time.Millisecond)
								payLoad = nil

							}
						}
						count++
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						_ = postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
					}
				}()
				stream.Start(filePath, m[categorySelect.Selected], progressFloat)

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

					var payLoad []Event
					var count = 0
					var chunkSize = batchSize
					for data := range stream.Watch() {
						if data.Error != nil {
							log.Println(data.Error)
						}
						log.Println(data.Event.EventDate, ":", data.Event.Program)
						payLoad = append(payLoad, data.Event)
						if count > chunkSize {
							count = 0
							j, err := json.Marshal(payLoad)
							if err == nil {
								log.Println(string(j))

								_ = postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
								time.Sleep(500 * time.Millisecond)
								payLoad = nil

							}
						}
						count++
					}
					if len(payLoad) > 0 {
						// Meaning batch size might have been bigger than available entities
						_ = postTrackerPayload(finalURL, payLoad, username.Text, password.Text)
					}
				}()
				stream.Start(filePath, m[categorySelect.Selected], progressFloat)

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
				"Form submitted successfully.",
				w,
			)

		},
	}
	form.Append("Username", username)
	form.Append("Password", password)
	form.Append("JSON File", uploadButton)
	form.Append("Selected File", fileLabel)
	//form.Append("Progress", bar)
	// return container.NewBorder(item, nil, nil, nil, form)
	rich := widget.NewRichTextFromMarkdown(`
# Submission Status


`)
	statsCol := container.NewVBox(progressBar, rich)
	return container.NewGridWithColumns(2, form, statsCol)
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
