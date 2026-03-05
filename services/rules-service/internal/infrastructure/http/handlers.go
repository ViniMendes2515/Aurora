package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"aurora/services/rules-service/internal/application"
	"aurora/services/rules-service/internal/domain"
)

// Handlers contém os handlers HTTP do rules-service
type Handlers struct {
	rulesEngine *application.RulesEngine
}

// NewHandlers cria uma nova instância de Handlers
func NewHandlers(rulesEngine *application.RulesEngine) *Handlers {
	return &Handlers{rulesEngine: rulesEngine}
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error string `json:"error"`
}

// ListRules handler para GET /rules
func (h *Handlers) ListRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	rules, err := h.rulesEngine.ListRules(userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "failed to fetch rules")
		return
	}

	h.respondWithJSON(w, http.StatusOK, rules)
}

// CreateRule handler para POST /rules
func (h *Handlers) CreateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)

	var req application.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.OwnerID = userID

	response, err := h.rulesEngine.CreateRule(req)
	if err != nil {
		switch err {
		case domain.ErrInvalidRule:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "failed to create rule")
		}
		return
	}

	h.respondWithJSON(w, http.StatusCreated, response)
}

// DeleteRule handler para DELETE /rules/{id}
func (h *Handlers) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ruleID := extractRuleID(r.URL.Path)
	if ruleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "rule ID is required")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)
	if err := h.rulesEngine.DeleteRule(ruleID, userID); err != nil {
		switch err {
		case domain.ErrRuleNotFound:
			h.respondWithError(w, http.StatusNotFound, err.Error())
		case domain.ErrRuleAccessDenied:
			h.respondWithError(w, http.StatusForbidden, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "failed to delete rule")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// UpdateRule handler para PUT /rules/{id}
func (h *Handlers) UpdateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ruleID := extractRuleID(r.URL.Path)
	if ruleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "rule ID is required")
		return
	}

	userID := r.Context().Value(UserIDKey).(string)

	var req application.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.rulesEngine.UpdateRule(ruleID, userID, req)
	if err != nil {
		switch err {
		case domain.ErrRuleNotFound:
			h.respondWithError(w, http.StatusNotFound, err.Error())
		case domain.ErrRuleAccessDenied:
			h.respondWithError(w, http.StatusForbidden, err.Error())
		case domain.ErrInvalidRule:
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			h.respondWithError(w, http.StatusInternalServerError, "failed to update rule")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, resp)
}

func extractRuleID(path string) string {
	path = strings.TrimPrefix(path, "/rules/")
	parts := strings.Split(path, "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Error: message})
}
