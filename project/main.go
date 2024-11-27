package main

import (
	"tickets/service"
)

func main() {

	// TODO: Use wire to initialize all this: https://github.com/google/wire/blob/main/_tutorial/README.md
	svc := service.DefaultFromEnv()

	err := svc.Run()
	if err != nil {
		panic(err)
	}
}
