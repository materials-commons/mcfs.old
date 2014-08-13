package common

import "fmt"

// Type is the different types of elements in the app data model.
type Type int

const (
	// ProjectType data model element is a Project
	ProjectType = iota + 1

	// SampleType data model element is a Sample
	SampleType

	// ReviewType data model element is a Review
	ReviewType

	// FileType data model element is a File
	FileType

	// DirectoryType data model element is a Directory
	DirectoryType

	// ProcessType data model element is a Process
	ProcessType
)

func (t Type) String() string {
	switch t {
	case ProjectType:
		return "ProjectType"
	case SampleType:
		return "SampleType"
	case ReviewType:
		return "ReviewType"
	case FileType:
		return "FileType"
	case DirectoryType:
		return "DirectoryType"
	case ProcessType:
		return "ProcessType"
	default:
		return fmt.Sprintf("Unknown Type: %d", e)
	}
}
