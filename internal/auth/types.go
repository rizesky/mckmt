package auth

// AuthenticatedUser represents the authenticated user in the system
type AuthenticatedUser struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// HasRole checks if the user has a specific role
func (u *AuthenticatedUser) HasRole(role string) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *AuthenticatedUser) HasAnyRole(roles ...string) bool {
	for _, requiredRole := range roles {
		if u.HasRole(requiredRole) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (u *AuthenticatedUser) IsAdmin() bool {
	return u.HasRole("admin")
}

// IsOperator checks if the user is an operator
func (u *AuthenticatedUser) IsOperator() bool {
	return u.HasAnyRole("admin", "operator")
}

// IsViewer checks if the user is a viewer
func (u *AuthenticatedUser) IsViewer() bool {
	return u.HasAnyRole("admin", "operator", "viewer")
}
