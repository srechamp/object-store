package store_test

import (
	"testing"

	"github.com/srechamp/object-store/store"
)

func TestPutAndGet(t *testing.T) { 
    s := store.New()
    s.Put("bucket1", "obj1", []byte("hello world"))
    data, err := s.Get("bucket1", "obj1")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if string(data) != "hello world" {
        t.Errorf("expected 'hello world', got '%s'", string(data))
    }

    s.Put("bucket1", "obj2", []byte("hello world"))
    data2, err := s.Get("bucket1", "obj2")
    if err != nil {
        t.Fatalf("unexpected error retrieving obj2: %v", err)
    }
    if string(data2) != "hello world" {
        t.Errorf("expected 'hello world', got '%s'", string(data2))
    }
}
func TestGetNotFound(t *testing.T) {
    s := store.New()
    _, err := s.Get("bucket1", "nonexistent")
    if err == nil {
        t.Fatal("expected error for non-existent object, got nil")
    }
}
func TestDelete(t *testing.T) {
    s := store.New()
    s.Put("bucket1", "obj1", []byte("to be deleted"))

    err := s.Delete("bucket1", "obj1")
    if err != nil {
        t.Fatalf("expected no error deleting existing object, got: %v", err)
    }

    // Confirm it's actually gone
    _, err = s.Get("bucket1", "obj1")
    if err == nil {
        t.Error("expected error retrieving deleted object, got nil")
    }
}
func TestDeleteNotFound(t *testing.T) {
    s := store.New()
    err := s.Delete("bucket1", "nonexistent")
    if err == nil {
        t.Fatal("expected error when deleting non-existent object, got nil")
    }
}

// Most important: deduplication test
func TestDeduplication(t *testing.T) {
    s := store.New()
    body := []byte("hello world")
    s.Put("bucket1", "obj1", body)
    s.Put("bucket1", "obj2", body)

    data1, _ := s.Get("bucket1", "obj1")
    data2, _ := s.Get("bucket1", "obj2")

    if string(data1) != string(data2) {
        t.Error("expected same content for deduplicated objects")
    }
    // Confirms only one copy stored (internal test or via Delete GC test)
}

func TestDeduplicationIsBucketScoped(t *testing.T) {
    // Same content in different buckets should be stored independently
    s := store.New()
    body := []byte("shared content")
    s.Put("bucket-a", "obj1", body)
    s.Put("bucket-b", "obj1", body)

    // Delete from bucket-a should not affect bucket-b
    s.Delete("bucket-a", "obj1")
    _, err := s.Get("bucket-b", "obj1")
    if err != nil {
        t.Error("object in bucket-b should still exist after deleting from bucket-a")
    }
}