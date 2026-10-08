package auth

import (
	"net/http"

	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/httpx"
)

// Handler exposes auth HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs an auth Handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type otpRequestBody struct {
	Phone string `json:"phone"`
}

// RequestOTP handles POST /auth/otp/request.
func (h *Handler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var body otpRequestBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.RequestOTP(r.Context(), body.Phone, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

type otpVerifyBody struct {
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

// VerifyOTP handles POST /auth/otp/verify.
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var body otpVerifyBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if body.ChallengeID == "" || body.OTP == "" {
		httpx.WriteError(w, domain.InvalidInput("challenge_id and otp are required"))
		return
	}
	res, err := h.svc.VerifyOTP(r.Context(), body.ChallengeID, body.OTP, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if body.RefreshToken == "" {
		httpx.WriteError(w, domain.InvalidInput("refresh_token is required"))
		return
	}
	res, err := h.svc.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var body refreshBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.svc.Logout(r.Context(), body.RefreshToken); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me handles GET /auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := domain.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, domain.Unauthorized("Authentication required"))
		return
	}
	user, err := h.svc.Me(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
