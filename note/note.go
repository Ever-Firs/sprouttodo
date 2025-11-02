package note

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const dataDir = "data"

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"update_at"`
}

type Notes []Note

func (notes Notes) nextID() int {
	maxID := 0
	for _, n := range notes {
		if n.ID > maxID {
			maxID = n.ID
		}
	}
	return maxID + 1
}

func (notes *Notes) Add(title, content string) {
	now := time.Now()
	newNote := Note{
		ID:        notes.nextID(),
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	*notes = append(*notes, newNote)
}

func (notes *Notes) Update(id int, title, content string) error {
	for i := range *notes {
		if (*notes)[i].ID == id {
			(*notes)[i].Title = title
			(*notes)[i].Content = content
			(*notes)[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return errors.New("Not note")
}

func (notes *Notes) Delete(id int) error {
	for i, n := range *notes {
		if n.ID == id {
			*notes = append((*notes)[:i], (*notes)[i+1:]...)
			return nil
		}
	}

	return errors.New("Not note")
}

func (notes Notes) Save(login string) error {
	dir := filepath.Join(dataDir, login)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "notes.json")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(notes)
}

func (notes *Notes) Load(login string) error {
	path := filepath.Join(dataDir, login, "notes.json")
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			*notes = Notes{}
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(notes)
}
