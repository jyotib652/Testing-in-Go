package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
	"webapp/pkg/data"
)

var pathToTemplates = "./templates/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "This is the Home page")
	// here, we're going to add something to the session and then display it, if it's exists, on the home page
	var td = make(map[string]any)

	if app.Session.Exists(r.Context(), "test") {
		msg := app.Session.GetString(r.Context(), "test")
		td["test"] = msg
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}

	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})

}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {

	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})

}

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {
	// parse template from disk
	// parsedTemplate, err := template.ParseFiles("./templates/" + t)
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())
	// log.Println(data.IP)

	// Get the error message or flash message from the session that we put into the session while handling Login() handler
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	// execute the template, passing it data, if any
	err = parsedTemplate.Execute(w, td)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// validate the data (doing some validation)
	form := NewForm(r.PostForm)
	form.Required("email", "password")
	if !form.Valid() {
		// fmt.Fprint(w, "failed validation")
		// redirect to the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid login credentials") // setting error message in session
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		// redirect to the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid login!") // setting error message in session
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// authenticate the user
	// if not authenticated then redirect with error
	if !app.authenticate(r, user, password) {
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// fmt.Fprint(w, email)

	// upon successful authentication, prevent session fixation attack [good practice straight from OWASP security advisory]
	// any time you log a user in or log a user out, you should renew the session.
	_ = app.Session.RenewToken(r.Context())

	// store success message in the session
	// redirect to some other page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!") // setting success message in session
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)

}

func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	valid, err := user.PasswordMatches(password)
	if err != nil || !valid {
		return false
	}

	app.Session.Put(r.Context(), "user", user)
	return true
}
