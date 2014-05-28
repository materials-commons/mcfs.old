package dir

// PatchType denotes the kind of patch operation
type PatchType int

const (
	// PatchCreate created item
	PatchCreate PatchType = iota

	// PatchDelete deleted item
	PatchDelete

	// PatchEdit item content was changed
	PatchEdit

	// PatchConflict there is a conflict with the specified file
	PatchConflict
)

// Patch is an instance of a difference when comparing two directories. It specifies
// the kind of change to apply.
type Patch struct {
	File FileInfo
	Type PatchType
}
