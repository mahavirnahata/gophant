package auth

import (
	gomvchttp "github.com/mahavirnahata/gophant/http"
	"github.com/mahavirnahata/gophant/session"
)

type Manager struct {
	SessionKey string
	RoleKey    string
}

func NewManager() *Manager {
	return &Manager{SessionKey: "user_id", RoleKey: "roles"}
}

func (a *Manager) Login(c *gomvchttp.Context, userID string) {
	s := session.FromContext(c)
	s.Regenerate()
	s.Set(a.SessionKey, userID)
}

func (a *Manager) Logout(c *gomvchttp.Context) {
	s := session.FromContext(c)
	s.Delete(a.SessionKey)
	s.Destroy()
}

func (a *Manager) UserID(c *gomvchttp.Context) (string, bool) {
	s := session.FromContext(c)
	if v, ok := s.Get(a.SessionKey); ok {
		if id, ok := v.(string); ok {
			return id, true
		}
	}
	return "", false
}

func (a *Manager) Check(c *gomvchttp.Context) bool {
	_, ok := a.UserID(c)
	return ok
}

func (a *Manager) SetRoles(c *gomvchttp.Context, roles []string) {
	s := session.FromContext(c)
	s.Set(a.RoleKey, roles)
}

func (a *Manager) Roles(c *gomvchttp.Context) []string {
	s := session.FromContext(c)
	if v, ok := s.Get(a.RoleKey); ok {
		switch r := v.(type) {
		case []string:
			return r
		case string:
			return []string{r}
		}
	}
	return nil
}

func (a *Manager) HasRole(c *gomvchttp.Context, role string) bool {
	for _, r := range a.Roles(c) {
		if r == role {
			return true
		}
	}
	return false
}
