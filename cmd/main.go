package main

import "tradebot/pkg/app"

func main() {
	application := app.New(".env")
	application.Start()
}
