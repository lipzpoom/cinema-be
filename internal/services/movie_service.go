package services

import (
	"context"
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
)

// ErrMovieTitleExists is a sentinel error indicating a duplicate movie title
var ErrMovieTitleExists = errors.New("movie_title_exists")

type MovieService struct {
	Repo *repositories.MovieRepository
}

func NewMovieService(repo *repositories.MovieRepository) *MovieService {
	return &MovieService{Repo: repo}
}

func (s *MovieService) Create(ctx context.Context, item *models.Movie) error {

	existingMovie, err := s.Repo.GetByTitle(ctx, item.Title)
	if err != nil {
		return err
	}
	if existingMovie != nil {
		return ErrMovieTitleExists
	}

	return s.Repo.Create(ctx, item)
}

func (s *MovieService) GetAll(ctx context.Context) ([]models.Movie, error) {
	return s.Repo.GetAll(ctx)
}

func (s *MovieService) GetByID(ctx context.Context, id string) (*models.Movie, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *MovieService) Update(ctx context.Context, id string, data bson.M) error {
	return s.Repo.Update(ctx, id, data)
}

func (s *MovieService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}
