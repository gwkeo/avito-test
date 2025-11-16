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

type updater interface {
	UpdateFlag(ctx context.Context, userId string, flag bool) (models.User, error)
}

type prGetter interface {
	UsersPRList(ctx context.Context, userId string) ([]models.PullRequestShort, error)
}

type UserHandler struct {
	updater
	prGetter
}

func NewUserHandler(updater updater, prGetter prGetter) *UserHandler {
	return &UserHandler{
		updater:  updater,
		prGetter: prGetter,
	}
}

func (h *UserHandler) HandleSetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to read request body"), http.StatusBadRequest)
		return
	}

	type isActiveRequest struct {
		UserId   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	var rs *isActiveRequest
	if err := json.Unmarshal(body, &rs); err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "specified json is not valid"), http.StatusBadRequest)
		return
	}

	user, err := h.updater.UpdateFlag(ctx, rs.UserId, rs.IsActive)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.Wrap(err.Error(), "unable to update flag"), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "error while marshaling json"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *UserHandler) HandleGetReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	UserId := r.URL.Query().Get("user_id")
	if UserId == "" {
		http.Error(w, "user_id not specified", http.StatusBadRequest)
		return
	}

	prList, err := h.prGetter.UsersPRList(ctx, UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusBadRequest)
			return
		}
		http.Error(w, internal_http.Wrap(err.Error(), "unable to get PR list"), http.StatusInternalServerError)
		return
	}

	type responsePRs struct {
		UserId       string                    `json:"user_id"`
		PullRequests []models.PullRequestShort `json:"pull_requests"`
	}
	responsePRList := &responsePRs{
		UserId:       UserId,
		PullRequests: prList,
	}

	response, err := json.Marshal(responsePRList)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to marshal response"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
