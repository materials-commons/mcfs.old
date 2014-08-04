package schema

type Tag struct {
	ID   string `gorethink:"id,omitempty"`
	Name string `gorethink:"name"`
}
