package conversation

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes conversation HTTP endpoints.
type Handler struct {
	svc        *Service
	maxAudioMB int
	audio      AudioStore
}

// NewHandler constructs a conversation Handler.
func NewHandler(svc *Service, audio AudioStore, maxAudioMB int) *Handler {
	if maxAudioMB <= 0 {
		maxAudioMB = 5
	}
	return &Handler{svc: svc, maxAudioMB: maxAudioMB, audio: audio}
}

// Start handles POST /conversations.
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	lang := domain.LanguageFromContext(r.Context())
	conv, err := h.svc.GetOrStart(r.Context(), claims.UserID, lang)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, conv)
}

// Get handles GET /conversations/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	conv, err := h.svc.Get(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, conv)
}

type messageBody struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

// Message handles POST /conversations/{id}/message.
func (h *Handler) Message(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	var body messageBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	lang := body.Language
	if lang == "" {
		lang = domain.LanguageFromContext(r.Context())
	}
	res, err := h.svc.HandleText(r.Context(), claims.UserID, chi.URLParam(r, "id"), body.Text, lang)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// Voice handles POST /conversations/{id}/voice (multipart/form-data).
func (h *Handler) Voice(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	maxBytes := int64(h.maxAudioMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+4096)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		httpx.WriteError(w, domain.InvalidInput("Audio upload is too large or malformed"))
		return
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		httpx.WriteError(w, domain.InvalidInput("Missing audio file"))
		return
	}
	defer file.Close()

	audio, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		httpx.WriteError(w, domain.Internal("Could not read audio", err))
		return
	}
	if int64(len(audio)) > maxBytes {
		httpx.WriteError(w, domain.InvalidInput("Audio exceeds the maximum allowed size"))
		return
	}
	lang := r.FormValue("language")
	if lang == "" {
		lang = domain.LanguageFromContext(r.Context())
	}
	format := formatFromFilename(header.Filename)

	res, err := h.svc.HandleVoice(r.Context(), claims.UserID, chi.URLParam(r, "id"), audio, format, lang)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// TempAudio handles GET /audio/temp/{id}.
func (h *Handler) TempAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	path, ok := h.audio.Path(id)
	if !ok {
		httpx.WriteError(w, domain.NotFound("Audio not found"))
		return
	}
	f, err := os.Open(path)
	if err != nil {
		httpx.WriteError(w, domain.NotFound("Audio not found"))
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", contentTypeForExt(filepath.Ext(path)))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.Copy(w, f)
}

func formatFromFilename(name string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	switch ext {
	case "webm", "ogg", "wav", "mp3", "m4a":
		return ext
	default:
		return "webm"
	}
}

func contentTypeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".ogg":
		return "audio/ogg"
	default:
		return "application/octet-stream"
	}
}
