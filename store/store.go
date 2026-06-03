package store

import (
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "sync"
)

var ErrNotFound = errors.New("object not found")

// Object holds the raw body of a stored object.
type Object struct {
    Body []byte
}

// bucket holds objects with per-bucket deduplication.
type bucket struct {
    mu       sync.RWMutex
    objects  map[string]string 
    content  map[string][]byte
}

func newBucket() *bucket {
    return &bucket{
        objects: make(map[string]string),
        content: make(map[string][]byte),
    }
}

// Store is the top-level in-memory object store.
type Store struct {
    mu      sync.RWMutex
    buckets map[string]*bucket
}

func New() *Store {
    return &Store{
        buckets: make(map[string]*bucket),
    }
}

func (s *Store) getOrCreateBucket(name string) *bucket {
    s.mu.Lock()
    defer s.mu.Unlock()
    if b, ok := s.buckets[name]; ok {
        return b
    }
    b := newBucket()
    s.buckets[name] = b
    return b
}

func hash(body []byte) string {
    h := sha256.Sum256(body)
    return hex.EncodeToString(h[:])
}

// Put stores an object in the given bucket, deduplicating by content hash.
func (s *Store) Put(bucketName, objectID string, body []byte) {
    b := s.getOrCreateBucket(bucketName)
    h := hash(body)

    b.mu.Lock()
    defer b.mu.Unlock()

    // Store content only if this hash is new in this bucket
    if _, exists := b.content[h]; !exists {
        b.content[h] = body
    }
    b.objects[objectID] = h
}

// Get retrieves an object by bucket and objectID.
func (s *Store) Get(bucketName, objectID string) ([]byte, error) {
    s.mu.RLock()
    b, ok := s.buckets[bucketName]
    s.mu.RUnlock()
    if !ok {
        return nil, ErrNotFound
    }

    b.mu.RLock()
    defer b.mu.RUnlock()

    h, ok := b.objects[objectID]
    if !ok {
        return nil, ErrNotFound
    }
    return b.content[h], nil
}

/* Delete removes an object reference from a bucket.
If no other objectID's reference the content, the content is also removed. */
func (s *Store) Delete(bucketName, objectID string) error {
    s.mu.RLock()
    b, ok := s.buckets[bucketName]
    s.mu.RUnlock()
    if !ok {
        return ErrNotFound
    }

    b.mu.Lock()
    defer b.mu.Unlock()

    h, ok := b.objects[objectID]
    if !ok {
        return ErrNotFound
    }

    delete(b.objects, objectID)

    // GC: remove content if no other objectID references this hash
    referenced := false
    for _, existingHash := range b.objects {
        if existingHash == h {
            referenced = true
            break
        }
    }
    if !referenced {
        delete(b.content, h)
    }

    return nil
}
