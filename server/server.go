package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sptodo/auth"
	"sptodo/todo"
	"strconv"
)

type AddTodoRequest struct {
	Title string `json:"title"`
}

var authSystem *auth.Auth // ← глобальная переменная

func Start() error {
	var err error
	authSystem, err = auth.NewAuth()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	// Публичные эндпоинты
	mux.HandleFunc("POST /api/register", handleRegister)
	mux.HandleFunc("POST /api/login", handleLogin)
	mux.HandleFunc("POST /api/account", requireAuth(handleDeleteAccount))

	// Защищённые эндпоинты
	mux.HandleFunc("GET /api/todos", requireAuth(getTodos))
	mux.HandleFunc("POST /api/todos", requireAuth(addTodo))
	mux.HandleFunc("PUT /api/todos/{id}/complete", requireAuth(completeTodo))
	mux.HandleFunc("DELETE /api/todos/{id}", requireAuth(deleteTodo))

	// Статика
	mux.Handle("/", http.FileServer(http.Dir("./web/")))

	return http.ListenAndServe(":8080", mux)
}

func handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value("user").(string)

	if err := authSystem.DeleteUser(login); err != nil {
		http.Error(w, "Ошибка удаления аккаунта: "+err.Error(), http.StatusInternalServerError)
		return
	}

	authSystem.ClearUserSessions(login)

	w.WriteHeader(http.StatusNoContent)
}

// Тяжелая и пока что не понятная для меня функция в плане написания кода
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Требуется вход", http.StatusUnauthorized)
			return // ⛔ прерываем выполнение — next НЕ вызывается
		}

		login, err := authSystem.GetLogin(cookie.Value)
		if err != nil {
			http.Error(w, "Сессия недействительна", http.StatusUnauthorized)
			return // ⛔ снова прерываем
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", login))

		// 4. Вызываем оригинальный обработчик
		next(w, r)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if err := authSystem.Register(req.Login, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	sessionID, err := authSystem.Login(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(auth.SessionTTl.Seconds()),
	})

	w.WriteHeader(http.StatusOK)
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value("user").(string) // ← получили логин

	var todos todo.Todos
	if err := todos.Load(login); err != nil { // ← передали логин
		http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
	}
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value("user").(string)

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
	if err := todos.Load(login); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}
	todos.Add(req.Title) // игнорируем ошибку "файл не найден"

	if err := todos.Save(login); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value("user").(string)

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
	if err := todos.Load(login); err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "Ошибка загрузки", http.StatusInternalServerError)
			return
		}
	}

	if err := todos.Delete(id); err != nil {
		http.Error(w, "Ошибка удаления: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := todos.Save(login); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	// 6. Отвечаем 204 No Content (успешно, без тела)
	w.WriteHeader(http.StatusNoContent)
}

func completeTodo(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value("user").(string)

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
	if err := todos.Load(login); err != nil {
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
	if err := todos.Save(login); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	// 6. Отвечаем 204
	w.WriteHeader(http.StatusNoContent)

}
