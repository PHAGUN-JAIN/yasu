package migrations

import "embed"

//go:embed postgres/*.sql
var Postgres embed.FS

//go:embed clickhouse/*.sql
var ClickHouse embed.FS
