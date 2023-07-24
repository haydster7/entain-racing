package db

import (
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of events.
	List(filter *sports.ListEventsRequestFilter, order_by string) ([]*sports.Event, error)

	// Get will return an individual event
	Get(id int64) (*sports.Event, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sports repository dummy data.
func (r *sportsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

func (r *sportsRepo) Get(id int64) (*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getSportQueries()[eventsList]

	query, args = r.applyGet(query, id)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}
	return r.scanEvent(rows, time.Now())
}

func (r *sportsRepo) List(filter *sports.ListEventsRequestFilter, order_by string) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getSportQueries()[eventsList]

	query, args = r.applyDbFilter(query, filter)

	query = r.applySort(query, order_by)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	events = r.applyDerivedFilter(events, filter)

	return events, err
}

func (r *sportsRepo) applyGet(query string, id int64) (string, []interface{}) {
	var args []interface{}
	args = append(args, id)
	query += " WHERE event.id = ?"
	return query, args
}

// Apply filters that apply directly to the SQL database query
func (r *sportsRepo) applyDbFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.Sport) > 0 {
		clauses = append(clauses, "(sport.name LIKE ?)")
		args = append(args, "%"+filter.Sport+"%")
	}

	if len(filter.Team) > 0 {
		clauses = append(clauses, "(home_team LIKE ? OR away_team LIKE ?)")
		args = append(args, "%"+filter.Team+"%", "%"+filter.Team+"%")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

// Apply filters that apply to derived or calculated attribtues
func (r *sportsRepo) applyDerivedFilter(events []*sports.Event, filter *sports.ListEventsRequestFilter) []*sports.Event {
	var matchingEvents []*sports.Event

	if filter != nil && len(filter.Status) > 0 {
		for _, event := range events {
			if strings.EqualFold(event.Status, filter.Status) {
				matchingEvents = append(matchingEvents, event)
			}
		}
		events = matchingEvents
	}

	return events
}

// If order_by parameter is provided, send through to the SQL query
func (r *sportsRepo) applySort(query string, order_by string) string {

	if len(order_by) > 0 {
		sanitised_order_by := sanitiseOrderBy(order_by)
		if len(sanitised_order_by) > 0 {
			query += " ORDER BY " + sanitised_order_by
		}
	}

	return query
}

// Sanitise sort order input to prevent sql injection
func sanitiseOrderBy(order_by string) string {
	orders_in := strings.Split(order_by, ",")
	orders_out := []string{}

	for _, order_in := range orders_in {
		//Match only valid sql order by syntax
		re := regexp.MustCompile("(?i)^ *([0-9a-z_]*) *( +(asc|desc))? *$")

		//Extract column name and sort orders from order by component
		matches := re.FindStringSubmatch(order_in)

		//Only add if component is valid, otherwise skip
		if len(matches) >= 2 {

			column_name := matches[1]

			//Add sort order if one was found in the input segment
			var sort_order string
			if len(matches) >= 3 {
				sort_order = matches[2]
			} else {
				sort_order = ""
			}

			orders_out = append(orders_out, column_name+sort_order)
		}
	}

	return strings.Join(orders_out, ",")
}

func (m *sportsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sports.Event, error) {
	var events []*sports.Event
	requestTime := time.Now()

	for rows.Next() {
		event, err := m.scanEvent(rows, requestTime)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (m *sportsRepo) scanEvent(rows *sql.Rows, requestTime time.Time) (*sports.Event, error) {
	var event sports.Event
	var advertisedStart time.Time
	var duration int

	if err := rows.Scan(&event.Id, &event.HomeTeam, &event.AwayTeam, &event.Sport, &event.Location, &event.Capacity, &advertisedStart, &duration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	ts := timestamppb.New(advertisedStart)
	event.AdvertisedStartTime = ts

	endTime := timestamppb.New(advertisedStart.Add(time.Minute * time.Duration(duration)))
	event.ExpectedEndTime = endTime

	event.Status = getEventStatus(&event, &requestTime)

	return &event, nil
}

func getEventStatus(event *sports.Event, requestTime *time.Time) string {
	if event.AdvertisedStartTime.AsTime().After(*requestTime) {
		//advertised start time is in the future
		return "OPEN"
	} else if event.ExpectedEndTime.AsTime().Before(*requestTime) {
		//expected end time is in the past
		return "CLOSED"
	} else {
		//current time is between advertised start time and expected end time
		return "INPROGRESS"
	}
}
