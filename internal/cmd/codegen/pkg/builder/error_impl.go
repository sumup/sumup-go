package builder

import (
	"fmt"
	"strings"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

type errorImplementation struct {
	Typ *TypeDeclaration
}

func (e errorImplementation) String() string {
	if e.Typ == nil {
		return ""
	}

	if len(e.Typ.Fields) == 0 {
		return fmt.Sprintf("\nfunc (e *%s) Error() string {\n\treturn fmt.Sprintf(\"\")\n}\n", e.Typ.Name)
	}

	parts := make([]string, 0, len(e.Typ.Fields))
	args := make([]string, 0, len(e.Typ.Fields))
	for _, field := range e.Typ.Fields {
		parts = append(parts, fmt.Sprintf("%s=%%v", field.Name))
		args = append(args, fmt.Sprintf("e.%s", strcase.ToCamel(field.Name)))
	}

	return fmt.Sprintf(
		"\nfunc (e *%s) Error() string {\n\treturn fmt.Sprintf(%q, %s)\n}\n",
		e.Typ.Name,
		strings.Join(parts, ", "),
		strings.Join(args, ", "),
	)
}

type staticErrorImplementation struct {
	Typ  string
	Name string
}

func (e staticErrorImplementation) String() string {
	return fmt.Sprintf("\nfunc (e *%s) Error() string {\n\treturn fmt.Sprintf(%q)\n}\n", e.Typ, e.Name)
}

type typeAssertionDeclaration struct {
	typ string
}

func (t typeAssertionDeclaration) String() string {
	return fmt.Sprintf("\nvar _ error = (*%s)(nil)\n", t.typ)
}
