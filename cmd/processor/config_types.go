package main

type config struct {
	kafkaBrokers   []string
	kafkaTopic     string
	dlqTopic       string
	consumerGroup  string
	kafkaUsername  string
	kafkaPassword  string
	workerCount    int
	postgresHost   string
	postgresPort   string
	postgresDB     string
	postgresUser   string
	postgresPass   string
	clickHouseHost string
	clickHousePort string
	clickHouseDB   string
	clickHouseUser string
	clickHousePass string
}
