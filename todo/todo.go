package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// Моя одна задача
type Todo struct {
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"` //Берем не указатель, потому что нам не нужно менять значение, мы просто в моменте его скопировали и присвоили, все
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Срез для создания списка задач
type Todos []Todo // По факту мы создаем этакий массив Todos в котором храним элементы типа Todo

func (todos *Todos) Add(title string) {
	newTask := Todo{
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	*todos = append(*todos, newTask)
}

func (todos Todos) List(completed_filter *bool) { // Берем не указатель, потому что нам не нужно менять значение, мы просто возвращаем копию
	filter_todo := todos.Filter(completed_filter)
	for index, task := range filter_todo {
		status := "❌"
		completedAt := ""

		if task.Completed {
			status = "✅"
			if task.CompletedAt != nil {
				completedAt = task.CompletedAt.Format("2006-01-02 15:04")
			}
		}

		createdAt := task.CreatedAt.Format("2006-01-02 15:04")
		fmt.Printf("%d %s %s %s %s\n", index+1, task.Title, status, createdAt, completedAt)
	}
}

func (todos *Todos) Complete(index int) error {
	if index < 0 || index >= len(*todos) {
		return errors.New("not index")
	}

	(*todos)[index].Completed = true

	now := time.Now()
	(*todos)[index].CompletedAt = &now
	// &now потому что CompletedAt это указатель на *time.Time
	return nil
}

func (todos *Todos) Delete(index int) error {
	if index < 0 || index >= len(*todos) {
		return errors.New("not index")
	}

	*todos = append((*todos)[:index], (*todos)[index+1:]...)

	return nil
}

// В web версии на данный момент не используется, но List работает через этот метод
func (todos Todos) Filter(complete_filter *bool) Todos {
	var result Todos

	for _, task := range todos {
		if complete_filter == nil {
			result = append(result, task)
		} else if task.Completed == *complete_filter {
			result = append(result, task)
		}
	}

	return result
}

func (todos Todos) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(todos)
}

func (todos *Todos) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	defer file.Close()
	var loaded Todos
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&loaded); err != nil {
		return err
	}

	*todos = loaded
	return nil
}
