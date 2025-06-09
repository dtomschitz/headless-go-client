package http

import (
	"fmt"
	"net/http"

	"github.com/dtomschitz/headless-go-client/example/backend/internal"
)

// Problem represents an RFC 7807 problem detail.
// It provides a standardized way to convey error information via HTTP APIs.
type Problem struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type,omitempty"`
	// Title is a short, human-readable summary of the problem type.
	Title string `json:"title"`
	// Status is the HTTP status code (e.g., 400, 404, 500).
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence of the problem.
	Instance string `json:"instance,omitempty"`
}

type ProblemOption func(*Problem)

func WithDetail(detail string) ProblemOption {
	return func(p *Problem) {
		p.Detail = detail
	}
}

func WithError(err error) ProblemOption {
	return func(p *Problem) {
		p.Detail = err.Error()
	}
}

func WithStatus(status int) ProblemOption {
	return func(p *Problem) {
		p.Status = status
	}
}

// NewProblem creates a new Problem instance.
func NewProblem(status int, title string, opts ...ProblemOption) *Problem {
	problem := &Problem{
		Title:  title,
		Status: status,
		Type:   fmt.Sprintf("about:blank"),
	}

	applyOptions(problem, opts...)

	return problem
}

func NewProblemFromError(err error, opts ...ProblemOption) (int, *Problem) {
	var problem *Problem

	if internal.IsNotFoundError(err) {
		problem = NewProblem(http.StatusNotFound, "Not Found", WithError(err))
	} else if internal.IsConflictError(err) {
		problem = NewProblem(http.StatusConflict, "Conflict", WithError(err))
	} else {
		problem = NewProblem(http.StatusInternalServerError, "Internal Server Error", WithError(err))
	}

	applyOptions(problem, opts...)

	return problem.Status, problem
}

func applyOptions(problem *Problem, opts ...ProblemOption) {
	for _, opt := range opts {
		opt(problem)
	}
}
