package main

import (
	"sptodo/server"
)

func main() {
	if err := server.Start(); err != nil {
		panic(err)
	}
}
