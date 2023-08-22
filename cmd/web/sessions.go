package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func getSession() *scs.SessionManager {
	// create a simple session that uses cookies
	// Now, create a simple session manager
	session := scs.New()
	session.Lifetime = 24 * time.Hour              // how long session will last
	session.Cookie.Persist = true                  // should cookies persists
	session.Cookie.SameSite = http.SameSiteLaxMode // that will avoid any errors with later versions of web browsers
	session.Cookie.Secure = true                   // this will encrypt the cookie. In production never ever use cookies that are not encrypted

	return session
}
