package main

import "time"

const (
	service            = "processor"
	jobCapacity        = 100
	workerDrainTimeout = 30 * time.Second
)
