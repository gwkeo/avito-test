package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	internal_http "github.com/gwkeo/avito-test/internal/http"
	"github.com/gwkeo/avito-test/internal/models"
	"github.com/gwkeo/avito-test/internal/storage"
)

type teamService interface {
	Add(ctx context.Context, team *models.Team) (*models.Team, error)
	Team(ctx context.Context, teamName string) (*models.Team, error)
}

type TeamsHandler struct {
	teamService
}

func NewTeamsHandler(teamService teamService) *TeamsHandler {
	return &TeamsHandler{
		teamService: teamService,
	}
}

func (h *TeamsHandler) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to read request body", err.Error()), http.StatusBadRequest)
		return
	}

	var team models.Team
	err = json.Unmarshal(body, &team)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to parse request body", err.Error()), http.StatusBadRequest)
		return
	}

	updatedTeam, err := h.teamService.Add(ctx, &team)
	if err != nil {
		if errors.Is(err, storage.ErrTeamExists) {
			http.Error(w, internal_http.Wrap(err.Error(), "team_name already exists"), http.StatusBadRequest)
			return
		}
		http.Error(w, internal_http.WrapWithoutCode("unable to add team", err.Error()), http.StatusInternalServerError)
		return
	}

	type response struct {
		Team models.Team `json:"team"`
	}
	resp := &response{
		Team: *updatedTeam,
	}

	responseBody, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to create JSON", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func (h *TeamsHandler) Team(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, internal_http.WrapMessageOnly("team_name not specified"), http.StatusBadRequest)
		return
	}

	team, err := h.teamService.Team(ctx, teamName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.WrapWithoutCode("unable to get team", err.Error()), http.StatusInternalServerError)
	}

	responseBody, err := json.Marshal(team)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to craete JSON", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
