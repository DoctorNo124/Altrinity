package services

import "altrinity/api/repositories"

type AdminService struct {
	Repo *repositories.KeycloakRepo
}

func (s *AdminService) ListUsers() ([]repositories.KeycloakUser, error) {
	return s.Repo.FetchUsers()
}

func (s *AdminService) UpdateUserRole(userID, role string) error {
	return s.Repo.AssignRole(userID, role)
}
