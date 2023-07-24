package db

const (
	eventsList = "list"
)

func getSportQueries() map[string]string {
	return map[string]string{
		eventsList: `
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
			INNER JOIN team team_home
				ON team_home.id = event.team_home_id
			INNER JOIN team team_away
				ON team_away.id = event.team_away_id
			INNER JOIN sport
				ON sport.id = event.sport_id
			INNER JOIN location
				ON location.id = event.location_id
		`,
	}
}
