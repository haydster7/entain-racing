package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/internal/test_utils"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper harness for running list service procedure
func listTestRun(t *testing.T, mockDb *sql.DB, request *sports.ListEventsRequest) *sports.ListEventsResponse {

	//Create service using mock db and filter
	mockSportsRepo := db.NewSportsRepo(mockDb)
	sportsService := NewSportsService(mockSportsRepo)

	//Call service
	listSportsResponse, err := sportsService.ListEvents(context.TODO(), request)

	//Fail test if errors occurred
	if err != nil {
		t.Error("Error listing events:")
		t.Error(err)
		return nil
	}

	t.Log("list sport events response:")
	t.Log(listSportsResponse)
	return listSportsResponse
}

func compareEvents(r1 *sports.Event, r2 *sports.Event, t *testing.T) bool {
	return r1.Id == r2.Id &&
		r1.AwayTeam == r2.AwayTeam &&
		r1.HomeTeam == r2.HomeTeam &&
		r1.Location == r2.Location &&
		r1.Capacity == r2.Capacity &&
		r1.Sport == r2.Sport &&
		r1.AdvertisedStartTime.Seconds == r2.AdvertisedStartTime.Seconds &&
		r1.ExpectedEndTime.Seconds == r2.ExpectedEndTime.Seconds &&
		r1.Status == r2.Status
}

func sportsResultAssertions(t *testing.T, sampleEvents []*sports.Event, responseEvents []*sports.Event, mock sqlmock.Sqlmock) {
	if len(responseEvents) != len(sampleEvents) {
		t.Errorf("Returned events not expected length. Expected %q, got %q", len(sampleEvents), len(responseEvents))
	}

	areEventsEqual := true
	for i, event := range sampleEvents {
		if !compareEvents(event, responseEvents[i], t) {
			areEventsEqual = false
			t.Logf("Event[%d] does not match", i)
			break
		}
	}
	if !areEventsEqual {
		t.Error("Returned events do not match expected events")
		t.Errorf("Expected: %v", sampleEvents)
		t.Errorf("Got: %v", responseEvents)
	}

	expectationsError := mock.ExpectationsWereMet()
	if expectationsError != nil {
		t.Error("One or more expectations were not met")
		t.Error(expectationsError)
	}
}

func rowValuesFromEvent(t *testing.T, event *sports.Event) (values []driver.Value) {
	startTime := event.AdvertisedStartTime.AsTime()
	endTime := event.ExpectedEndTime.AsTime()
	duration := endTime.Sub(startTime) / time.Minute

	return []driver.Value{
		event.Id,
		event.HomeTeam,
		event.AwayTeam,
		event.Sport,
		event.Location,
		event.Capacity,
		startTime,
		duration,
	}
}

// Tests list procedure with team filter
func TestListEventsWithTeamFilter(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockSportDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results
	teamSearch := "brisbane"

	//Randomly chosed fixed date to use where time is not part of test
	var mockStartTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000)))
	var mockEndTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 50, 57, 0, time.FixedZone("", 36000)))

	//Events to return for expected query/args
	sampleEvents := []*sports.Event{
		{
			Id:                  1,
			HomeTeam:            "Brisbane Broncos",
			AwayTeam:            "Gold Coast Titans",
			Sport:               "Rugby league",
			Location:            "Brisbane",
			Capacity:            30000,
			AdvertisedStartTime: &mockStartTimestamp,
			ExpectedEndTime:     &mockEndTimestamp,
			Status:              "CLOSED",
		},
		{
			Id:                  2,
			HomeTeam:            "Sydney Swans",
			AwayTeam:            "Brisbane Cowboys",
			Sport:               "Rugby league",
			Location:            "Sydney",
			Capacity:            40000,
			AdvertisedStartTime: &mockStartTimestamp,
			ExpectedEndTime:     &mockEndTimestamp,
			Status:              "CLOSED",
		},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.JoinColumnNames)
	for _, event := range sampleEvents {
		includedRows.AddRow(rowValuesFromEvent(t, event)...)
	}

	mockDb.Mock.
		ExpectQuery(`
			SELECT
				event.id,
				team_home.name as home_team,
				team_away.name as away_team,
				sport.name as sport,
				location.city as location,
				location.capacity,
				event.advertised_start_time,
				event.duration
			FROM event
			INNER JOIN team team_home ON team_home.id = event.team_home_id
			INNER JOIN team team_away ON team_away.id = event.team_away_id
			INNER JOIN sport ON sport.id = event.sport_id
			INNER JOIN location ON location.id = event.location_id
			WHERE (home_team LIKE ? OR away_team LIKE ?)`).
		WithArgs("%"+teamSearch+"%", "%"+teamSearch+"%").
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	listEventsRequest := sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{
			Team: teamSearch,
		},
	}

	listResponse := listTestRun(t, mockDb.DB, &listEventsRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	sportsResultAssertions(t, sampleEvents, listResponse.Events, mockDb.Mock)
}

// Tests list procedure with status filter
func TestListEventsWithStatusFilter(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockSportDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var currentTime = time.Now()
	var mockStartTimestamp timestamppb.Timestamp = *timestamppb.New(currentTime.Add(10 * time.Minute))
	var mockEndTimestamp timestamppb.Timestamp = *timestamppb.New(currentTime.Add(40 * time.Minute))

	//Events to return for expected query/args
	sampleEvents := []*sports.Event{
		{
			Id:                  1,
			HomeTeam:            "Brisbane Broncos",
			AwayTeam:            "Gold Coast Titans",
			Sport:               "Rugby league",
			Location:            "Brisbane",
			Capacity:            30000,
			AdvertisedStartTime: &mockStartTimestamp,
			ExpectedEndTime:     &mockEndTimestamp,
			Status:              "OPEN",
		},
		{
			Id:                  2,
			HomeTeam:            "Sydney Swans",
			AwayTeam:            "Brisbane Cowboys",
			Sport:               "Rugby league",
			Location:            "Sydney",
			Capacity:            40000,
			AdvertisedStartTime: &mockStartTimestamp,
			ExpectedEndTime:     &mockEndTimestamp,
			Status:              "OPEN",
		},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.JoinColumnNames)
	for _, event := range sampleEvents {
		includedRows.AddRow(rowValuesFromEvent(t, event)...)
	}

	mockDb.Mock.
		ExpectQuery(`
			SELECT
				event.id,
				team_home.name as home_team,
				team_away.name as away_team,
				sport.name as sport,
				location.city as location,
				location.capacity,
				event.advertised_start_time,
				event.duration
			FROM event
			INNER JOIN team team_home ON team_home.id = event.team_home_id
			INNER JOIN team team_away ON team_away.id = event.team_away_id
			INNER JOIN sport ON sport.id = event.sport_id
			INNER JOIN location ON location.id = event.location_id`).
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	listEventsRequest := sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{
			Status: "OPEN",
		},
	}

	listResponse := listTestRun(t, mockDb.DB, &listEventsRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	sportsResultAssertions(t, sampleEvents, listResponse.Events, mockDb.Mock)
}

// Tests list procedure with status calculation based on time
func TestListEventsStatusCalculation(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockSportDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var currentTime = time.Now()

	//Events to return for expected query/args
	sampleEvents := []*sports.Event{
		{
			Id:                  1,
			HomeTeam:            "Brisbane Broncos",
			AwayTeam:            "Gold Coast Titans",
			Sport:               "Rugby league",
			Location:            "Brisbane",
			Capacity:            30000,
			AdvertisedStartTime: timestamppb.New(currentTime.Add(-10 * time.Minute)),
			ExpectedEndTime:     timestamppb.New(currentTime.Add(10 * time.Minute)),
			Status:              "INPROGRESS",
		},
		{
			Id:                  2,
			HomeTeam:            "Sydney Swans",
			AwayTeam:            "Brisbane Cowboys",
			Sport:               "Rugby league",
			Location:            "Sydney",
			Capacity:            40000,
			AdvertisedStartTime: timestamppb.New(currentTime.Add(-20 * time.Minute)),
			ExpectedEndTime:     timestamppb.New(currentTime.Add(-10 * time.Minute)),
			Status:              "CLOSED",
		},
		{
			Id:                  3,
			HomeTeam:            "Sydney Swans",
			AwayTeam:            "Canberra Raiders",
			Sport:               "Rugby league",
			Location:            "Sydney",
			Capacity:            40000,
			AdvertisedStartTime: timestamppb.New(currentTime.Add(10 * time.Minute)),
			ExpectedEndTime:     timestamppb.New(currentTime.Add(20 * time.Minute)),
			Status:              "OPEN",
		},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.JoinColumnNames)
	for _, event := range sampleEvents {
		includedRows.AddRow(rowValuesFromEvent(t, event)...)
	}

	mockDb.Mock.
		ExpectQuery(`
			SELECT
				event.id,
				team_home.name as home_team,
				team_away.name as away_team,
				sport.name as sport,
				location.city as location,
				location.capacity,
				event.advertised_start_time,
				event.duration
			FROM event
			INNER JOIN team team_home ON team_home.id = event.team_home_id
			INNER JOIN team team_away ON team_away.id = event.team_away_id
			INNER JOIN sport ON sport.id = event.sport_id
			INNER JOIN location ON location.id = event.location_id`).
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	listEventsRequest := sports.ListEventsRequest{}

	listResponse := listTestRun(t, mockDb.DB, &listEventsRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	sportsResultAssertions(t, sampleEvents, listResponse.Events, mockDb.Mock)
}

// Test getting a single event by id
func TestGetRace(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockSportDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var mockStartTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000)))
	var mockEndTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 50, 57, 0, time.FixedZone("", 36000)))

	//Events to return for expected query/args
	sampleEvents := []*sports.Event{
		{
			Id:                  2,
			HomeTeam:            "Sydney Swans",
			AwayTeam:            "Brisbane Cowboys",
			Sport:               "Rugby league",
			Location:            "Sydney",
			Capacity:            40000,
			AdvertisedStartTime: &mockStartTimestamp,
			ExpectedEndTime:     &mockEndTimestamp,
			Status:              "CLOSED",
		},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.JoinColumnNames)
	for _, event := range sampleEvents {
		includedRows.AddRow(rowValuesFromEvent(t, event)...)
	}

	mockDb.Mock.
		ExpectQuery(`
			SELECT
				event.id,
				team_home.name as home_team,
				team_away.name as away_team,
				sport.name as sport,
				location.city as location,
				location.capacity,
				event.advertised_start_time,
				event.duration
			FROM event
			INNER JOIN team team_home ON team_home.id = event.team_home_id
			INNER JOIN team team_away ON team_away.id = event.team_away_id
			INNER JOIN sport ON sport.id = event.sport_id
			INNER JOIN location ON location.id = event.location_id
			WHERE event.id = ?`).
		WithArgs(2).
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	getEventRequest := sports.GetEventRequest{
		Id: 2,
	}

	//Create service using mock db and filter
	mockSportsRepo := db.NewSportsRepo(mockDb.DB)
	sportsService := NewSportsService(mockSportsRepo)

	//Call service
	getEventResponse, err := sportsService.GetEvent(context.TODO(), &getEventRequest)

	//Fail test if errors occurred
	if err != nil {
		t.Error("Error listing event(s):")
		t.Error(err)
	}

	t.Log("get event response:")
	t.Log(getEventResponse)

	//Cleanup mock database
	mockDbHelper.Close()

	sportsResultAssertions(t, sampleEvents, []*sports.Event{getEventResponse.Event}, mockDb.Mock)
}
