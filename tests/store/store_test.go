package store_test

import (
	"sync"
	"testing"

	"message-board/internal/store"
)

func TestStore_Create(t *testing.T) {
	s := store.NewStore()

	msg := s.Create("hello")
	if msg.ID != 1 {
		t.Errorf("expected ID 1, got %d", msg.ID)
	}
	if msg.Message != "hello" {
		t.Errorf("expected message 'hello', got '%s'", msg.Message)
	}
	if msg.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	msg2 := s.Create("world")
	if msg2.ID != 2 {
		t.Errorf("expected ID 2, got %d", msg2.ID)
	}
}

func TestStore_Create_Empty(t *testing.T) {
	s := store.NewStore()
	msg := s.Create("")
	if msg.Message != "" {
		t.Error("should accept empty message — validation is in the handler")
	}
}

func TestStore_List_Empty(t *testing.T) {
	s := store.NewStore()
	messages := s.List()
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestStore_List_Order(t *testing.T) {
	s := store.NewStore()
	s.Create("first")
	s.Create("second")
	s.Create("third")

	messages := s.List()
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	if messages[0].Message != "third" {
		t.Errorf("expected first element 'third', got '%s'", messages[0].Message)
	}
	if messages[1].Message != "second" {
		t.Errorf("expected second element 'second', got '%s'", messages[1].Message)
	}
	if messages[2].Message != "first" {
		t.Errorf("expected third element 'first', got '%s'", messages[2].Message)
	}
}

func TestStore_List_NonNilEmpty(t *testing.T) {
	s := store.NewStore()
	messages := s.List()
	if messages == nil {
		t.Error("expected non-nil slice for empty store")
	}
}

func TestStore_Delete_Found(t *testing.T) {
	s := store.NewStore()
	s.Create("msg1")
	s.Create("msg2")

	if !s.Delete(1) {
		t.Error("expected Delete(1) to return true")
	}

	messages := s.List()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message after delete, got %d", len(messages))
	}
	if messages[0].ID != 2 {
		t.Errorf("expected remaining message ID 2, got %d", messages[0].ID)
	}
}

func TestStore_Delete_NotFound(t *testing.T) {
	s := store.NewStore()
	if s.Delete(999) {
		t.Error("expected Delete(999) to return false")
	}
}

func TestStore_Delete_RemovesCorrectItem(t *testing.T) {
	s := store.NewStore()
	for i := 1; i <= 5; i++ {
		s.Create("msg")
	}

	s.Delete(3)

	messages := s.List()
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(messages))
	}
	for _, m := range messages {
		if m.ID == 3 {
			t.Error("message with ID 3 should have been deleted")
		}
	}
}

func TestStore_Concurrent(t *testing.T) {
	s := store.NewStore()
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			s.Create("concurrent")
		}()
	}
	wg.Wait()

	messages := s.List()
	if len(messages) != iterations {
		t.Errorf("expected %d messages, got %d", iterations, len(messages))
	}

	ids := make(map[int]bool)
	for _, m := range messages {
		if ids[m.ID] {
			t.Errorf("duplicate ID found: %d", m.ID)
		}
		ids[m.ID] = true
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= iterations; i++ {
			s.Delete(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			s.Create("during-delete")
		}
	}()
	wg.Wait()
}
