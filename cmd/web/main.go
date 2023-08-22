package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/db"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DSN string
	// DB      *sql.DB
	DB      db.PostgresConn
	Session *scs.SessionManager
}

func main() {
	// register our user (user type) in the session. It would return a pointer to data.User
	// now our user type is available in the whole application.
	gob.Register(data.User{})

	// set up an app config
	app := application{}

	// Get the DSN from terminal as an input when running the app or binary from command line. If the user don't provide it
	// then it will use the default value which is already provided in the next line.
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres Connection")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	// close the connections gracefully when we exit the application (main function)
	defer conn.Close()

	// app.DB = conn
	app.DB = db.PostgresConn{DB: conn}

	// get a session manager
	app.Session = getSession()

	// get application routes
	// mux := app.routes()

	// print out a message
	log.Println("starting server on port 8080...")

	// start the server
	// err := http.ListenAndServe(":8080", mux)
	err = http.ListenAndServe(":8080", app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
