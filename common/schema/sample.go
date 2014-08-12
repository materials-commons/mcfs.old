package schema

import (
	"time"

	"github.com/materials-commons/mcfs/base/schema"
)

type Property struct {
	Name  string      `gorethink:"name"`
	Type  string      `gorethink:"type"`
	Unit  string      `gorethink:"unit"`
	Value interface{} `gorethink:"value"`
}

type Sample struct {
	ID          string              `gorethink:"id,omitempty"`
	Name        string              `gorethink:"name"`
	Owner       string              `gorethink:"owner"`
	ProjectID   string              `gorethink:"project_id"`
	Birthtime   time.Time           `gorethink:"birthtime"`
	Description string              `gorethink:"description"`
	Notes       []schema.Note       `gorethink:"notes"`
	Properties  map[string]Property `gorethink:"properties"`
}
