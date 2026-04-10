package store

// CHStore handles analytics writes and reads
// Use: github.com/ClickHouse/clickhouse-go/v2

type CHStore struct {
	// TODO: hold clickhouse.Conn
}

// TODO: implement BatchInsert(ctx, []ClickEvent) error
// Key learning: use clickhouse.PrepareBatch for efficient bulk inserts
// Read the clickhouse-go docs on batch inserts - it's different from pgx

// TODO: implement GetStats(ctx, shortCode string) (*Stats, error)
// Hint: think about what aggregations you need
// SELECT count(), uniq(ip), any(country) ... GROUP BY toDate(clicked_at)

type Stats struct {
	// TODO: TotalClicks, ByCountry map[string]int, ByDevice map[string]int, ByDay []DayStat
}