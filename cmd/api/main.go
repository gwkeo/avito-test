package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gwkeo/avito-test/internal/http/handler"
	"github.com/gwkeo/avito-test/internal/services/pull_request_service"
	"github.com/gwkeo/avito-test/internal/services/team_service"
	"github.com/gwkeo/avito-test/internal/services/user_service"
	"github.com/gwkeo/avito-test/internal/storage/postgres"
)

func main() {

	connectionString := os.Getenv("CONNECTION_STRING")
	if connectionString == "" {
		log.Fatal("CONNECTION_STRING not specified")
	}

	ctx := context.TODO()
	db, err := postgres.New(ctx, connectionString)
	if err != nil {
		log.Fatal("unable to connect to db")
	}

	usersService := user_service.New(db)
	teamsService := team_service.New(db, usersService)
	pullRequestService := pull_request_service.New(db, usersService, teamsService)

	usersHandler := handler.NewUserHandler(usersService)
	teamsHandler := handler.NewTeamsHandler(teamsService)
	pullRequestsHandler := handler.NewPRHandler(pullRequestService)

	mux := setupRoutes(*usersHandler, *teamsHandler, *pullRequestsHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("started at ")

	if err = server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(
	usersHandler handler.UserHandler,
	teamsHandler handler.TeamsHandler,
	pullRequestHandler handler.PullRequestHandler,
) http.Handler {
	mux := http.NewServeMux()

	userMux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", userMux))

	userMux.HandleFunc("/getReview", usersHandler.HandleGetReview)
	userMux.HandleFunc("/setIsActive", usersHandler.HandleSetIsActive)

	teamMux := http.NewServeMux()
	mux.Handle("/teams/", http.StripPrefix("/teams", teamMux))

	teamMux.HandleFunc("/add", teamsHandler.Add)
	teamMux.HandleFunc("/get", teamsHandler.Team)

	prMux := http.NewServeMux()
	mux.Handle("/pullRequest/", http.StripPrefix("/pullRequest", prMux))

	prMux.HandleFunc("/create", pullRequestHandler.CreatePR)
	prMux.HandleFunc("/merge", pullRequestHandler.MergePR)
	prMux.HandleFunc("/reassign", pullRequestHandler.ReassignPR)

	return mux
}
