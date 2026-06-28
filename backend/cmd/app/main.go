package main

import (
	_ "eventlineup/docs"
	"eventlineup/internal/app"
)

// @title        EventLineup API
// @version      1.0
// @description  DJ scheduling and event management API
// @host         localhost:8080
// @BasePath     /
func main() { app.Run() }
