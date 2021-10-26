package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Khan/genqlient/graphql"
	"github.com/StevenACoffman/teamboard/pkg"
	"github.com/StevenACoffman/teamboard/pkg/github"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"
)

func RunServer(logger *log.Logger, graphqlClient graphql.Client) error {
	// =========================================================================
	// Start API Service
	api := NewHTTPServer(logger, graphqlClient)
	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		logger.Printf("main : API listening on %s", api.Addr)
		// listen and serve blocks until error or shutdown is called
		serverErrors <- api.ListenAndServe()
	}()

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	// listen for all interrupt signals, send them to quit channel
	signal.Notify(shutdown,
		os.Interrupt,    // interrupt = SIGINT = Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
	)

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		logger.Fatalf("error: listening and serving: %s", err)
		return err

	case <-shutdown:
		logger.Println("runServer : Start shutdown")

		// Give outstanding requests a deadline for completion.
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			logger.Printf("runServer : Graceful shutdown did not complete in %v : %v", timeout, err)
			err = api.Close()
			return err
		}
		return err
	}
}

// NewHTTPServer is factory function to initialize a new server
func NewHTTPServer(logger *log.Logger, graphqlClient graphql.Client) *http.Server {
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":3000"
	}

	s := &ServerHandler{graphqlClient: graphqlClient}
	// pass logger
	s.SetLogger(logger)

	h := &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return h
}

// ServerHandler implements type http.Handler interface, with our logger
type ServerHandler struct {
	logger *log.Logger
	mux    *http.ServeMux
	once   sync.Once
	graphqlClient graphql.Client
}

// SetLogger provides external injection of logger
func (s *ServerHandler) SetLogger(logger *log.Logger) {
	s.logger = logger
}

// ServeHTTP satisfies Handler interface, sets up the Path Routing
func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// on the first request only, lazily initialize
	s.once.Do(func() {
		if s.logger == nil {
			s.logger = log.New(os.Stdout,
				"INFO: ",
				log.Ldate|log.Ltime|log.Lshortfile)
			s.logger.Printf("Default Logger used")
		}
		s.mux = http.NewServeMux()

		// any static asset files in the pkg/asset folder will be available as
		// localhost:3000/static/
		s.mux.Handle("/static/",
			http.StripPrefix("/static/",
				http.FileServer(http.FS(pkg.AssetData))))
		s.mux.HandleFunc("/redirect", s.RedirectToHome)
		s.mux.HandleFunc("/health", HealthCheck)
		s.mux.HandleFunc("/", s.DefaultPage)
	})

	s.mux.ServeHTTP(w, r)
}

func (s *ServerHandler) DefaultPage(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	//TODO: If unset, the a page with a form to set these should be displayed
	org := req.URL.Query().Get("org")
	if org == "" {
		org = "Khan"
	}

	team := req.URL.Query().Get("team")
	if team == "" {
		team = "districts"
	}

	myLogin, err := github.GetLogin(req.Context(), s.graphqlClient)
	// TODO: this is not good error handling
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	var myOrgs []string
	myOrgs, err = github.GetOrgs(req.Context(), s.graphqlClient)
	// TODO: this is not good error handling
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	var myTeams []string

	myTeams, err = github.GetTeams(req.Context(), s.graphqlClient, myLogin, org)
	// TODO: this is not good error handling
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	var teammates []string
	teammates, err = github.GetTeamMembers(req.Context(), s.graphqlClient, org, team)
	// TODO: this is not good error handling
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println("<!-- myTeams: ", myTeams, " -->")
	fmt.Println("<!-- myOrgs: ", myOrgs, " -->")
	fmt.Println("<!-- TeamMates: ", teammates, " -->")
	pulls, pullsErr := github.GetPulls(req.Context(), s.graphqlClient, myLogin, org, team, teammates)
	if pullsErr != nil {
		fmt.Printf("%+v\n", pullsErr)
		return
	}

	t, tmplParseErr := template.ParseFS(pkg.AssetData, "assets/team-pr-template.html")
	if tmplParseErr != nil {
		fmt.Println(tmplParseErr)
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, pulls); err != nil {
		panic(err)
	}
	fragment := buf.String()
	_, err = w.Write([]byte(fragment))
	// TODO: this is not good error handling
	if err != nil {
		s.logger.Println("error writing:", err)
	}
}

// HealthCheck verifies externally that the program is still responding
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(200)
}

// RedirectToHome Will Log the Request, and respond with a HTTP 303 to redirect to /
func (s *ServerHandler) RedirectToHome(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("Redirected request %v to /", r.RequestURI)
	w.Header().Add("location", "/")
	w.WriteHeader(http.StatusSeeOther)
}
