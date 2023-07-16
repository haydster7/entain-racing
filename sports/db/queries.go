package db

const (
	eventsList = "list"
)

func getSportQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT 
				id, 
				meeting_id, 
				name, 
				number, 
				visible, 
				advertised_start_time 
			FROM sports
		`,
	}
}
