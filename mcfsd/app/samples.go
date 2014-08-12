package app

import (
	"time"

	"github.com/materials-commons/mcfs/base/schema"
)

type Property struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Unit  string      `json:"unit"`
	Value interface{} `json:"value"`
}

type OID struct {
	Type string `json:"type"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Sample struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Owner       string              `json:"owner"`
	ProjectID   string              `json:"project_id"`
	Birthtime   time.Time           `json:"birthtime"`
	Description string              `json:"description"`
	Notes       []schema.Note       `json:"notes"`
	Properties  map[string]Property `json:"properties"`
	Projects    []OID               `json:"projects"`
}

type SamplesService interface {
	Create(Sample) error
	Get(id string) Sample
}

type samplesService struct {
}

func sampleToAppSample(s *schema.Sample) Sample {
	sample := Sample{
		ID:          s.ID,
		Name:        s.Name,
		Owner:       s.Owner,
		ProjectID:   s.ProjectID,
		Birthtime:   s.Birthtime,
		Description: s.Description,
		Notes:       append(sample.Notes, s.Notes...),
		Properties:  make(map[string]Property{}),
	}

	for k, v := range s.Properties {
		sample.Properties[k] = Property{
			Name:  v.Name,
			Type:  v.Type,
			Unit:  v.Unit,
			Value: v.Value,
		}
	}

	return sample
}
