package services

import (
	"context"
	"errors"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var ErrShowtimeOverlap = errors.New("showtime_overlap")

type ShowtimeService struct {
	Repo      *repositories.ShowtimeRepository
	MovieRepo *repositories.MovieRepository
}

func NewShowtimeService(repo *repositories.ShowtimeRepository, movieRepo *repositories.MovieRepository) *ShowtimeService {
	return &ShowtimeService{Repo: repo, MovieRepo: movieRepo}
}

func (s *ShowtimeService) Create(ctx context.Context, item *models.Showtime) error {

	movie, err := s.MovieRepo.GetByID(ctx, item.MovieID.Hex())
	if err != nil {
		return errors.New("movie not found")
	}

	endTime := item.StartTime.Add(time.Duration(movie.Duration) * time.Minute)
	item.EndTime = endTime

	overlap, err := s.Repo.CheckOverlap(ctx, item.TheaterNumber, item.StartTime, endTime, "")
	if err != nil {
		return err
	}
	if overlap {
		return ErrShowtimeOverlap
	}

	return s.Repo.Create(ctx, item)
}

func (s *ShowtimeService) GetAll(ctx context.Context) ([]models.Showtime, error) {
	return s.Repo.GetAll(ctx)
}

func (s *ShowtimeService) GetByID(ctx context.Context, id string) (*models.Showtime, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *ShowtimeService) Update(ctx context.Context, id string, data bson.M) error {

	hasTimeUpdate := false
	startTime := time.Time{}
	theaterNumber := ""

	if st, ok := data["start_time"]; ok {
		if str, isStr := st.(string); isStr {
			parsed, _ := time.Parse(time.RFC3339, str)
			startTime = parsed
			data["start_time"] = parsed // ensure db saves as date
		} else if t, isTime := st.(time.Time); isTime {
			startTime = t
		}
		hasTimeUpdate = true
	}
	if tn, ok := data["theater_number"]; ok {
		theaterNumber = tn.(string)
		hasTimeUpdate = true
	}

	if hasTimeUpdate {
		// get current to fill in gaps
		current, err := s.Repo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		movieID := current.MovieID.Hex()
		if mID, ok := data["movie_id"]; ok {
			movieID = mID.(string) // if they are updating movie_id as well
		}

		movie, err := s.MovieRepo.GetByID(ctx, movieID)
		if err != nil {
			return errors.New("movie not found")
		}

		if startTime.IsZero() {
			startTime = current.StartTime
		}
		if theaterNumber == "" {
			theaterNumber = current.TheaterNumber
		}

		endTime := startTime.Add(time.Duration(movie.Duration) * time.Minute)
		data["end_time"] = endTime // Store calculated end_time into DB payload

		overlap, err := s.Repo.CheckOverlap(ctx, theaterNumber, startTime, endTime, id)
		if err != nil {
			return err
		}
		if overlap {
			return ErrShowtimeOverlap
		}
	}

	return s.Repo.Update(ctx, id, data)
}

func (s *ShowtimeService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}

func (s *ShowtimeService) GetByMovieID(ctx context.Context, movieID string) ([]models.Showtime, error) {
	return s.Repo.GetByMovieID(ctx, movieID)
}

func (s *ShowtimeService) GetNowShowing(ctx context.Context, dateStr string) ([]bson.M, error) {
	// 1. Fetch all movies
	movies, err := s.MovieRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var showtimes []models.Showtime

	// 2. Fetch showtimes. If a date is provided, parse it in the local timezone (Asia/Bangkok)
	if dateStr != "" {
		loc, _ := time.LoadLocation("Asia/Bangkok")
		
		// Parse the incoming YYYY-MM-DD string into the local timezone
		parsedDate, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err == nil {
			// Start of Day (Thai Time translated to UTC)
			startOfDay := parsedDate
			// End of Day (Thai Time translated to UTC)
			endOfDay := startOfDay.Add(24*time.Hour).Add(-time.Nanosecond)

			// Get from MongoDB using accurate range boundary
			showtimes, err = s.Repo.GetByDateRange(ctx, startOfDay, endOfDay)
			if err != nil {
				return nil, err
			}
		} else {
			// Fallback if parsing fails for some reason
			showtimes, err = s.Repo.GetAll(ctx)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Fetch all if no date string provided
		showtimes, err = s.Repo.GetAll(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Filter and Group showtimes by movie
	showtimeMap := make(map[string][]models.Showtime)
	for _, st := range showtimes {
		mID := st.MovieID.Hex()
		showtimeMap[mID] = append(showtimeMap[mID], st)
	}

	// 3. Build response format mapping through showtimes without creating new models
	var result []bson.M
	for _, m := range movies {
		mID := m.ID.Hex()
		sts, exists := showtimeMap[mID]
		if !exists || len(sts) == 0 {
			continue // Skip movies that don't have showtimes for the selected date
		}

		result = append(result, bson.M{
			"id":          m.ID,
			"title":       m.Title,
			"duration":    m.Duration,
			"description": m.Description,
			"showtimes":   sts,
		})
	}

	return result, nil
}
