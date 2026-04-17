package initialize

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

// postgress
// clickhouse
// redis

type Clients struct {
	PgClient *sql.DB
	Cache    *redis.Client
}

func GetClients() *Clients {
	return &Clients{
		PgClient: GetPostgressClient(),
		Cache:    GetRedisClient(),
	}
}
