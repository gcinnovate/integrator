package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/buger/jsonparser"
	"github.com/gcinnovate/integrator/controllers"
	"github.com/gcinnovate/integrator/models"
	"github.com/gcinnovate/integrator/pages"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = time.RFC3339
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	Dispatcher2Conf = Dispatcher2Config{
		Dispatcher2Db:             "postgres://postgres:postgres@localhost/integrator?sslmode=disable",
		MaxRetries:                3,
		MaxConcurrent:             5,
		ServerPort:                9191,
		LogDIR:                    "/tmp/intergrator.log",
		DefaultQueueStatus:        "ready",
		StartOfSubmissionPeriod:   0,
		EndOfSubmissionPeriod:     23,
		UseGlobalSubmissionPeriod: "true",
		UseSSL:                    "false",
		RequestProcessInterval:    4,
	}
}

// RequestObj is our object used by consumers
type RequestObj struct {
	ID                 models.RequestID     `db:"id"`
	Source             int                  `db:"source"`
	Destination        int                  `db:"destination"`
	Body               string               `db:"body"`
	Retries            int                  `db:"retries"`
	InSubmissionPeriod bool                 `db:"in_submission_period"`
	ContentType        string               `db:"ctype"`
	ObjectType         string               `db:"object_type"`
	BodyIsQueryParams  bool                 `db:"body_is_query_param"`
	SubmissionID       int64                `db:"submissionid"`
	URLSurffix         string               `db:"url_suffix"`
	Suspended          bool                 `db:"suspended"`
	Status             models.RequestStatus `db:"status"`
	StatusCode         string               `db:"statuscode"`
	Errors             string               `db:"errors"`
}

const updateRequestSQL = `
UPDATE requests SET (status, statuscode, errors, retries, updated)
	= (:status, :statuscode, :errors, :retries, timeofday()::::timestamp) WHERE id = :id
`
const updateStatusSQL = `
	UPDATE requests SET (status,  updated) = (:status, timeofday()::::timestamp)
	WHERE id = :id`

// updateRequest is used by consumers to update request in the db
func (r *RequestObj) updateRequest(tx *sqlx.Tx) {
	_, err := tx.NamedExec(updateRequestSQL, r)
	if err != nil {
		log.WithError(err).Error("Error updating request status")
	}
}

// updateRequestStatus
func (r *RequestObj) updateRequestStatus(tx *sqlx.Tx) {
	_, err := tx.NamedExec(updateStatusSQL, r)
	if err != nil {
		log.WithError(err).Error("Error updating request")
	}
}
func (r *RequestObj) withStatus(s models.RequestStatus) *RequestObj { r.Status = s; return r }

func (r *RequestObj) canSendRequest(tx *sqlx.Tx, server models.Server) bool {
	// check if we have exceeded retries
	if r.Retries > Dispatcher2Conf.MaxRetries {
		r.Status = models.RequestStatusExpired
		r.updateRequestStatus(tx)
		return false
	}
	// check if we're  suspended
	if server.Suspended() {
		log.WithFields(log.Fields{
			"server": server.ID(),
			"name":   server.Name(),
		}).Info("Destination server is suspended")
		return false
	}
	// check if we're out of submission period
	if !r.InSubmissionPeriod {
		log.WithFields(log.Fields{
			"server": server.ID(),
			"name":   server.Name(),
		}).Info("Destination server out of submission period")
		return false
	}
	// check if this request is  blacklisted
	if r.Suspended {
		r.Errors = "Blacklisted"
		r.StatusCode = "ERROR7"
		r.Retries += 1
		r.Status = models.RequestStatusCanceled
		r.updateRequest(tx)
		log.WithFields(log.Fields{
			"request": r.ID,
		}).Info("Request blacklisted")
		return false
	}
	// check if body is empty
	if len(strings.TrimSpace(r.Body)) == 0 {
		r.Status = models.RequestStatusFailed
		r.StatusCode = "ERROR1"
		r.Errors = "Request has empty body"
		r.updateRequest(tx)
		log.WithFields(log.Fields{
			"request": r.ID,
		}).Info("Request has empty body")
		return false
	}
	return true
}

func (r *RequestObj) unMarshalBody() (interface{}, error) {
	var data interface{}
	switch r.ObjectType {
	case "DATA_VALUES":
		data = models.DataValuesRequest{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}
	case "BULK_DATA_VALUES":
		data = models.BulkDataValuesRequest{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}
	case "TRACKED_ENTITIES":
		data = pages.TeisPayload{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}
	case "EVENTS":
		data = pages.EventsPayload{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}
	case "ENROLLMENTS":
		data = pages.EnrollmentsPayload{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}
	default:
		data = map[string]interface{}{}
		err := json.Unmarshal([]byte(r.Body), &data)
		if err != nil {
			return nil, err
		}

	}
	return data, nil
}

// sendRequest sends request to destination server
func (r *RequestObj) sendRequest(destination models.Server) (*http.Response, error) {
	data, err := r.unMarshalBody()
	if err != nil {
		return nil, err
	}
	marshalled, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal request body")
		return nil, err
	}
	req, err := http.NewRequest(destination.HTTPMethod(), destination.URL(), bytes.NewReader(marshalled))

	switch destination.AuthMethod() {
	case "Token":
		// Add API token
		tokenAuth := "ApiToken " + destination.AuthToken()
		req.Header.Set("Authorization", tokenAuth)
		log.WithField("AuthToken", tokenAuth).Info("The authentication token:")
	default: // Basic Auth
		// Add basic authentication
		auth := destination.Username() + ":" + destination.Password()
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", basicAuth)

	}

	req.Header.Set("Content-Type", r.ContentType)
	// Create custom transport with TLS settings
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// Set any necessary TLS settings here
			// For example, to disable certificate validation:
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func produce(db *sqlx.DB, jobs chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Producer staring:!!!")
	for {
		log.Println("Going to read requests")
		rows, err := db.Queryx(`
                SELECT id FROM requests WHERE status = $1 ORDER BY created LIMIT 100000
                `, "ready")
		if err != nil {
			log.Fatalln(err)
		}

		for rows.Next() {
			var requestID int
			err := rows.Scan(&requestID)
			if err != nil {
				log.Fatalln("==>", err)
			}
			// log.Printf("Adding request [id: %v]\n", requestID)

			go func() {
				jobs <- requestID
			}()
			log.Printf("Added Request [id: %v]\n", requestID)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
		}
		rows.Close()

		log.Println("Fetch Requests")
		log.Printf("Going to sleep for: %v", Dispatcher2Conf.RequestProcessInterval)
		// Not good enough but let's bare with the sleep this initial version
		time.Sleep(
			time.Duration(Dispatcher2Conf.RequestProcessInterval) * time.Second)
	}
}

// consume is the consumer go routine
func consume(db *sqlx.DB, worker int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Calling Consumer")

	for req := range jobs {
		fmt.Printf("Message %v is consumed by worker %v.\n", req, worker)

		reqObj := RequestObj{}
		tx := db.MustBegin()
		err := tx.QueryRowx(`
                SELECT
                        id, source, destination, body, retries, in_submission_period(destination),
                        ctype, object_type, body_is_query_param, submissionid, url_suffix,suspended,
                        statuscode, status, errors
                        
                FROM requests
                WHERE id = $1 FOR UPDATE NOWAIT`, req).StructScan(&reqObj)
		if err != nil {
			log.WithError(err).Error("Error reading request for processing")
		}
		log.WithFields(log.Fields{
			"worker":     worker,
			"request-ID": req}).Info("Handling Request")
		/* Work on the request */
		// dest = utils.GetServer(reqObj.Destination)
		log.WithFields(log.Fields{"servers": models.ServerMap}).Info("Servers")
		if server, ok := models.ServerMap[strconv.Itoa(reqObj.Destination)]; ok {
			fmt.Printf("Found Server Config: %v, URL: %s\n", server, server.URL())
			if reqObj.canSendRequest(tx, server) {
				log.WithFields(log.Fields{"request": reqObj.ID}).Info("Request can be processed")
				// send request
				resp, err := reqObj.sendRequest(server)
				if err != nil {
					log.WithError(err).WithField("RequestID", reqObj.ID).Error(
						"Failed to send request")
					reqObj.Status = models.RequestStatusFailed
					reqObj.StatusCode = "ERROR02"
					reqObj.Errors = "Server possibly unreachable"
					reqObj.Retries += 1
					reqObj.updateRequest(tx)
					return
				}

				if !server.UseAsync() {
					result := models.ImportSummary{}
					json.NewDecoder(resp.Body).Decode(&result)
					if resp.StatusCode/100 == 2 {
						reqObj.withStatus(models.RequestStatusCompleted).updateRequestStatus(tx)
						log.WithFields(log.Fields{
							"status":      result.Response.Status,
							"description": result.Response.Description,
							"importCount": result.Response.ImportCount,
							"conflicts":   result.Response.Conflicts,
						}).Info("Request completed successfully!")
						// fmt.Printf("Request Completed Successfully: %v\n", result)
					}
				} else {
					// var result map[string]interface{}
					// json.NewDecoder(resp.Body).Decode(&result)
					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						reqObj.withStatus(models.RequestStatusFailed).updateRequestStatus(tx)
						log.WithError(err).Error("Could not read response")
						return
					}
					log.WithField("responseBytes", bodyBytes).Info("Response Payload")
					if resp.StatusCode/100 == 2 {
						v, _, _, _ := jsonparser.Get(bodyBytes, "status")
						fmt.Println(v)
					}

				}
				resp.Body.Close()
			}

		} else {
			log.WithFields(log.Fields{"server": reqObj.Destination}).Info(
				"Failed to load server configuration")
		}

		tx.Commit()
	}

}

const preferenceCurrentPage = "currentPage"

var topWindow fyne.Window

func main() {

	a := app.NewWithID("com.gcinnovate.integrator")
	logLifecycle(a)
	prefs := a.Preferences()

	pages.SaveStructPreferences(prefs, Dispatcher2Conf)

	m, err := migrate.New(
		"file://db/migrations",
		Dispatcher2Conf.Dispatcher2Db)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Error running migration:", err)
	}
	dbConn, err := sqlx.Connect("postgres", Dispatcher2Conf.Dispatcher2Db)
	if err != nil {
		log.Fatalln(err)
	}
	jobs := make(chan int)
	var wg sync.WaitGroup

	// Start the producer goroutine
	wg.Add(1)
	go produce(dbConn, jobs, &wg)

	// Start the consumer goroutine
	wg.Add(1)
	go startConsumers(jobs, &wg, a)

	go startAPIServer()
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

	go func() {
		wg.Wait()
		a.Quit()
	}()
	w.ShowAndRun()
	// Wait for all goroutines to finish
}

func logLifecycle(a fyne.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
		os.Exit(1)
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

func startAPIServer() {
	// defer wg.Done()
	router := gin.Default()
	// done := make(chan bool)
	v2 := router.Group("/api", BasicAuth())
	{
		v2.GET("/test2", func(c *gin.Context) {
			c.String(200, "Authorized")
		})

		q := new(controllers.QueueController)
		v2.POST("/queue", q.Queue)
		v2.GET("/queue", q.Requests)
		v2.GET("/queue/:id", q.GetRequest)
		v2.DELETE("/queue/:id", q.DeleteRequest)

	}
	// Handle error response when a route is not defined
	router.NoRoute(func(c *gin.Context) {
		c.String(404, "Page Not Found!")
	})

	router.Run(":" + fmt.Sprintf("%d", Dispatcher2Conf.ServerPort))
}

func startConsumers(jobs <-chan int, wg *sync.WaitGroup, app fyne.App) {
	defer wg.Done()

	preferences := app.Preferences()
	consumerCount := preferences.IntWithFallback("MaxConcurrent", 5)
	dbURI := preferences.StringWithFallback(
		"Dispatcher2Db", "postgres://postgres:postgres@localhost/integrator?sslmode=disable")

	fmt.Printf("Going to create %d Consumers!!!!!\n", consumerCount)
	for i := 1; i <= consumerCount; i++ {

		newConn, err := sqlx.Connect("postgres", dbURI)
		if err != nil {
			log.Fatalln("Request processor failed to connect to database: %v", err)
		}
		fmt.Printf("Adding Consumer: %d\n", i)
		wg.Add(1)
		go consume(newConn, i, jobs, wg)
	}
	log.WithFields(log.Fields{"MaxConsumers": consumerCount}).Info("Created Consumers: ")
}
