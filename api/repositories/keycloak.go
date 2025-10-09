package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type KeycloakRepo struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

type KeycloakUser struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Enabled   bool     `json:"enabled"`
	Roles     []string `json:"roles,omitempty"`
}

// internal helper: get an admin access token
func (r *KeycloakRepo) getAdminToken() (string, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", r.BaseURL, r.Realm)

	resp, err := http.PostForm(url, map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {r.ClientID},
		"client_secret": {r.ClientSecret},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

// FetchUserCompositeRoles fetches all effective (composite) realm roles for a given user
func (r *KeycloakRepo) FetchUserRoles(userID string) ([]string, error) {
	token, err := r.getAdminToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm/composite", r.BaseURL, r.Realm, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get composite roles: %s", string(body))
	}

	var roles []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	return roleNames, nil
}

// FetchUserByID fetches a single user by ID (with roles)
func (r *KeycloakRepo) FetchUserByID(userID string) (*KeycloakUser, error) {
	token, err := r.getAdminToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", r.BaseURL, r.Realm, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user: %s", string(body))
	}

	var user KeycloakUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}

	// enrich with roles
	roles, err := r.FetchUserRoles(user.ID)
	if err == nil {
		user.Roles = roles
	}

	return &user, nil
}

// FetchUsers lists users and enriches them with their roles
func (r *KeycloakRepo) FetchUsers() ([]KeycloakUser, error) {
	token, err := r.getAdminToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/admin/realms/%s/users", r.BaseURL, r.Realm)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var users []KeycloakUser
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}

	for i, u := range users {
		roles, err := r.FetchUserRoles(u.ID)
		if err == nil {
			users[i].Roles = roles
		}
	}

	return users, nil
}

// AssignRole assigns a realm role to a user and removes the default composite role
func (r *KeycloakRepo) AssignRole(userID, roleName string) error {
	token, err := r.getAdminToken()
	if err != nil {
		return err
	}

	// 1. Fetch the full role object
	roleURL := fmt.Sprintf("%s/admin/realms/%s/roles/%s", r.BaseURL, r.Realm, roleName)
	req, _ := http.NewRequest("GET", roleURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch role %q: %s", roleName, string(b))
	}

	var roleObj struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roleObj); err != nil {
		return fmt.Errorf("failed to decode role response: %w", err)
	}

	// 2. Assign the new role to the user
	assignURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", r.BaseURL, r.Realm, userID)
	assignBody, _ := json.Marshal([]map[string]interface{}{
		{"id": roleObj.ID, "name": roleObj.Name},
	})

	req, _ = http.NewRequest("POST", assignURL, bytes.NewBuffer(assignBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to assign role %q: %s", roleName, string(b))
	}

	// 3. Remove the default composite role from the user
	defaultRoleName := fmt.Sprintf("default-roles-%s", r.Realm)
	defaultRoleURL := fmt.Sprintf("%s/admin/realms/%s/roles/%s", r.BaseURL, r.Realm, defaultRoleName)

	req, _ = http.NewRequest("GET", defaultRoleURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var defRoleObj struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&defRoleObj); err == nil {
			removeURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", r.BaseURL, r.Realm, userID)
			removeBody, _ := json.Marshal([]map[string]interface{}{
				{"id": defRoleObj.ID, "name": defRoleObj.Name},
			})

			req, _ = http.NewRequest("DELETE", removeURL, bytes.NewBuffer(removeBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err = http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode >= 300 {
					b, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to remove default role: %s", string(b))
				}
			}
		}
	}

	return nil
}
