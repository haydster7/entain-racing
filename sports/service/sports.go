package service

import (
	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"golang.org/x/net/context"
)

type Sports interface {
	// ListSports will return a collection of sport events
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
	// GetSport will return a single sport event matching the requested id
	GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error)
}

// sportsService implements the Sports interface.
type sportsService struct {
	sportsRepo db.SportsRepo
}

// NewSportsService instantiates and returns a new sportsService.
func NewSportsService(sportsRepo db.SportsRepo) Sports {
	return &sportsService{sportsRepo}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	events, err := s.sportsRepo.List(in.Filter, in.OrderBy)
	if err != nil {
		return nil, err
	}

	return &sports.ListEventsResponse{Events: events}, nil
}

func (s *sportsService) GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error) {
	event, err := s.sportsRepo.Get(in.Id)
	if err != nil {
		return nil, err
	}

	return &sports.GetEventResponse{Event: event}, nil
}
