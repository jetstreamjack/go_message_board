// Package store provides in-memory storage for messages with thread-safe access.
package store

import (
	"sync"
	"time"
)

// Message represents a single message on the board.
//
// Fields:
//   - ID: unique identifier, auto-incremented
//   - Message: the text content
//   - CreatedAt: UTC timestamp when the message was created

type Message struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// Store holds all messages in memory with thread-safe access.
type Store struct {
	mu       sync.RWMutex
	nextID   int
	messages []Message
}

// NewStore creates a new empty store.
func NewStore() *Store {
	return &Store{
		messages: make([]Message, 0),
		nextID:   1,
	}
}

// Create adds a new message and returns it.
// The message ID is auto-incremented starting from 1.
func (s *Store) Create(text string) Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:        s.nextID,
		Message:   text,
		CreatedAt: time.Now().UTC(),
	}
	s.nextID++
	s.messages = append(s.messages, msg)
	return msg
}

// List returns all messages, newest first.
// Returns an empty slice (not nil) if no messages exist.
func (s *Store) List() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Message, len(s.messages))
	for i, m := range s.messages {
		result[len(s.messages)-1-i] = m
	}
	return result
}

// Delete removes a message by ID. Returns true if found.
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, m := range s.messages {
		if m.ID == id {
			s.messages = append(s.messages[:i], s.messages[i+1:]...)
			return true
		}
	}
	return false
}
