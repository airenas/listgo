package upload

// RecognizerMap provides the recognizer ID by key
type RecognizerMap interface {
	Get(key string) (string, error)
}
