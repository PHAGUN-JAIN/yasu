package initialize

import "database/sql"

// postgress
// clickhouse
// redis

type Clients struct {
	PgClient *sql.DB
}

func GetClients() *Clients {
	return &Clients{
		PgClient: GetPostgressClient(),
	}
}
