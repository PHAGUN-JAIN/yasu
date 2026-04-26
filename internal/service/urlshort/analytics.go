package service

import (
	"context"
	"fmt"
	"time"

	"urlshort/internal/repository"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type AnalyticsService struct {
	chConn driver.Conn
}

// TODO: build a materialised view in ClickHouse that pre-aggregates analytics data by day, country, device, and source to speed up queries. The Get method can then query this materialised view instead of the raw clicks table for faster response times.

func (s *AnalyticsService) Get(ctx context.Context, code string) (*repository.Stats, error) {
	stats := &repository.Stats{
		Code:      code,
		ByCountry: make(map[string]uint64),
		ByDevice:  make(map[string]uint64),
		BySource:  make(map[string]uint64),
	}

	// total clicks
	err := s.chConn.QueryRow(ctx,
		"SELECT count() FROM clicks WHERE short_code = $1", code,
	).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("total clicks: %w", err)
	}

	// by country
	rows, err := s.chConn.Query(ctx,
		`SELECT country, count() as clicks
		 FROM clicks WHERE short_code = $1
		 GROUP BY country ORDER BY clicks DESC`, code,
	)
	if err != nil {
		return nil, fmt.Errorf("by country: %w", err)
	}
	for rows.Next() {
		var country string
		var clicks uint64
		if err := rows.Scan(&country, &clicks); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByCountry[country] = clicks
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	// by device
	rows, err = s.chConn.Query(ctx,
		`SELECT device, count() as clicks
		 FROM clicks WHERE short_code = $1
		 GROUP BY device ORDER BY clicks DESC`, code,
	)
	if err != nil {
		return nil, fmt.Errorf("by device: %w", err)
	}
	for rows.Next() {
		var device string
		var clicks uint64
		if err := rows.Scan(&device, &clicks); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByDevice[device] = clicks
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	// by source
	rows, err = s.chConn.Query(ctx,
		`SELECT source, count() as clicks
		 FROM clicks WHERE short_code = $1
		 GROUP BY source ORDER BY clicks DESC`, code,
	)
	if err != nil {
		return nil, fmt.Errorf("by source: %w", err)
	}
	for rows.Next() {
		var source string
		var clicks uint64
		if err := rows.Scan(&source, &clicks); err != nil {
			rows.Close()
			return nil, err
		}
		stats.BySource[source] = clicks
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	// by day
	rows, err = s.chConn.Query(ctx,
		`SELECT toDate(clicked_at) as date, count() as clicks
		 FROM clicks WHERE short_code = $1
		 GROUP BY date ORDER BY date`, code,
	)
	if err != nil {
		return nil, fmt.Errorf("by day: %w", err)
	}
	for rows.Next() {
		var date time.Time
		var clicks uint64
		if err := rows.Scan(&date, &clicks); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByDay = append(stats.ByDay, repository.DayStat{
			Date:   date,
			Clicks: clicks,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	return stats, nil
}

func NewAnalyticsService(chConn driver.Conn) *AnalyticsService {
	return &AnalyticsService{
		chConn: chConn,
	}
}
