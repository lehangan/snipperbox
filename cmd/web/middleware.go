package main

import (
	"context"
	"net/http"
	"fmt"

	"github.com/justinas/nosurf"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: This is split across multiple lines for readability. You don't
		// need to do this in your own code.
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		)

		w.Header().Set(
			"Referrer-Policy",
			"origin-when-cross-origin",
		)

		w.Header().Set(
			"X-Content-Type-Options",
			"nosniff",
		)

		w.Header().Set(
			"X-Frame-Options",
			"deny",
		)

		w.Header().Set(
			"X-XSS-Protection",
			"0",
		)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the built-in recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")

				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}


func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page.
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Prevent authenticated pages from being stored in the browser cache.
		w.Header().Add("Cache-Control", "no-store")

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}


func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	// Configure the CSRF cookie.
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}


func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the authenticated user ID from the session.
		id := app.sessionManager.GetInt(
			r.Context(),
			"authenticatedUserID",
		)

		// If no user ID exists, continue to the next handler.
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Check whether the user exists in the database.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// If the user exists, mark the request as authenticated.
		if exists {
			ctx := context.WithValue(
				r.Context(),
				isAuthenticatedContextKey,
				true,
			)
			r = r.WithContext(ctx)
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
