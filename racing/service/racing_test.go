package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/internal/test_utils"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper harness for running list service procedure
func listTestRun(t *testing.T, mockDb *sql.DB, request *racing.ListRacesRequest) *racing.ListRacesResponse {

	//Create service using mock db and filter
	mockRacesRepo := db.NewRacesRepo(mockDb)
	racingService := NewRacingService(mockRacesRepo)

	//Call service
	listRacesResponse, err := racingService.ListRaces(context.TODO(), request)

	//Fail test if errors occurred
	if err != nil {
		t.Error("Error listing races:")
		t.Error(err)
		return nil
	}

	t.Log("list races response:")
	t.Log(listRacesResponse)
	return listRacesResponse
}

func compareRaces(r1 *racing.Race, r2 *racing.Race, t *testing.T) bool {
	return r1.Id == r2.Id &&
		r1.MeetingId == r2.MeetingId &&
		r1.Name == r2.Name &&
		r1.Number == r2.Number &&
		r1.Visible == r2.Visible &&
		r1.AdvertisedStartTime.Seconds == r2.AdvertisedStartTime.Seconds
}

func listAssertions(t *testing.T, sampleRaces []*racing.Race, response *racing.ListRacesResponse, mock sqlmock.Sqlmock) {
	responseRaces := response.Races

	if len(responseRaces) != len(sampleRaces) {
		t.Errorf("Returned races not expected length. Expected %q, got %q", len(sampleRaces), len(responseRaces))
	}

	areRacesEqual := true
	for i, race := range sampleRaces {
		if !compareRaces(race, responseRaces[i], t) {
			areRacesEqual = false
			t.Logf("Race[%d] does not match", i)
			break
		}
	}
	if !areRacesEqual {
		t.Error("Returned races do not match expected races")
		t.Errorf("Expected: %v", sampleRaces)
		t.Errorf("Got: %v", sampleRaces)
	}

	expectationsError := mock.ExpectationsWereMet()
	if expectationsError != nil {
		t.Error("One or more expectations were not met")
		t.Error(expectationsError)
	}
}

// Tests list procedure with meeting id filter
func TestListRacesWithMeetingFilter(t *testing.T) {
	//Initiliase mock database
	mockDbHelper := test_utils.NewMockRaceDb(t)
	mockDb := mockDbHelper.Init()

	//Configure mock database for test data and expected results

	//Randomly chosed fixed date to use where time is not part of test
	var mockTimestamp timestamppb.Timestamp = *timestamppb.New(time.Date(2021, time.March, 3, 11, 30, 57, 0, time.FixedZone("", 36000)))

	//Add sample data for test in the format
	//{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}
	meetingIds := []int64{1, 9}

	//Races to return for expected query/args
	sampleRaces := []*racing.Race{
		{Id: 1, MeetingId: meetingIds[0], Name: "Mock race 1", Number: 2, Visible: false, AdvertisedStartTime: &mockTimestamp},
		{Id: 2, MeetingId: meetingIds[1], Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: &mockTimestamp},
	}

	includedRows := mockDb.Mock.NewRows(mockDb.ColumnNames)
	for _, race := range sampleRaces {
		includedRows.AddRow(race.Id, race.MeetingId, race.Name, race.Number, race.Visible, race.AdvertisedStartTime.AsTime())
	}

	mockDb.Mock.
		ExpectQuery(`SELECT id, meeting_id, name, number, visible, advertised_start_time FROM races WHERE meeting_id IN (?,?)`).
		WithArgs(meetingIds[0], meetingIds[1]).
		WillReturnRows(includedRows)

	//Create mock request and filter as input
	listRacesRequest := racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			MeetingIds: meetingIds,
		},
	}

	listResponse := listTestRun(t, mockDb.DB, &listRacesRequest)

	//Cleanup mock database
	mockDbHelper.Close()

	listAssertions(t, sampleRaces, listResponse, mockDb.Mock)
}

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
		{Id: 1, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: &mockTimestamp},
		{Id: 2, MeetingId: 9, Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: &mockTimestamp},
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

	listAssertions(t, sampleRaces, listResponse, mockDb.Mock)
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
		{Id: 1, MeetingId: 1, Name: "Mock race 1", Number: 2, Visible: true, AdvertisedStartTime: timestamppb.New(mockTime.Add(time.Second * 2))},
		{Id: 2, MeetingId: 9, Name: "Mock race 3", Number: 5, Visible: true, AdvertisedStartTime: timestamppb.New(mockTime.Add(time.Second * 1))},
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

	listAssertions(t, sampleRaces, listResponse, mockDb.Mock)
}
