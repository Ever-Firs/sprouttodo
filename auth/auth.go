package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	dataDir    = "data"
	usersFile  = "users.json"
	SessionTTl = 24 * time.Hour
)

type User struct {
	Login        string `json:"login"`
	PasswordHash []byte `json:"password_hash"`
}

type Session struct {
	Login  string    `json:"login"`
	Expiry time.Time `json:"expity"`
}

type Auth struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	users    map[string]*User
}

// Создаем систему авторизации
func NewAuth() (*Auth, error) {
	a := &Auth{
		sessions: make(map[string]*Session),
		users:    make(map[string]*User),
	}
	// Загружаем пользователей
	if err := a.loadUsers(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		//Если файла нету - создаем папку
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			return nil, err
		}
	}

	go a.cleanupSessions() //Очистка просроченных сессий

	return a, nil

}

func (a *Auth) generateSessionID() (string, error) {
	bytes := make([]byte, 32)                   //Создаем срез при помощи функции make для создания
	if _, err := rand.Read(bytes); err != nil { //crypto/rand
		return "", err
	}

	return hex.EncodeToString(bytes), nil
	// С помощью hex.EncodeToString (из пакета encoding/hex) байты преобразуются в строку из 64
	// шестнадцатеричных символов (поскольку каждый байт кодируется двумя hex-символами: например, 0xA3 → "a3")
}

func (a *Auth) CreateSession(login string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sessionID, err := a.generateSessionID()
	if err != nil {
		return "", err
	}

	a.sessions[sessionID] = &Session{
		Login:  login,
		Expiry: time.Now().Add(SessionTTl),
	}

	return sessionID, nil
}

func (a *Auth) GetLogin(sessionID string) (string, error) {
	a.mu.Lock() //RLock Не оправдывает себя (delete в функции)
	defer a.mu.Unlock()

	session, ok := a.sessions[sessionID]
	if !ok {
		return "", errors.New("No session")
	}

	if time.Now().After(session.Expiry) {
		delete(a.sessions, sessionID)
		return "", errors.New("Сессия истекла")
	}

	return session.Login, nil
}

func (a *Auth) Register(login, password string) error {
	if login == "" || password == "" {
		return errors.New("Логин и пароль обязательны")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.users[login]; exists {
		return errors.New("Пользователь с таким логином уже существует")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	a.users[login] = &User{
		Login:        login,
		PasswordHash: hash,
	}

	return a.saveUsers()
}

func (a *Auth) Login(login, password string) (string, error) {
	a.mu.Lock()
	user, exists := a.users[login]
	a.mu.Unlock()

	if !exists {
		return "", errors.New("Неверный логин или пароль")
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", errors.New("Неверный логин или пароль")
	}

	return a.CreateSession(login)

}

func (a *Auth) DeleteUser(login string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.users[login]; !exists {
		return errors.New("пользователь не найден")
	}

	userDir := filepath.Join(dataDir, login)
	if err := os.RemoveAll(userDir); err != nil {
		return fmt.Errorf("ошибка удаления данных пользователя: %w", err)
	}

	delete(a.users, login)
	return a.saveUsers()
}

func (a *Auth) loadUsers() error {
	data, err := os.ReadFile(filepath.Join(dataDir, usersFile))
	if err != nil {
		return err
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	for _, u := range users {
		a.users[u.Login] = &u
	}

	return nil
}

func (a *Auth) saveUsers() error {
	var users []*User
	for _, u := range a.users {
		users = append(users, u)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dataDir, usersFile), data, 0600)
}

func (a *Auth) cleanupSessions() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		for id, session := range a.sessions {
			if now.After(session.Expiry) {
				delete(a.sessions, id)
			}
		}
		a.mu.Unlock()
	}
}

func (a *Auth) ClearUserSessions(login string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for id, session := range a.sessions {
		if session.Login == login {
			delete(a.sessions, id)
		}
	}
}
