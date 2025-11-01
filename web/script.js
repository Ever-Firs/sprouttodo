class TodoApp {
    constructor() {
        this.baseURL = '';
        this.currentUser = localStorage.getItem('currentUser');
        this.initializeApp();
    }

    initializeApp() {
        this.bindEvents();
        this.checkAuth();
    }

    bindEvents() {
        // Auth forms
        document.getElementById('showRegister').addEventListener('click', (e) => {
            e.preventDefault();
            this.showForm('register');
        });

        document.getElementById('showLogin').addEventListener('click', (e) => {
            e.preventDefault();
            this.showForm('login');
        });

        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        document.getElementById('registerForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleRegister();
        });

        // Settings and logout
        document.getElementById('settingsBtn').addEventListener('click', () => {
            this.showSettingsModal();
        });

        document.getElementById('logoutBtn').addEventListener('click', () => {
            this.handleLogout();
        });

        // Modal controls
        document.querySelector('.close-modal').addEventListener('click', () => {
            this.hideSettingsModal();
        });

        document.getElementById('deleteAccountBtn').addEventListener('click', () => {
            this.showDeleteConfirmation();
        });

        document.getElementById('confirmDeleteBtn').addEventListener('click', () => {
            this.handleDeleteAccount();
        });

        document.getElementById('cancelDeleteBtn').addEventListener('click', () => {
            this.hideDeleteConfirmation();
        });

        // Close modal when clicking outside
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideAllModals();
                }
            });
        });

        // Todo form
        document.getElementById('addTodoForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addTodo();
        });
    }

    showForm(formType) {
        document.getElementById('login-form').classList.remove('active');
        document.getElementById('register-form').classList.remove('active');
        
        if (formType === 'register') {
            document.getElementById('register-form').classList.add('active');
        } else {
            document.getElementById('login-form').classList.add('active');
        }
    }

    async handleLogin() {
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        try {
            this.showLoading();
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ login: username, password: password }),
                credentials: 'include'
            });

            if (response.ok) {
                this.currentUser = username;
                localStorage.setItem('currentUser', this.currentUser);
                this.showTodoSection();
                await this.loadTodos();
            } else {
                const errorText = await response.text();
                alert(`Ошибка входа: ${errorText}`);
            }
        } catch (error) {
            console.error('Login error:', error);
            alert('Ошибка подключения к серверу');
        } finally {
            this.hideLoading();
        }
    }

    async handleRegister() {
        const username = document.getElementById('registerUsername').value;
        const password = document.getElementById('registerPassword').value;

        try {
            this.showLoading();
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ login: username, password: password })
            });

            if (response.status === 201) {
                alert('Аккаунт успешно создан! Теперь войдите.');
                this.showForm('login');
                document.getElementById('registerForm').reset();
            } else {
                const errorText = await response.text();
                alert(`Ошибка регистрации: ${errorText}`);
            }
        } catch (error) {
            console.error('Register error:', error);
            alert('Ошибка подключения к серверу');
        } finally {
            this.hideLoading();
        }
    }

    handleLogout() {
        // Просто очищаем куки и выходим, не удаляя аккаунт
        document.cookie = "session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        this.currentUser = null;
        localStorage.removeItem('currentUser');
        this.showAuthSection();
        this.hideAllModals();
    }

    async handleDeleteAccount() {
        try {
            this.showLoading();
            const response = await fetch('/api/account', {
                method: 'POST',
                credentials: 'include'
            });

            if (response.status === 204) {
                alert('Аккаунт успешно удален.');
                // Очищаем фронтенд
                document.cookie = "session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
                this.currentUser = null;
                localStorage.removeItem('currentUser');
                this.hideAllModals();
                this.showAuthSection();
            } else {
                const errorText = await response.text();
                alert(`Ошибка при удалении аккаунта: ${errorText}`);
            }
        } catch (error) {
            console.error('Error deleting account:', error);
            alert('Ошибка подключения к серверу');
        } finally {
            this.hideLoading();
        }
    }

    showSettingsModal() {
        document.getElementById('settingsModal').classList.add('active');
    }

    hideSettingsModal() {
        document.getElementById('settingsModal').classList.remove('active');
    }

    showDeleteConfirmation() {
        this.hideSettingsModal();
        document.getElementById('deleteConfirmModal').classList.add('active');
    }

    hideDeleteConfirmation() {
        document.getElementById('deleteConfirmModal').classList.remove('active');
    }

    hideAllModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.classList.remove('active');
        });
    }

    async checkAuth() {
        if (this.currentUser) {
            try {
                const response = await fetch('/api/todos', {
                    credentials: 'include'
                });
                if (response.ok) {
                    this.showTodoSection();
                    await this.loadTodos();
                    return;
                }
            } catch (error) {
                console.error('Auth check failed:', error);
            }
        }
        this.showAuthSection();
    }

    showAuthSection() {
        document.getElementById('auth-section').classList.add('active');
        document.getElementById('todo-section').classList.remove('active');
    }

    showTodoSection() {
        document.getElementById('auth-section').classList.remove('active');
        document.getElementById('todo-section').classList.add('active');
        document.getElementById('usernameDisplay').textContent = this.currentUser;
    }

    async loadTodos() {
        try {
            const response = await fetch('/api/todos', {
                credentials: 'include'
            });
            
            if (response.ok) {
                const todos = await response.json();
                this.renderTodos(todos);
            } else if (response.status === 401) {
                this.handleLogout();
            } else {
                console.error('Failed to load todos, status:', response.status);
            }
        } catch (error) {
            console.error('Error loading todos:', error);
        }
    }

    async addTodo() {
        const input = document.getElementById('newTodoInput');
        const title = input.value.trim();

        if (!title) return;

        try {
            const response = await fetch('/api/todos', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ title: title }),
                credentials: 'include'
            });

            if (response.status === 201) {
                input.value = '';
                await this.loadTodos();
            } else {
                const errorText = await response.text();
                alert(`Ошибка при добавлении задачи: ${errorText}`);
            }
        } catch (error) {
            console.error('Error adding todo:', error);
            alert('Ошибка подключения к серверу');
        }
    }

    async toggleTodoComplete(todoId) {
        try {
            const response = await fetch(`/api/todos/${todoId}/complete`, {
                method: 'PUT',
                credentials: 'include'
            });
            
            if (response.status === 204) {
                await this.loadTodos();
            } else {
                const errorText = await response.text();
                console.error('Failed to toggle todo:', errorText);
            }
        } catch (error) {
            console.error('Error toggling todo:', error);
        }
    }

    async deleteTodo(todoId) {
        if (!confirm('Удалить эту задачу?')) return;

        try {
            const response = await fetch(`/api/todos/${todoId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.status === 204) {
                await this.loadTodos();
            } else {
                const errorText = await response.text();
                alert(`Ошибка при удалении задачи: ${errorText}`);
            }
        } catch (error) {
            console.error('Error deleting todo:', error);
            alert('Ошибка подключения к серверу');
        }
    }

    renderTodos(todos) {
        const container = document.getElementById('todosList');
        
        if (!todos || todos.length === 0) {
            container.innerHTML = '<div class="empty-state">Задач пока нет</div>';
            return;
        }

        container.innerHTML = todos.map((todo, index) => `
            <div class="todo-item ${todo.completed ? 'completed' : ''}" data-id="${index}">
                <div class="todo-checkbox ${todo.completed ? 'checked' : ''}" 
                     onclick="app.toggleTodoComplete(${index})">
                </div>
                <div class="todo-content">
                    <div class="todo-text">${this.escapeHtml(todo.title)}</div>
                    <div class="todo-dates">
                        <small>Создано: ${this.formatDate(todo.created_at)}</small>
                        ${todo.completed_at ? `<small>Завершено: ${this.formatDate(todo.completed_at)}</small>` : ''}
                    </div>
                </div>
                <button class="todo-delete" onclick="app.deleteTodo(${index})">×</button>
            </div>
        `).join('');
    }

    formatDate(dateString) {
        try {
            const date = new Date(dateString);
            return date.toLocaleDateString('ru-RU') + ' ' + date.toLocaleTimeString('ru-RU', {
                hour: '2-digit',
                minute: '2-digit'
            });
        } catch (e) {
            return dateString;
        }
    }

    escapeHtml(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    showLoading() {
        document.getElementById('loadingOverlay').classList.add('active');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('active');
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new TodoApp();
});