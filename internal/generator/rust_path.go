package generator

import (
	"fmt"
	"strings"

	"github.com/huandu/xstrings"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// rustKeywords are the words prost and buffa escape in module and type names.
var rustKeywords = map[string]bool{
	"abstract": true, "as": true, "async": true, "await": true, "become": true,
	"box": true, "break": true, "const": true, "continue": true, "do": true,
	"dyn": true, "else": true, "enum": true, "extern": true, "false": true,
	"final": true, "fn": true, "for": true, "gen": true, "if": true,
	"impl": true, "in": true, "let": true, "loop": true, "macro": true,
	"match": true, "mod": true, "move": true, "mut": true, "override": true,
	"priv": true, "pub": true, "ref": true, "return": true, "static": true,
	"struct": true, "trait": true, "true": true, "try": true, "type": true,
	"typeof": true, "unsafe": true, "unsized": true, "use": true,
	"virtual": true, "where": true, "while": true, "yield": true,
}

// rustPath returns the Rust path of the message or enum name, e.g.
// ".acme.v1.Book.Format" under root "crate::buffa" is
// "crate::buffa::acme::v1::book::Format", the way prost and buffa lay out
// generated code: one module per package segment, and a snake_case module per
// parent message.
func rustPath(root, name string) (string, error) {
	desc, err := requestFiles.FindDescriptorByName(protoreflect.FullName(strings.TrimPrefix(name, ".")))
	if err != nil {
		return "", fmt.Errorf("rustPath: %q is not a type in the request: %w", name, err)
	}
	switch desc.(type) {
	case protoreflect.MessageDescriptor, protoreflect.EnumDescriptor:
	default:
		return "", fmt.Errorf("rustPath: %q is not a message or an enum", name)
	}

	var segments []string
	if root != "" {
		segments = append(segments, root)
	}
	if pkg := desc.ParentFile().Package(); pkg != "" {
		for _, part := range strings.Split(string(pkg), ".") {
			segments = append(segments, rustModule(part))
		}
	}

	var parents []string
	for parent := desc.Parent(); parent != nil; parent = parent.Parent() {
		if _, ok := parent.(protoreflect.MessageDescriptor); !ok {
			break
		}
		parents = append(parents, rustModule(string(parent.Name())))
	}
	for i := len(parents) - 1; i >= 0; i-- {
		segments = append(segments, parents[i])
	}

	return strings.Join(append(segments, rustTypeName(string(desc.Name()))), "::"), nil
}

func rustModule(name string) string {
	return rustIdent(xstrings.ToSnakeCase(name))
}

func rustTypeName(name string) string {
	return rustIdent(xstrings.ToPascalCase(xstrings.ToSnakeCase(name)))
}

// rustIdent escapes a keyword the way prost does: as a raw identifier, or with
// a trailing underscore for the keywords a raw identifier can't spell.
func rustIdent(ident string) string {
	switch {
	case ident == "self" || ident == "super" || ident == "crate" || ident == "Self":
		return ident + "_"
	case rustKeywords[ident]:
		return "r#" + ident
	}
	return ident
}
