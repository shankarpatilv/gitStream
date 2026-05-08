package storage

import "time"

// eventHour defines the aggregation bucket width for events_hourly rows.
const eventHour = time.Hour

// events_hourly receives one count row per event and lets ClickHouse sum later.
const insertEventsHourlySQL = `
INSERT INTO events_hourly (
	hour,
	repo_name,
	event_type,
	count
)`

// events_timeseries receives one lightweight event-time row per event.
const insertEventsTimeseriesSQL = `
INSERT INTO events_timeseries (
	timestamp,
	event_type,
	repo_name
)`
