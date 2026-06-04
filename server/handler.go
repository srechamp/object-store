package server

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"

    "github.com/srechamp/object-store/store"
)

type Handler struct {
    store *store.Store
}

func NewHandler(s *store.Store) *Handler {
    return &Handler{store: s}
}

type putResponse struct {
    ID string `json:"id"`
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
    bucket := r.PathValue("bucket")
    objectID := r.PathValue("objectID")

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    h.store.Put(bucket, objectID, body)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(putResponse{ID: objectID})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    bucket := r.PathValue("bucket")
    objectID := r.PathValue("objectID")

    data, err := h.store.Get(bucket, objectID)
    if err != nil {
        if errors.Is(err, store.ErrNotFound) {
            // Spec requires 400 for not found (spec-compliant deviation from 404)
            http.Error(w, "object not found", http.StatusBadRequest)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    bucket := r.PathValue("bucket")
    objectID := r.PathValue("objectID")

    if err := h.store.Delete(bucket, objectID); err != nil {
        if errors.Is(err, store.ErrNotFound) {
            http.Error(w, "object not found", http.StatusBadRequest)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
