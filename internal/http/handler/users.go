package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gwkeo/avito-test/internal/dto"
)

type updater interface {
	Update(ctx context.Context, userId string, flag bool) (dto.User, error)
}

type flagGetter interface {
	Flag(ctx context.Context, userId string)
}

type UsersHandler struct {
	updater
	flagGetter
}

func New(updater updater, flagGetter flagGetter) *UsersHandler {
	return &UsersHandler{
		updater:    updater,
		flagGetter: flagGetter,
	}
}

func (h *UsersHandler) HandleSetIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("http: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	type requested struct {
		UserId   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	var rs *requested
	if err := json.Unmarshal(body, &rs); err != nil {
		http.Error(w, fmt.Sprintf("http: %s", err.Error()), http.StatusBadRequest)
	}

	user, err := h.updater.Update(ctx, rs.UserId, rs.IsActive)
	if err != nil {
		http.Error(w, fmt.Sprintf("http: %s", err.Error()), http.StatusInternalServerError)
	}

	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *UsersHandler) HandleGetReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user_id := r.URL.Query().Get("user_id")
	if user_id == "" {
		http.Error(w, "user_id not specified", http.StatusBadRequest)
	}

}
