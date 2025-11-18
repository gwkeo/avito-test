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

type pullRequestService interface {
	Create(ctx context.Context, ID, name, authorID string) (*models.PullRequest, error)
	Merge(ctx context.Context, ID string) (*models.PullRequest, error)
	Reassign(ctx context.Context, ID, oldReviewerID string) (*models.PullRequest, string, error)
}

type PullRequestHandler struct {
	pullRequestService
}

func NewPRHandler(pullRequestService pullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		pullRequestService: pullRequestService,
	}
}

func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to read request body", err.Error()), http.StatusBadRequest)
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
		http.Error(w, internal_http.WrapWithoutCode("unable to parse json body", err.Error()), http.StatusInternalServerError)
		return
	}

	pr, err := h.pullRequestService.Create(ctx, cr.PRID, cr.PRName, cr.AuthorID)
	if err != nil {
		if errors.Is(err, storage.ErrPRExists) {
			http.Error(w, internal_http.Wrap(err.Error(), "PR id already exists"), http.StatusConflict)
			return
		} else if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.WrapWithoutCode("unable to create PR", err.Error()), http.StatusInternalServerError)
		return
	}

	type PRCreateResponse struct {
		PR models.PullRequest `json:"pr"`
	}
	response := PRCreateResponse{
		PR: *pr,
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to create JSON", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to read body", err.Error()), http.StatusBadRequest)
		return
	}

	type mergeRequest struct {
		PRID string `json:"pull_request_id"`
	}

	var mr mergeRequest
	err = json.Unmarshal(body, &mr)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to parse JSON", err.Error()), http.StatusBadRequest)
		return
	}

	pr, err := h.pullRequestService.Merge(ctx, mr.PRID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		http.Error(w, internal_http.WrapWithoutCode("unable to merge pull request", err.Error()), http.StatusInternalServerError)
		return
	}

	type prMergeResponse struct {
		PR models.PullRequest `json:"pr"`
	}

	prmerge := prMergeResponse{
		PR: *pr,
	}

	responseBody, err := json.Marshal(prmerge)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to create JSON", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (h *PullRequestHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to read request body", err.Error()), http.StatusBadRequest)
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

	pr, replacedBy, err := h.pullRequestService.Reassign(ctx, rr.PRID, rr.OldReviewerID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, internal_http.Wrap(err.Error(), "resource not found"), http.StatusNotFound)
			return
		}
		if errors.Is(err, storage.ErrPRMerged) {
			http.Error(w, internal_http.Wrap(err.Error(), "cannot reassign on merged PR"), http.StatusConflict)
			return
		}
		http.Error(w, internal_http.WrapWithoutCode("unable to reassign merge reviewer", err.Error()), http.StatusInternalServerError)
		return
	}

	type responsePRReassign struct {
		PR         models.PullRequest `json:"pr"`
		ReplacedBy string             `json:"replace_by"`
	}
	response := responsePRReassign{
		PR:         *pr,
		ReplacedBy: replacedBy,
	}
	responseBody, err := json.Marshal(&response)
	if err != nil {
		http.Error(w, internal_http.WrapWithoutCode("unable to create JSON", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
