package schema

import "time"

type Review struct {
	ID          string    `gorethink:"id,omitempty"`
	Birthtime   time.Time `gorethink:"birthtime"`
	ItemID      string    `gorethink:"item_id"`
	ItemName    string    `gorethink:"item_name"`
	ItemType    string    `gorethink:"item_type"`
	ProjectID   string    `gorethink:"project_id"`
	RequestedBy string    `gorethink:"requested_by"`
	RequestTo   string    `gorethink:"request_to"`
	Status      string    `gorethink:"status"`
	Notes       []Note    `gorethink:"notes"`
}
