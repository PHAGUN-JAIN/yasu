package initialize

import (
	"database/sql"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/redis/go-redis/v9"
)

// postgress
// clickhouse
// redis

type Clients struct {
	PgClient *sql.DB
	Cache    *redis.Client
	ChClient *driver.Conn
}

func GetClients() *Clients {
	return &Clients{
		PgClient: GetPostgressClient(),
		Cache:    GetRedisClient(),
		ChClient: GetChClient(),
	}
}
