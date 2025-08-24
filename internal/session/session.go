package session

import (
	"context"
	"time"
)

// Session represents a user session
type Session struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CompletionTokens int64     `json:"completion_tokens"`
	PromptTokens     int64     `json:"prompt_tokens"`
}

// Service provides session operations (minimal interface for jiralogs)
type Service interface {
	Create(ctx context.Context, title string) (*Session, error)
	Get(ctx context.Context, id string) (*Session, error)
	List(ctx context.Context) ([]*Session, error)
}