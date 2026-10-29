package main

import (
	"errors"
	"snippetbox.alexewards.net/internal/models"
	"snippetbox.alexewards.net/internal/validator"
	"fmt"
	"net/http"
	"strconv"
	// "strings"
	// "unicode/utf8"
	"github.com/go-playground/form/v4"
	"github.com/julienschmidt/httprouter"
)

// Define a snippetCreateForm struct to represent the form data and validation
// errors for the form fields. Note that all the struct fields are deliberately
// exported (i.e. start with a capital letter). This is because struct fields
// must be exported in order to be read by the html/template package when
// rendering the template.
type snippetCreateForm struct {
	Title       string `form:"title"`
	Content     string `form:"content"`
	Expires     int `form:"expires"`
	FieldErrors map[string]string
	validator.Validator `form:"-"` // Embed a validator.Validator type to hold the validation errors for our form fields.
}


type userSignupForm struct {
    Name     string `form:"name"`
    Email    string `form:"email"`
    Password string `form:"password"`

    validator.Validator `form:"-"`
}

// Create a new userLoginForm struct.
type userLoginForm struct {
    Email    string `form:"email"`
    Password string `form:"password"`

    validator.Validator `form:"-"`
}

func (a *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := a.snippets.Latest()
	if err != nil {
		a.serverError(w, err)
		return
	}

	data := a.newTemplateData(r)
	data.Snippets = snippets

	a.render(w, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, the values of any named parameters
	// will be stored in the request context. We'll talk about request context
	// in detail later in the book, but for now it's enough to know that you can
	// use the ParamsFromContext() function to retrieve a slice containing these
	// parameter names and values like so:
	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the "id" named
	// parameter from the slice and validate it as normal.
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// flash := app.sessionManager.PopString(r.Context(), "flash")

	data := app.newTemplateData(r)
	data.Snippet = snippet

	// data.Flash = flash

	app.render(w, http.StatusOK, "view.html", data)
}

	
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(w, http.StatusOK, "create.html", data)
}



func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form data.
	form.CheckField(
		validator.NotBlank(form.Title),
		"title",
		"This field cannot be blank",
	)

	form.CheckField(
		validator.MaxChars(form.Title, 100),
		"title",
		"This field cannot be more than 100 characters long",
	)

	form.CheckField(
		validator.NotBlank(form.Content),
		"content",
		"This field cannot be blank",
	)

	// Use the generic PermittedValue() function.
	form.CheckField(
		validator.PermittedValue(form.Expires, 1, 7, 365),
		"expires",
		"This field must equal 1, 7 or 365",
	)

	// Redisplay the form if validation fails.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(
			w,
			http.StatusUnprocessableEntity,
			"create.tmpl",
			data,
		)
		return
	}

	// Insert the snippet into the database.
	id, err := app.snippets.Insert(
		form.Title,
		form.Content,
		form.Expires,
	)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add a flash message to the session.
	app.sessionManager.Put(
		r.Context(),
		"flash",
		"Snippet successfully created!",
	)

	// Redirect to the newly created snippet.
	http.Redirect(
		w,
		r,
		fmt.Sprintf("/snippet/view/%d", id),
		http.StatusSeeOther,
	)
}



// Create a new decodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// createSnippetPost handler.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call Decode() on our decoder instance, passing the target destination as
	// the first parameter.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError. We use
		// errors.As() to check for this and raise a panic rather than returning
		// the error.
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// For all other errors, we return them as normal.
		return err
	}

	return nil
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, http.StatusOK, "signup.html", data)
}



func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	// Parse the form data.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents.
	form.CheckField(
		validator.NotBlank(form.Name),
		"name",
		"This field cannot be blank",
	)

	form.CheckField(
		validator.NotBlank(form.Email),
		"email",
		"This field cannot be blank",
	)

	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)

	form.CheckField(
		validator.NotBlank(form.Password),
		"password",
		"This field cannot be blank",
	)

	form.CheckField(
		validator.MinChars(form.Password, 8),
		"password",
		"This field must be at least 8 characters long",
	)

	// Redisplay the signup form if validation fails.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(
			w,
			http.StatusUnprocessableEntity,
			"signup.html",
			data,
		)
		return
	}

	// Try to create a new user record in the database.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			// Add an error message if the email already exists.
			form.AddFieldError(
				"email",
				"Email address is already in use",
			)

			data := app.newTemplateData(r)
			data.Form = form

			app.render(
				w,
				http.StatusUnprocessableEntity,
				"signup.html",
				data,
			)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Add a confirmation flash message to the session.
	app.sessionManager.Put(
		r.Context(),
		"flash",
		"Your signup was successful. Please log in.",
	)

	// Redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Display the login page.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)

    data.Form = userLoginForm{}

    app.render(w, http.StatusOK, "login.html", data)
}


func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Decode the form data into the userLoginForm struct.
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form data.
	form.CheckField(
		validator.NotBlank(form.Email),
		"email",
		"This field cannot be blank",
	)

	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)

	form.CheckField(
		validator.NotBlank(form.Password),
		"password",
		"This field cannot be blank",
	)

	// Redisplay the login form if validation fails.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(
			w,
			http.StatusUnprocessableEntity,
			"login.html",
			data,
		)
		return
	}

	// Check whether the credentials are valid.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			// Display a generic error for invalid credentials.
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form

			app.render(
				w,
				http.StatusUnprocessableEntity,
				"login.html",
				data,
			)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Renew the session token to prevent session fixation.
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Store the authenticated user's ID in the session.
	app.sessionManager.Put(
		r.Context(),
		"authenticatedUserID",
		id,
	)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}



func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Renew the session token.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Remove the authenticated user ID from the session.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Add a flash message to confirm logout.
	app.sessionManager.Put(
		r.Context(),
		"flash",
		"You've been logged out successfully!",
	)

	// Redirect the user to the application home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
