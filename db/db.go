package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //import postgres
)

var db *sqlx.DB

func init() {
	//currentAppPreferences := fyne.CurrentApp().Preferences()
	//dbURI := currentAppPreferences.StringWithFallback(
	//	"Dispatcher2Db", "postgresql://postgres:postgres@localhost:5431/test_dispatcher2?sslmode=disable")
	psqlInfo := "postgresql://postgres:postgres@localhost:5432/integrator?sslmode=disable"
	//
	var err error
	db, err = ConnectDB(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(Schema)
}

// ConnectDB ...
func ConnectDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return db, nil
}

//GetDB ...
func GetDB() *sqlx.DB {
	return db
}
