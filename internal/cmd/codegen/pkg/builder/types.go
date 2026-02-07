package builder

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/sumup/sumup-go/internal/cmd/codegen/internal/strcase"
)

type Writable interface {
	String() string
}

func (tt *TypeDeclaration) String() string {
	buf := new(strings.Builder)
	if tt.Comment != "" {
		fmt.Fprintf(buf, "// %s\n", tt.Comment)
	}
	fmt.Fprintf(buf, "type %s %s", tt.Name, tt.Type)
	if tt.Fields != nil {
		slices.SortFunc(tt.Fields, func(a, b StructField) int {
			return strings.Compare(a.Name, b.Name)
		})
		fmt.Fprint(buf, " {\n")
		for _, ft := range tt.Fields {
			fmt.Fprint(buf, ft.String())
			fmt.Fprint(buf, "\n")
		}
		fmt.Fprint(buf, "}")
	}
	fmt.Fprint(buf, "\n")
	return buf.String()
}

func (o *OneOfDeclaration) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "type %s struct {\n", o.Name)
	for _, opt := range o.Options {
		fmt.Fprintf(buf, "\t%s *%s\n", opt, opt)
	}
	fmt.Fprintf(buf, "}\n\n")
	for _, opt := range o.Options {
		fmt.Fprintf(buf, "func (r *%s) As%s() (*%s, bool) {\n", o.Name, opt, opt)
		fmt.Fprintf(buf, "\tif r.%s != nil {\n", opt)
		fmt.Fprintf(buf, "\t\treturn r.%s, true\n", opt)
		fmt.Fprint(buf, "\t}\n\n")
		fmt.Fprint(buf, "\t\treturn nil, false\n")
		fmt.Fprint(buf, "\t}\n\n")
	}
	return buf.String()
}

func (f *StructField) String() string {
	buf := new(strings.Builder)
	if f.Comment != "" {
		fmt.Fprintf(buf, "// %s\n", f.Comment)
	}
	name := f.Name

	// TODO: extract into helper
	if strings.HasPrefix(name, "+") {
		name = strings.Replace(name, "+", "Plus", 1)
	}
	if strings.HasPrefix(name, "-") {
		name = strings.Replace(name, "-", "Minus", 1)
	}
	if strings.HasPrefix(name, "@") {
		name = strings.Replace(name, "@", "At", 1)
	}
	if strings.HasPrefix(name, "$") {
		name = strings.Replace(name, "$", "", 1)
	}

	name = strcase.ToCamel(name)
	if f.Pointer {
		fmt.Fprintf(buf, "\t%s *%s", name, f.Type)
	} else {
		fmt.Fprintf(buf, "\t%s %s", name, f.Type)
	}
	if len(f.Tags) > 0 {
		fmt.Fprint(buf, " `")
		for k, v := range f.Tags {
			fmt.Fprintf(buf, "%s:%q", k, strings.Join(v, ","))
		}
		fmt.Fprint(buf, "`")
	}

	return buf.String()
}

func (et *EnumDeclaration[E]) String() string {
	buf := new(strings.Builder)
	fmt.Fprint(buf, et.Type.String())
	fmt.Fprint(buf, "\nconst (\n")
	slices.SortFunc(et.Values, func(a, b EnumOption[E]) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, v := range et.Values {
		fmt.Fprintf(buf, "\t%s %s = %#v\n", v.Name, et.Type.Name, v.Value)
	}
	fmt.Fprint(buf, ")\n")
	return buf.String()
}

func dereferenceSchema(ref *base.SchemaProxy) *base.SchemaProxy {
	if ref == nil {
		return nil
	}
	if ref.IsReference() || ref.Schema() == nil {
		return ref
	}
	if len(ref.Schema().AllOf) > 0 {
		return dereferenceSchema(ref.Schema().AllOf[0])
	}
	return ref
}

func paramToString(name string, param *v3.Parameter) string {
	if param == nil || param.Schema == nil {
		return name
	}

	schema := dereferenceSchema(param.Schema)
	if schema == nil {
		return name
	}

	// HACK: also handles component references wrapped via allOf used for nullable enums.
	if schema.IsReference() {
		return fmt.Sprintf("string(%s)", name)
	}

	if schema.Schema() == nil {
		return name
	}

	schemaValue := schema.Schema()

	switch {
	case slices.Contains(schemaValue.Type, "string"):
		if isNumericStringFormat(schemaValue.Format) {
			return fmt.Sprintf("%s.String()", name)
		}

		switch schemaValue.Format {
		case "date-time":
			name = strings.TrimPrefix(name, "*")
			return fmt.Sprintf("%s.Format(time.RFC3339)", name)
		case "date":
			name = strings.TrimPrefix(name, "*")
			return fmt.Sprintf("%s.String()", name)
		case "time":
			name = strings.TrimPrefix(name, "*")
			return fmt.Sprintf("%s.String()", name)
		default:
			return name
		}
	case slices.Contains(schemaValue.Type, "integer"):
		switch schemaValue.Format {
		case "int32":
			return fmt.Sprintf("strconv.FormatInt(int64(%s), 10)", name)
		case "int64":
			return fmt.Sprintf("strconv.FormatInt(%s, 10)", name)
		default:
			return fmt.Sprintf("strconv.Itoa(%s)", name)
		}
	case slices.Contains(schemaValue.Type, "boolean"):
		return fmt.Sprintf("strconv.FormatBool(%s)", name)
	case slices.Contains(schemaValue.Type, "number"):
		switch schemaValue.Format {
		case "float":
			return fmt.Sprintf("strconv.FormatFloat(float64(%s), 'f', -1, 32)", name)
		case "double":
			return fmt.Sprintf("strconv.FormatFloat(%s, 'f', -1, 64)", name)
		default:
			return fmt.Sprintf("strconv.FormatFloat(%s, 'f', -1, 64)", name)
		}
	case slices.Contains(schemaValue.Type, "array"):
		// For array items that are schema references (e.g., enums), we need to convert each item to string
		if schemaValue.Items != nil && schemaValue.Items.IsA() && schemaValue.Items.A != nil && schemaValue.Items.A.IsReference() {
			return fmt.Sprintf("string(%s)", name)
		}
		return name
	default:
		slog.Warn("need to implement conversion for",
			slog.String("ref", schema.GetReference()),
			slog.String("type", strings.Join(schemaValue.Type, ",")),
			slog.String("name", name),
		)
		return name
	}
}

type toQueryValues struct {
	Typ *TypeDeclaration
}

func (e toQueryValues) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "// QueryValues converts [%s] into [url.Values].\n", e.Typ.Name)
	fmt.Fprintf(buf, "func (p *%s) QueryValues() url.Values {\n", e.Typ.Name)
	fmt.Fprintf(buf, "\tq := make(url.Values)\n\n")
	for _, f := range e.Typ.Fields {
		name := strcase.ToCamel(f.Name)
		if f.Parameter.Schema != nil && f.Parameter.Schema.Schema() != nil && slices.Contains(f.Parameter.Schema.Schema().Type, "array") {
			field := fmt.Sprintf("p.%s", name)
			fmt.Fprintf(buf, "\tfor _, v := range %s {\n", field)
			fmt.Fprintf(buf, "\t\tq.Add(%q, %s)\n", f.Name, paramToString("v", f.Parameter))
			fmt.Fprintf(buf, "\t}\n")
		} else {
			required := f.Parameter.Required != nil && *f.Parameter.Required
			if required {
				field := fmt.Sprintf("p.%s", name)
				fmt.Fprintf(buf, "\tq.Set(%q, %s)\n", f.Name, paramToString(field, f.Parameter))
			} else {
				fmt.Fprintf(buf, "\tif p.%s != nil {\n", name)
				field := fmt.Sprintf("*p.%s", name)
				fmt.Fprintf(buf, "\t\tq.Set(%q, %s)\n", f.Name, paramToString(field, f.Parameter))
				fmt.Fprintf(buf, "\t}\n")
			}
		}
		fmt.Fprint(buf, "\n")
	}
	fmt.Fprint(buf, "\treturn q\n")
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}
