package sweetpl

import (
	"fmt"
)

type TemplateError struct {
	Format     string
	Parameters []interface{}
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf(e.Format, e.Parameters...)
}

func Errf(format string, parameters ...interface{}) error {
	return &TemplateError{
		Format:     format,
		Parameters: parameters,
	}
}
