package ds

// Marshallable ...
type Marshallable interface {
	// MarshalJSON returns the JSON bytes of this map.
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON consumes a slice of JSON bytes to populate this map.
	UnmarshalJSON(b []byte) error
}
