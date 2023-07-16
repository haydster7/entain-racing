package service

import (
	"context"
	"database/sql"
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
		r1.MeetingId == r2.MeetingId &&
		r1.Name == r2.Name &&
		r1.Number == r2.Number &&
		r1.Visible == r2.Visible &&
		r1.AdvertisedStartTime.Seconds == r2.AdvertisedStartTime.Seconds &&
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
			t.Logf("Race[%d] does not match", i)
			break
		}
	}
	if !areEventsEqual {
		t.Error("Returned events do not match expected events")
		t.Errorf("Expected: %v", sampleEvents)
		t.Errorf("Got: %v", sampleEvents)
	}

	expectationsError := mock.ExpectationsWereMet()
	if expectationsError != nil {
		t.Error("One or more expectations were not met")
		t.Error(expectationsError)
	}
}

// Tests list procedure with meeting id filter
func TestListEventsWithMeetingFilter(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockSportDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var mockTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000)))

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}
	meetingIds := []int64{1, 9}

	//Events to return for expected query/args
	sampleEvents := []*sports.Event{
		{Id: 1, MeetingId: meetingIds[0], Name: "Mock sport event 1", Number: 2, Visible: false, AdvertisedStartTime: &mockTimestamp, Status: "CLOSED"},
		{Id: 2, MeetingId: meetingIds[1], Name: "Mock sport event 3", Number: 5, Visible: true, AdvertisedStartTime: &mockTimestamp, Status: "CLOSED"},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, event := range sampleEvents {
		includedRows.AddRow(event.Id, event.MeetingId, event.Name, event.Number, event.Visible, event.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM sports WHERE meeting_id IN (?,?)`).
		WithArgs(meetingIds[0], meetingIds[1]).
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	listEventsRequest := sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{
			MeetingIds: meetingIds,
		},
	}

	listResponse := listTestRun(t, mockDb.DB, &listEventsRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	sportsResultAssertions(t, sampleEvents, listResponse.Events, mockDb.Mock)
}

/**************
// Tests list procedure with visibility filter
func TestListRacesWithVisibilityFilter(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockRaceDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var mockTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000)))

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}

	//Races to return for expected query/args
	sampleRaces := []*racing.Race{
		{Id: 1, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: &mockTimestamp, Status: "CLOSED"},
		{Id: 2, MeetingId: 9, Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: &mockTimestamp, Status: "CLOSED"},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, race := range sampleRaces {
		includedRows.AddRow(race.Id, race.MeetingId, race.Name, race.Number, race.Visible, race.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM races WHERE visible = ?`).
		WithArgs(true).
		WillReturnRows(includedRows)

	//Create mock filter
	visibility := new(bool)
	*visibility = true

	//Create mock request and filter as input
	listRacesRequest := racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			Visible: visibility,
		},
	}

	listResponse := listTestRun(t, mockDb.DB, &listRacesRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	raceResultAssertions(t, sampleRaces, listResponse.Races, mockDb.Mock)
}

// Tests list prodecure with sort order specified
func TestListRacesWithSortOrder(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockRaceDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	mockTime := time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000))

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}

	//Races to return for expected query/args
	sampleRaces := []*racing.Race{
		{Id: 1, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: timestamppb.New(mockTime.Add(time.Second * 2)), Status: "CLOSED"},
		{Id: 2, MeetingId: 9, Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: timestamppb.New(mockTime.Add(time.Second * 1)), Status: "CLOSED"},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, race := range sampleRaces {
		includedRows.AddRow(race.Id, race.MeetingId, race.Name, race.Number, race.Visible, race.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM races ORDER BY advertised_start_time desc`).
		WillReturnRows(includedRows)

	//Create mock request as input
	listRacesRequest := racing.ListRacesRequest{
		OrderBy: "advertised_start_time desc",
	}

	listResponse := listTestRun(t, mockDb.DB, &listRacesRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	raceResultAssertions(t, sampleRaces, listResponse.Races, mockDb.Mock)
}

// Tests list prodecure with varying statuses
func TestListRacesStatuses(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockRaceDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	mockTime := time.Now()
	mockTimePast := mockTime.Add(time.Minute * -1)
	mockTimeFuture := mockTime.Add(time.Minute)

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}

	//Races to return for expected query/args
	sampleRaces := []*racing.Race{
		{Id: 1, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: timestamppb.New(mockTimePast), Status: "CLOSED"},
		{Id: 2, MeetingId: 9, Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: timestamppb.New(mockTimeFuture), Status: "OPEN"},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, race := range sampleRaces {
		includedRows.AddRow(race.Id, race.MeetingId, race.Name, race.Number, race.Visible, race.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM races`).
		WillReturnRows(includedRows)

	//Create mock request as input
	listRacesRequest := racing.ListRacesRequest{}

	listResponse := listTestRun(t, mockDb.DB, &listRacesRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	raceResultAssertions(t, sampleRaces, listResponse.Races, mockDb.Mock)
}

// Test getting a single race by id
func TestGetRace(t *testing.T) {

	//Initiliase mock database
	mockDbHelper := test_utils.NewMockRaceDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	mockTime := time.Now()

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}

	//Races to return for expected query/args
	sampleRaces := []*racing.Race{
		{Id: 2, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: timestamppb.New(mockTime), Status: "CLOSED"},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, race := range sampleRaces {
		includedRows.AddRow(race.Id, race.MeetingId, race.Name, race.Number, race.Visible, race.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM races WHERE id = ?`).
		WithArgs(2).
		WillReturnRows(includedRows)

	//Create mock request as input
	var getRaceRequest racing.GetRaceRequest
	getRaceRequest.Id = 2

	//Create service using mock db and filter
	mockRacesRepo := db.NewRacesRepo(mockDb.DB)
	racingService := NewRacingService(mockRacesRepo)

	//Call service
	getRaceResponse, err := racingService.GetRace(context.TODO(), &getRaceRequest)

	//Fail test if errors occurred
	if err != nil {
		t.Error("Error listing races:")
		t.Error(err)
	}

	t.Log("get race response:")
	t.Log(getRaceResponse)

	//Cleanup mock database
	mockDbHelper.Close()

	raceResultAssertions(t, sampleRaces, []*racing.Race{getRaceResponse.Race}, mockDb.Mock)
}
*****************/
