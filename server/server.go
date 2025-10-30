package server

import (
	"encoding/json"
	"net/http"
	"os"
	"sptodo/todo"
	"strconv"
	// "sptodo/todo"
)

type AddTodoRequest struct {
	Title string `json:"title"`
}

func Start() error {
	mux := http.NewServeMux()

	//API
	mux.HandleFunc("GET /api/todos", getTodos)
	mux.HandleFunc("POST /api/todos", addTodo)
	mux.HandleFunc("DELETE /api/todos/{id}", deleteTodo)
	mux.HandleFunc("PUT /api/todos/{id}/complete", completeTodo)

	// Статика
	mux.Handle("/", http.FileServer(http.Dir("./web/")))

	return http.ListenAndServe(":8080", mux)
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	var todos todo.Todos
	_ = todos.Load("todos.json")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
	}
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Тело запроса должно быть в формате JSON", http.StatusBadRequest)
		return
	}

	var req AddTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка десериализации", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Заголовок задачи не может быть пустым", http.StatusBadRequest)
		return
	}
	var todos todo.Todos
	_ = todos.Load("todos.json")
	todos.Add(req.Title) // игнорируем ошибку "файл не найден"

	if err := todos.Save("todos.json"); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Not id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный id", http.StatusBadRequest)
		return
	}

	var todos todo.Todos
	if err := todos.Load("todos.json"); err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
			return
		}
	}

	if err := todos.Delete(id); err != nil {
		http.Error(w, "Ошибка удаления: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := todos.Save("todos.json"); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	// 6. Отвечаем 204 No Content (успешно, без тела)
	w.WriteHeader(http.StatusNoContent)
}

func completeTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Not id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неправильный id", http.StatusBadRequest)
		return
	}

	var todos todo.Todos
	if err := todos.Load("todos.json"); err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "Ошибка загрузки задач", http.StatusInternalServerError)
			return
		}
	}

	if err := todos.Complete(id); err != nil {
		http.Error(w, "Ошибка завершения: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 5. Сохраняем
	if err := todos.Save("todos.json"); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	// 6. Отвечаем 204
	w.WriteHeader(http.StatusNoContent)

}
