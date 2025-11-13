package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gwkeo/avito-test/internal/dto"
	"github.com/gwkeo/avito-test/internal/storage"
)

func wrap(message string, err error) string {
	return fmt.Sprintf("%s: %s", message, err.Error())
}

type updater interface {
	UpdateFlag(ctx context.Context, userId string, flag bool) (dto.User, error)
}

type prGetter interface {
	UsersPRList(ctx context.Context, userId string) ([]dto.PullRequestShort, error)
}

type UsersHandler struct {
	updater
	prGetter
}

func New(updater updater, prGetter prGetter) *UsersHandler {
	return &UsersHandler{
		updater:  updater,
		prGetter: prGetter,
	}
}

func (h *UsersHandler) HandleSetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, wrap("unable to read request body", err), http.StatusBadRequest)
		return
	}

	type isActiveRequest struct {
		UserId   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	var rs *isActiveRequest
	if err := json.Unmarshal(body, &rs); err != nil {
		http.Error(w, wrap("specified json not valid", err), http.StatusBadRequest)
	}

	user, err := h.updater.UpdateFlag(ctx, rs.UserId, rs.IsActive)
	if err != nil {
		http.Error(w, wrap("http", err), http.StatusInternalServerError)
	}

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, wrap("http", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *UsersHandler) HandleGetReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	UserId := r.URL.Query().Get("user_id")
	if UserId == "" {
		http.Error(w, "user_id not specified", http.StatusBadRequest)
	}

	prs, err := h.prGetter.UsersPRList(ctx, UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, wrap("user_id not found", err), http.StatusBadRequest)
			return
		}
		http.Error(w, wrap("pr", err), http.StatusInternalServerError)
		return
	}

	type responsetPRs struct {
		UserId       string                 `json:"user_id"`
		PullRequests []dto.PullRequestShort `json:"pull_requests"`
	}
	var rprs responsetPRs

	rprs.PullRequests = prs
	rprs.UserId = UserId
	response, err := json.Marshal(rprs)
	if err != nil {
		http.Error(w, wrap("error marshaling json", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
