package handler

import (
	"net/http"
)

const (
	ReadPermission  = "read"
	WritePermission = "write"

	AdminRole = "admin"
	UserRole  = "user"
)

var (
	rolePermissions = map[string][]string{
		AdminRole: {ReadPermission, WritePermission},
		UserRole:  {ReadPermission},
	}
)

var (
	userRoles = map[string][]string{
		"User":  {UserRole},
		"Admin": {AdminRole},
	}
)

func (h *Handler) Verify(token string, permission string, w http.ResponseWriter) (autorization bool) {
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		h.Log.Error("Пользователь не авторизован")
		return false
	}

	role, err := h.S.CheckToken(token, h.Ctx)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		h.Log.Error("Пользователь не имеет доступа", err)
		return false
	}

	for _, roles := range userRoles[role] {
		for _, storedPermission := range rolePermissions[roles] {
			if permission == storedPermission {
				h.Log.Info("Пользователь успешно авторизован")
				return true
			}
		}
	}

	w.WriteHeader(http.StatusForbidden)
	h.Log.Error("Пользователь не имеет доступа")
	return false
}
