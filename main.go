package main

import (
	"errors"
	"fmt"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"html/template"
	"net/http"
	"os"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()))
	router.GET("/", indexHandler)
	router.Use(newErrorMiddleware).
		Use(newCorsMiddleware([]string{"http://localhost:8080"})).
		WithGroup("/api/v1/", func(g *bunrouter.Group) {
			g.GET("/error", failingHandler)
		})
	
	err := http.ListenAndServe(":8080", router)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func newCorsMiddleware(allowedOrigins []string) bunrouter.MiddlewareFunc {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
	})

	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return bunrouter.HTTPHandler(corsHandler.Handler(next))
	}
}

func newErrorMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
	}
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

var indexTmpl = `
<html>
	<ul>
		<li><a ref="/api/v1/users/123">/api/v1/users/123</a></li>
		<li><a ref="/api/v1/error">/api/v1/error</a></li>
		
		<li><a ref="/api/v1/users/123">/api/v1/users/123</a></li>
		<li><a ref="/api/v1/error">/api/v1/error</a></li>
	</ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}

func failingHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return errors.New("Just an error")
}
