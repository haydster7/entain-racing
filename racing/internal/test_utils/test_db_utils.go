package test_utils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockDbHelper interface {
	Init() *mockRaceDb
	Close()
}

type mockRaceDb struct {
	t           *testing.T
	DB          *sql.DB
	Mock        sqlmock.Sqlmock
	ColumnNames []string
}

func NewMockRaceDb(t *testing.T) MockDbHelper {
	return &mockRaceDb{t: t}
}

func (m *mockRaceDb) Init() *mockRaceDb {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	if err != nil {
		m.t.Errorf("Error creating sqlmock database %q", err)
	}

	m.DB = db
	m.Mock = mock
	m.ColumnNames = []string{"id", "meeting_id", "name", "number", "visible", "advertised_start_time"}

	m.Mock.MatchExpectationsInOrder(false)

	return m
}

func (m *mockRaceDb) Close() {
	m.DB.Close()
}
