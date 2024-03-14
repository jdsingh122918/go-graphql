package main

import (
	"embed"
	"errors"
	"fmt"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"html/template"
	"net/http"
	"os"
)

//go:embed files
var fileFS embed.FS

func main() {
	fileserver := http.FileServer(http.FS(fileFS))
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()))
	router.GET("/", indexHandler)
	router.GET("/files/*path", bunrouter.HTTPHandler(fileserver))
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
  <h1>Welcome</h1>
  <ul>
    <li><a href="/files/">/files/</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}

func failingHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return errors.New("Just an error")
}
