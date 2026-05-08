package main

import "time"

const (
	service                 = "processor"
	jobCapacity             = 100
	clickHouseBatchSize     = 100
	clickHouseFlushInterval = 5 * time.Second
	workerDrainTimeout      = 30 * time.Second
)
