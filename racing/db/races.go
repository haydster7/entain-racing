package db

import (
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter, order_by string) ([]*racing.Race, error)

	// Get will return an individual race
	Get(id int64) (*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) Get(id int64) (*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyGet(query, id)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}
	return r.scanRace(rows, time.Now())
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter, order_by string) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	query = r.applySort(query, order_by)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyGet(query string, id int64) (string, []interface{}) {
	var args []interface{}
	args = append(args, id)
	query += " WHERE id = ?"
	return query, args
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	//Only apply visibility filter if it is defined, to avoid default value of false
	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		args = append(args, filter.GetVisible())
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

// If order_by parameter is provided, send through to the SQL query
func (r *racesRepo) applySort(query string, order_by string) string {

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

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race
	requestTime := time.Now()

	for rows.Next() {
		race, err := m.scanRace(rows, requestTime)
		if err != nil {
			return nil, err
		}
		races = append(races, race)
	}

	return races, nil
}

func (m *racesRepo) scanRace(rows *sql.Rows, requestTime time.Time) (*racing.Race, error) {
	var race racing.Race
	var advertisedStart time.Time

	if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	race.Status = getRaceStatus(&advertisedStart, &requestTime)

	ts := timestamppb.New(advertisedStart)

	race.AdvertisedStartTime = ts

	return &race, nil
}

func getRaceStatus(startTime *time.Time, requestTime *time.Time) string {
	if startTime.Before(*requestTime) {
		//advertised start time is in the past
		return "CLOSED"
	} else {
		//advertised start time is equal to current time or in the future
		return "OPEN"
	}
}
