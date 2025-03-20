package main

import "WildberriesGo_bot/pkg/app"

func main() {
	application := app.NewApplication("variables.env")
	application.Start()
}
