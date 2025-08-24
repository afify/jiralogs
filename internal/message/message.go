package message

import "time"

// Role represents the role of a message sender
type Role string

const (
	User      Role = "user"
	Assistant Role = "assistant"
	System    Role = "system"
)

// Message represents a message in the system
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Parts     []Part    `json:"parts,omitempty"`
}

// Part represents a part of a message
type Part struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Content returns the combined content of all parts
func (m *Message) Content() MessageContent {
	return MessageContent{content: m.Content}
}

// MessageContent wraps message content
type MessageContent struct {
	content string
}

// String returns the content as string
func (mc MessageContent) String() string {
	return mc.content
}

// Service provides message operations (minimal interface for jiralogs)
type Service interface {
	// Add minimal methods as needed
}