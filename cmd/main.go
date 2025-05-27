package main

import "tradebot/pkg/app"

func main() {
	application := app.NewApplication(".env")
	application.Start()
}
