package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type contextKey string

const contextUserKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

func (app *application) addIPToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.Background()
		// get the ip (as accurately as possible)
		// ip := r.RemoteAddr  // it is relatively accurate but not in all cases
		ip, err := getIP(r)
		if err != nil {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if len(ip) == 0 {
				ip = "unknown"
			}
			// now, add ip to context
			ctx = context.WithValue(r.Context(), contextUserKey, ip)
		} else {
			ctx = context.WithValue(r.Context(), contextUserKey, ip)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown", err
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}

	// check if the request(ip) came through a proxy
	forward := r.Header.Get("X-Forwarded-For") // if this header exists that means the request is forwarded (came through a proxy)
	if len(forward) > 0 {
		ip = forward
	}

	// this situation would probably never come to be true
	// but it is still implemented if it happens in future
	if len(ip) == 0 {
		ip = "forward"
	}

	return ip, nil
}

func (app *application) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.Session.Exists(r.Context(), "user") {
			app.Session.Put(r.Context(), "error", "Log in first!")
			// http.Redirect(w, r, "/", http.StatusSeeOther)  OR,
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		// Now, that everything OK with session values then middleware is fine and we should keep going
		next.ServeHTTP(w, r)
	})
}
