package main

import (
	"api"
	"collector"
	"smtp_listener"
)

func main() {

	// GARBAGE COLLECTOR
	go collector.Collect()

	// Listen emails
	go smtp_listener.Listen()

	// API HANDLER
	router := api.Handle()
	router.Run(":8080")
}
