// web/script.js
class TaskFlowApp {
    constructor() {
        this.currentUser = null;
        this.tasks = [];
        this.notes = [];
        this.currentEditingNote = null;
        this.init();
    }

    init() {
        this.checkAuth();
        this.bindEvents();
    }

    async checkAuth() {
        try {
            const response = await fetch('/api/todos');
            if (response.ok) {
                this.currentUser = true;
                this.showMainScreen();
                this.loadTodos();
                this.loadNotes();
            } else {
                this.showAuthScreen();
            }
        } catch (error) {
            this.showAuthScreen();
        }
    }

    bindEvents() {
        // Навигация
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('nav-item')) {
                document.querySelectorAll('.nav-item').forEach(item => {
                    item.classList.remove('active');
                });
                e.target.classList.add('active');
            }
        });
    }

    showAuthScreen() {
        document.getElementById('authScreen').classList.add('active');
        document.getElementById('mainScreen').classList.remove('active');
    }

    showMainScreen() {
        document.getElementById('authScreen').classList.remove('active');
        document.getElementById('mainScreen').classList.add('active');
    }

    showSection(sectionName) {
        document.querySelectorAll('.content-section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(sectionName + 'Section').classList.add('active');
    }

    showModal(modalId) {
        document.getElementById(modalId).classList.add('active');
    }

    hideModals() {
        document.querySelectorAll('.modal').forEach(modal => {
            modal.classList.remove('active');
        });
        this.currentEditingNote = null;
    }

    // Авторизация
    async handleLogin() {
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;

        if (!username || !password) {
            alert('Please fill in all fields');
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ login: username, password })
            });

            if (response.ok) {
                this.currentUser = username;
                document.getElementById('usernameDisplay').textContent = username;
                this.showMainScreen();
                this.loadTodos();
                this.loadNotes();
            } else {
                alert('Login failed');
            }
        } catch (error) {
            alert('Network error');
        }
    }

    async handleRegister() {
        const username = document.getElementById('registerUsername').value;
        const password = document.getElementById('registerPassword').value;

        if (!username || !password) {
            alert('Please fill in all fields');
            return;
        }

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ login: username, password })
            });

            if (response.ok) {
                alert('Account created! Please sign in.');
                showLoginForm();
            } else {
                alert('Registration failed');
            }
        } catch (error) {
            alert('Network error');
        }
    }

    async handleLogout() {
        try {
            // Очищаем куки сессии
            document.cookie = 'session_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
            this.currentUser = null;
            this.showAuthScreen();
        } catch (error) {
            console.error('Logout error:', error);
        }
    }

    // Задачи
    async loadTodos() {
        try {
            const response = await fetch('/api/todos');
            if (response.ok) {
                this.tasks = await response.json();
                this.renderTodos();
            }
        } catch (error) {
            console.error('Error loading todos:', error);
        }
    }

    renderTodos() {
        const todoList = document.getElementById('todoList');
        const completedList = document.getElementById('completedList');
        
        todoList.innerHTML = '';
        completedList.innerHTML = '';

        this.tasks.forEach((task, index) => {
            const taskElement = this.createTaskElement(task, index);
            if (task.completed) {
                completedList.appendChild(taskElement);
            } else {
                todoList.appendChild(taskElement);
            }
        });
    }

    createTaskElement(task, index) {
        const taskDiv = document.createElement('div');
        taskDiv.className = 'task-item';
        taskDiv.innerHTML = `
            <div class="task-checkbox ${task.completed ? 'checked' : ''}" 
                 onclick="app.toggleTask(${index})"></div>
            <div class="task-content">
                <div class="task-title">${task.title}</div>
                <div class="task-date">${new Date(task.created_at).toLocaleDateString()}</div>
            </div>
            <button class="delete-btn" onclick="app.deleteTask(${index})">🗑️</button>
        `;
        return taskDiv;
    }

    async addTask() {
        const title = document.getElementById('taskTitleInput').value;
        if (!title) return;

        try {
            await fetch('/api/todos', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title })
            });

            this.hideModals();
            document.getElementById('taskTitleInput').value = '';
            this.loadTodos();
        } catch (error) {
            alert('Error creating task');
        }
    }

    async toggleTask(index) {
        try {
            await fetch(`/api/todos/${index}/complete`, {
                method: 'PUT'
            });
            this.loadTodos();
        } catch (error) {
            alert('Error updating task');
        }
    }

    async deleteTask(index) {
        try {
            await fetch(`/api/todos/${index}`, {
                method: 'DELETE'
            });
            this.loadTodos();
        } catch (error) {
            alert('Error deleting task');
        }
    }

    // Заметки
    async loadNotes() {
        try {
            const response = await fetch('/api/notes');
            if (response.ok) {
                const notePreviews = await response.json();
                // Загружаем полные данные для каждой заметки
                this.notes = await Promise.all(
                    notePreviews.map(async (preview) => {
                        try {
                            const fullNoteResponse = await fetch(`/api/notes/${preview.id}`);
                            if (fullNoteResponse.ok) {
                                return await fullNoteResponse.json();
                            }
                        } catch (error) {
                            console.error('Error loading full note:', error);
                        }
                        return preview;
                    })
                );
                this.renderNotes();
            }
        } catch (error) {
            console.error('Error loading notes:', error);
        }
    }

    renderNotes() {
        const notesGrid = document.getElementById('notesGrid');
        notesGrid.innerHTML = '';

        this.notes.forEach(note => {
            const noteElement = document.createElement('div');
            noteElement.className = 'note-item';
            
            // Исправляем поле updated_at (в бэкенде оно называется update_at)
            const updatedAt = note.update_at || note.updated_at;
            const createdAt = note.created_at;
            
            noteElement.innerHTML = `
                <div class="note-header">
                    <div class="note-title">${note.title}</div>
                    <button class="delete-btn" onclick="event.stopPropagation(); app.deleteNote(${note.id})">🗑️</button>
                </div>
                <div class="note-content-preview">${note.content || 'No content yet...'}</div>
                <div class="note-date">
                    Updated: ${new Date(updatedAt).toLocaleDateString()}
                </div>
            `;
            
            // Добавляем обработчик клика для просмотра полной заметки
            noteElement.addEventListener('click', () => {
                this.viewNote(note);
            });
            
            notesGrid.appendChild(noteElement);
        });
    }

    viewNote(note) {
        // Исправляем поле updated_at (в бэкенде оно называется update_at)
        const updatedAt = note.update_at || note.updated_at;
        const createdAt = note.created_at;
        
        document.getElementById('viewNoteTitle').textContent = note.title;
        document.getElementById('viewNoteContent').textContent = note.content || 'No content yet...';
        document.getElementById('viewNoteCreatedAt').textContent = new Date(createdAt).toLocaleString();
        document.getElementById('viewNoteUpdatedAt').textContent = new Date(updatedAt).toLocaleString();
        
        // Сохраняем текущую заметку для возможного редактирования
        this.currentEditingNote = note;
        
        this.showModal('viewNoteModal');
    }

    enableNoteEdit() {
        if (!this.currentEditingNote) return;
        
        document.getElementById('editNoteTitle').value = this.currentEditingNote.title;
        document.getElementById('editNoteContent').value = this.currentEditingNote.content || '';
        
        this.hideModals();
        this.showModal('editNoteModal');
    }

    async saveNoteEdit() {
        if (!this.currentEditingNote) return;

        const title = document.getElementById('editNoteTitle').value;
        const content = document.getElementById('editNoteContent').value;

        if (!title) {
            alert('Note title is required');
            return;
        }

        try {
            const response = await fetch(`/api/notes/${this.currentEditingNote.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title, content })
            });

            if (response.ok) {
                this.hideModals();
                this.loadNotes();
            } else {
                alert('Error updating note');
            }
        } catch (error) {
            alert('Network error');
        }
    }

    showViewNoteModal() {
        this.hideModals();
        if (this.currentEditingNote) {
            this.viewNote(this.currentEditingNote);
        }
    }

    async addNote() {
        const title = document.getElementById('noteTitleInput').value;
        const content = document.getElementById('noteContentInput').value;

        if (!title) return;

        try {
            await fetch('/api/notes', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title, content })
            });

            this.hideModals();
            document.getElementById('noteTitleInput').value = '';
            document.getElementById('noteContentInput').value = '';
            this.loadNotes();
        } catch (error) {
            alert('Error creating note');
        }
    }

    async deleteNote(id) {
        if (!confirm('Are you sure you want to delete this note?')) return;
        
        try {
            await fetch(`/api/notes/${id}`, {
                method: 'DELETE'
            });
            this.loadNotes();
        } catch (error) {
            alert('Error deleting note');
        }
    }

    // Управление аккаунтом
    async deleteAccount() {
        if (!confirm('Are you sure? This action cannot be undone!')) return;

        try {
            await fetch('/api/account', {
                method: 'POST'
            });

            this.hideModals();
            this.handleLogout();
        } catch (error) {
            alert('Error deleting account');
        }
    }
}

// Глобальные функции для обработки событий
function showLoginForm() {
    document.getElementById('loginForm').classList.add('active');
    document.getElementById('registerForm').classList.remove('active');
}

function showRegisterForm() {
    document.getElementById('loginForm').classList.remove('active');
    document.getElementById('registerForm').classList.add('active');
}

function showAddTaskModal() {
    document.getElementById('addTaskModal').classList.add('active');
}

function showAddNoteModal() {
    document.getElementById('addNoteModal').classList.add('active');
}

function showDeleteAccountModal() {
    document.getElementById('deleteAccountModal').classList.add('active');
}

function hideModals() {
    app.hideModals();
}

function handleLogin() {
    app.handleLogin();
}

function handleRegister() {
    app.handleRegister();
}

function handleLogout() {
    app.handleLogout();
}

function showSection(section) {
    app.showSection(section);
}

function addTask() {
    app.addTask();
}

function addNote() {
    app.addNote();
}

function deleteAccount() {
    app.deleteAccount();
}

function enableNoteEdit() {
    app.enableNoteEdit();
}

function saveNoteEdit() {
    app.saveNoteEdit();
}

function showViewNoteModal() {
    app.showViewNoteModal();
}

// Инициализация приложения
const app = new TaskFlowApp();

// Назначаем методы app глобально для обработчиков onclick
window.app = app;
window.showLoginForm = showLoginForm;
window.showRegisterForm = showRegisterForm;
window.showAddTaskModal = showAddTaskModal;
window.showAddNoteModal = showAddNoteModal;
window.showDeleteAccountModal = showDeleteAccountModal;
window.hideModals = hideModals;
window.handleLogin = handleLogin;
window.handleRegister = handleRegister;
window.handleLogout = handleLogout;
window.showSection = showSection;
window.addTask = addTask;
window.addNote = addNote;
window.deleteAccount = deleteAccount;
window.enableNoteEdit = enableNoteEdit;
window.saveNoteEdit = saveNoteEdit;
window.showViewNoteModal = showViewNoteModal;