package utils

import "fmt"

type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key not found: %s", e.Key)
}

type GenericNotFoundError struct {
	Thing string
}

func (e *GenericNotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Thing)
}
