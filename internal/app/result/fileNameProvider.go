package result

// FileNameProvider provider audio file name by ID
type FileNameProvider interface {
	Get(ID string) (string, error)
}
