/* Common Go code across all generated bindings */

type GError struct {
	Code int
	Message string
}

func (self GError) Error() string {
	// ???: include error code?
	return self.Message
}
