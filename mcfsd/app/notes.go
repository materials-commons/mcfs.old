package app

import (
	"time"

	"github.com/materials-commons/mcfs/common"
)

// Note is a user note entry.
type Note struct {
	ID        string    `json:"id"`
	Birthtime time.Time `json:"birthtime"`
	MTime     time.Time `json:"mtime"`
	Date      int       `json:"date"`
	Message   string    `json:"message"`
	Who       string    `json:"who"`
}

type NotesService interface {
	Insert(t common.Type, id string, note Note) (Note, error)
	Update(t common.Type, id string, note Note) (Note, error)
	Remove(t common.Type, id string, noteID string) (Note, error)
}

type notesService struct {
}
