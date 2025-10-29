package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	var todos Todos
	_ = todos.Load("todos.json")

	if os.Args[1] != "list" {
		if len(os.Args) != 3 {
			os.Exit(1)
		}
	}
	switch os.Args[1] {
	case "add":
		title := os.Args[2]
		todos.Add(title)
		todos.Save("todos.json")
	case "list":
		var filter *bool

		if len(os.Args) == 3 {
			switch os.Args[2] {
			case "--active":
				active := false
				filter = &active
			case "--completed":
				completed := true
				filter = &completed
			default:
				fmt.Println("Неизвестный флаг. Используй --active или --completed")
				os.Exit(1)
			}
		}
		todos.List(filter)
	case "complete":
		index, err := strconv.Atoi(os.Args[2])
		if err != nil {
			os.Exit(1)
		}
		todos.Complete(index - 1)
		todos.Save("todos.json")
	case "delete":
		index, err := strconv.Atoi(os.Args[2])
		if err != nil {
			os.Exit(1)
		}
		todos.Delete(index - 1)
		todos.Save("todos.json")

	}

}
