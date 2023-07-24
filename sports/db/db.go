package db

import (
	"time"

	"syreclabs.com/go/faker"
)

var sport_names = [6]string{
	"Rugby league",
	"Aussie rules",
	"Baseball",
	"Basketball",
	"Hockey",
	"Cricket",
}

const number_of_teams int = 20
const number_of_locations int = 20

// Create all required tables and fill with dummy seed data
func (r *sportsRepo) seed() error {
	err := r.createEventTable()
	if err != nil {
		return err
	}
	err = r.createTeamTable()
	if err != nil {
		return err
	}

	err = r.createSportTable()
	if err != nil {
		return err
	}

	err = r.createLocationTable()

	return err
}

func (r *sportsRepo) createEventTable() error {
	err := executeStatement(r,
		`CREATE TABLE IF NOT EXISTS event (
			id INTEGER PRIMARY KEY,
			team_home_id INTEGER,
			team_away_id INTEGER,
			sport_id INTEGER,
			location_id INTEGER,
			advertised_start_time DATETIME,
			duration INTEGER,
			FOREIGN KEY(team_home_id) REFERENCES team(id),
			FOREIGN KEY(team_away_id) REFERENCES team(id),
			FOREIGN KEY(sport_id) REFERENCES sport(id),
			FOREIGN KEY(location_id) REFERENCES location(id)
		)`,
	)
	if err != nil {
		return err
	}

	for i := 1; i <= 100; i++ {
		err = executeStatement(r,
			`INSERT OR IGNORE INTO event (
				id,
				team_home_id,
				team_away_id,
				sport_id,
				location_id,
				advertised_start_time,
				duration
			) VALUES (?,?,?,?,?,?,?)`,
			i,
			faker.Number().Between(1, number_of_teams/2),
			faker.Number().Between((number_of_teams/2)+1, number_of_teams),
			faker.Number().Between(0, len(sport_names)-1),
			faker.Number().Between(1, number_of_locations),
			faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			faker.Numerify("#0"),
		)
		if err != nil {
			return err
		}
	}
	return err
}

func (r *sportsRepo) createTeamTable() error {
	err := executeStatement(r,
		`CREATE TABLE IF NOT EXISTS team (
			id INTEGER PRIMARY KEY,
			name TEXT,
			rank INTEGER
		)
	`)
	if err != nil {
		return err
	}

	for i := 1; i <= number_of_teams; i++ {
		err = executeStatement(r,
			`INSERT OR IGNORE INTO team (
				id,
				name,
				rank
			) VALUES (?,?,?)`,
			i,
			faker.Team().Name(),
			((number_of_teams + 1) - i),
		)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *sportsRepo) createSportTable() error {
	err := executeStatement(r,
		`CREATE TABLE IF NOT EXISTS sport (
			id INTEGER PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		return err
	}

	for i := 1; i <= len(sport_names); i++ {
		err = executeStatement(r,
			`INSERT OR IGNORE INTO sport (
				id,
				name
			) VALUES (?,?)`,
			i,
			sport_names[i-1],
		)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *sportsRepo) createLocationTable() error {
	err := executeStatement(r,
		`CREATE TABLE IF NOT EXISTS location (
			id INTEGER PRIMARY KEY,
			city TEXT,
			capacity INTEGER
		)
	`)
	if err != nil {
		return err
	}

	for i := 1; i <= number_of_locations; i++ {
		err = executeStatement(r,
			`INSERT OR IGNORE INTO location (
				id,
				city,
				capacity
			) VALUES (?,?,?)`,
			i,
			faker.Address().City(),
			faker.Number().Between(2000, 80000),
		)
		if err != nil {
			return err
		}
	}

	return err
}

func executeStatement(r *sportsRepo, statement string, args ...any) error {
	s, err := r.db.Prepare(statement)
	if err == nil {
		_, err = s.Exec(args...)
	}

	return err
}
