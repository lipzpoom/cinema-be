package services

import (
	"context"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) Create(ctx context.Context, item *models.User) error {
	return s.Repo.Create(ctx, item)
}

func (s *UserService) GetAll(ctx context.Context) ([]models.User, error) {
	return s.Repo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, data bson.M) error {
	return s.Repo.Update(ctx, id, data)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}
