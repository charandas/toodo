package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/pop/popmw"
	"github.com/gobuffalo/envy"
	csrf "github.com/gobuffalo/mw-csrf"
	forcessl "github.com/gobuffalo/mw-forcessl"
	i18n "github.com/gobuffalo/mw-i18n"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/packr/v2"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/toodo/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_toodo_session",
		})
		// Automatically redirect to SSL
		app.Use(forcessl.Middleware(secure.Options{
			SSLRedirect:     ENV == "production",
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		}))

		if ENV == "development" {
			app.Use(paramlogger.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.New("../locales", "../locales"), "en-US"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())

		app.GET("/", HomeHandler)

		// serve files from the public directory:
		app.Use(SetCurrentUser)
		app.Use(Authorize)

		app.GET("/users/new", UsersNew)
		app.POST("/users", UsersCreate)
		app.GET("/signin", AuthNew)
		app.POST("/signin", AuthCreate)
		app.DELETE("/signout", AuthDestroy)
		app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)

		app.Resource("/items", ItemsResource{})

		app.ServeFiles("/", assetsBox)
	}

	return app
}
