package marshaling

// Marshaler defines the interface for marshaling.
type Marshaler interface {
	Marshal(interface{}) error
}

// Unmarshaler defines the interface for unmarshaling.
type Unmarshaler interface {
	Unmarshal(interface{}) error
}

// MarshalUnmarshaler defines the marshal/unmarshal interface.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}
