package main

import "tradebot/pkg/app"

func main() {
	application := app.NewApplication("config/.env")
	application.Start()
}
