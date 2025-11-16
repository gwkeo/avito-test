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

type creator interface {
	Create(ctx context.Context, ID, name, authorID string) (models.PullRequest, error)
}

type merger interface {
	Merge(ctx context.Context, ID string) (models.PullRequest, error)
}

type reassigner interface {
	Reassign(ctx context.Context, ID, oldReviewerID string) (models.PullRequest, string, error)
}

type PRHandler struct {
	creator
	merger
	reassigner
}

func NewPRHandler(creator creator, merger merger, reassigner reassigner) *PRHandler {
	return &PRHandler{
		creator:    creator,
		merger:     merger,
		reassigner: reassigner,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to read request body"), http.StatusBadRequest)
		return
	}

	type createRequest struct {
		PRID     string `json:"pull_request_id"`
		PRName   string `json:"pull_request_name"`
		AuthorID string `json:"author_id"`
	}

	var cr createRequest
	err = json.Unmarshal(body, &cr)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), ""), http.StatusInternalServerError)
		return
	}

	pr, err := h.creator.Create(ctx, cr.PRID, cr.PRName, cr.AuthorID)
	if err != nil {
		if errors.Is(err, storage.ErrPRExists) {
			http.Error(w, internal_http.Wrap(err.Error(), "PR id already exists"), http.StatusConflict)
			return
		} else if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.Wrap(err.Error(), "unable to create PR"), http.StatusInternalServerError)
	}

	type PRCreateResponse struct {
		PR models.PullRequest `json:"pr"`
	}
	response := PRCreateResponse{
		PR: pr,
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to create JSON"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), ""), http.StatusBadRequest)
		return
	}

	type mergeRequest struct {
		PRID string `json:"pull_request_id"`
	}

	var mr mergeRequest
	err = json.Unmarshal(body, &mr)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to parse JSON"), http.StatusBadRequest)
		return
	}

	pr, err := h.merger.Merge(ctx, mr.PRID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.Wrap(err.Error(), "unable to merge pull request"), http.StatusInternalServerError)
		return
	}

	type prMergeResponse struct {
		PR models.PullRequest `json:"pr"`
	}

	prmerge := prMergeResponse{
		PR: pr,
	}

	responseBody, err := json.Marshal(prmerge)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to create JSON"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (h *PRHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to read request body"), http.StatusBadRequest)
		return
	}

	type reassignRequest struct {
		PRID          string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	var rr reassignRequest
	err = json.Unmarshal(body, &rr)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), ""), http.StatusBadRequest)
		return
	}

	pr, replacedBy, err := h.reassigner.Reassign(ctx, rr.PRID, rr.OldReviewerID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		if errors.Is(err, storage.ErrPRMerged) {
			http.Error(w, internal_http.Wrap(err.Error(), "cannot reassign on merged PR"), http.StatusConflict)
			return
		}
		http.Error(w, internal_http.Wrap(err.Error(), "unable to reassign merge reviewer"), http.StatusInternalServerError)
		return
	}

	type responsePRReassign struct {
		PR         models.PullRequest `json:"pr"`
		ReplacedBy string             `json:"replace_by"`
	}
	response := responsePRReassign{
		PR:         pr,
		ReplacedBy: replacedBy,
	}
	responseBody, err := json.Marshal(&response)
	if err != nil {
		http.Error(w, internal_http.Wrap(err.Error(), "unable to create JSON"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
