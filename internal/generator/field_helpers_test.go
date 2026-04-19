package generator_test

import (
	"strings"
	"text/template"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// scalarField returns a FieldDescriptorProto for a primitive scalar.
func scalarField(name string, t descriptorpb.FieldDescriptorProto_Type, repeated bool) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	}
	return &descriptorpb.FieldDescriptorProto{
		Name:  proto.String(name),
		Type:  t.Enum(),
		Label: label.Enum(),
	}
}

// messageField returns a FieldDescriptorProto of message type referring to
// the given fully-qualified message type name (e.g. ".demo.Bar").
func messageField(name, typeName string, repeated bool) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	}
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		Label:    label.Enum(),
		TypeName: proto.String(typeName),
	}
}

// enumField returns a FieldDescriptorProto of enum type referring to the
// given fully-qualified enum type name (e.g. ".demo.Color").
func enumField(name, typeName string, repeated bool) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	}
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
		Label:    label.Enum(),
		TypeName: proto.String(typeName),
	}
}

// renderWithData executes a tiny template using the funcmap against `data`.
func renderWithData(src string, data any) (string, error) {
	tmpl, err := template.New("t").Funcs(generator.ProtoHelpersFuncMap).Parse(src)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

var _ = Describe("descriptor-walking helpers", func() {
	DescribeTable("goType on scalars",
		func(t descriptorpb.FieldDescriptorProto_Type, repeated bool, want string) {
			f := scalarField("x", t, repeated)
			out, err := renderWithData(`{{ goType "" . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, false, "int32"),
		Entry("repeated int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, true, "[]int32"),
		Entry("bool", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false, "bool"),
		Entry("string", descriptorpb.FieldDescriptorProto_TYPE_STRING, false, "string"),
		Entry("repeated string", descriptorpb.FieldDescriptorProto_TYPE_STRING, true, "[]string"),
		Entry("float32", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, false, "float32"),
		Entry("float64", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false, "float64"),
		Entry("uint64", descriptorpb.FieldDescriptorProto_TYPE_UINT64, false, "uint64"),
	)

	It("goType with package prefixes message types", func() {
		f := messageField("b", ".demo.Bar", false)
		out, err := renderWithData(`{{ goType "pkg" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("*pkg.Bar"))
	})

	It("goType on repeated messages emits slice-of-pointer", func() {
		f := messageField("b", ".demo.Bar", true)
		out, err := renderWithData(`{{ goType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("[]*Bar"))
	})

	DescribeTable("isFieldMessage / isFieldRepeated",
		func(f *descriptorpb.FieldDescriptorProto, wantMsg, wantRepeated string) {
			out, err := renderWithData(`{{ isFieldMessage . }}|{{ isFieldRepeated . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(wantMsg + "|" + wantRepeated))
		},
		Entry("scalar singular", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false), "false", "false"),
		Entry("scalar repeated", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, true), "false", "true"),
		Entry("message singular", messageField("x", ".demo.Bar", false), "true", "false"),
		Entry("message repeated", messageField("x", ".demo.Bar", true), "true", "true"),
	)

	It("haskellType renders primitives", func() {
		f := scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false)
		out, err := renderWithData(`{{ haskellType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Text"))
	})

	DescribeTable("haskellType on scalars",
		func(t descriptorpb.FieldDescriptorProto_Type, repeated bool, want string) {
			f := scalarField("x", t, repeated)
			out, err := renderWithData(`{{ haskellType "" . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("double", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false, "Float"),
		Entry("repeated double", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, true, "[Float]"),
		Entry("float", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, false, "Float"),
		Entry("int64", descriptorpb.FieldDescriptorProto_TYPE_INT64, false, "Int64"),
		Entry("repeated int64", descriptorpb.FieldDescriptorProto_TYPE_INT64, true, "[Int64]"),
		Entry("uint64", descriptorpb.FieldDescriptorProto_TYPE_UINT64, false, "Word"),
		Entry("int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, false, "Int"),
		Entry("repeated int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, true, "[Int]"),
		Entry("uint32 repeated", descriptorpb.FieldDescriptorProto_TYPE_UINT32, true, "[Word]"),
		Entry("bool", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false, "Bool"),
		Entry("bool repeated", descriptorpb.FieldDescriptorProto_TYPE_BOOL, true, "[Bool]"),
		Entry("repeated string", descriptorpb.FieldDescriptorProto_TYPE_STRING, true, "[Text]"),
		Entry("bytes", descriptorpb.FieldDescriptorProto_TYPE_BYTES, false, "Word8"),
		Entry("repeated bytes", descriptorpb.FieldDescriptorProto_TYPE_BYTES, true, "[Word8]"),
	)

	It("haskellType on message with pkg", func() {
		f := messageField("b", ".demo.Bar", false)
		out, err := renderWithData(`{{ haskellType "pkg" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("pkg.Bar"))
	})

	It("haskellType on repeated message", func() {
		f := messageField("b", ".demo.Bar", true)
		out, err := renderWithData(`{{ haskellType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("[Bar]"))
	})

	It("haskellType on enum", func() {
		f := enumField("c", ".demo.Color", false)
		out, err := renderWithData(`{{ haskellType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("Color"))
	})

	DescribeTable("rustType on scalars",
		func(t descriptorpb.FieldDescriptorProto_Type, repeated bool, want string) {
			f := scalarField("x", t, repeated)
			out, err := renderWithData(`{{ rustType "" . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("double", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false, "f64"),
		Entry("float", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, false, "f32"),
		Entry("int64", descriptorpb.FieldDescriptorProto_TYPE_INT64, false, "i64"),
		Entry("uint64", descriptorpb.FieldDescriptorProto_TYPE_UINT64, false, "u64"),
		Entry("int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, false, "i32"),
		Entry("uint32", descriptorpb.FieldDescriptorProto_TYPE_UINT32, false, "u32"),
		Entry("bool", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false, "bool"),
		Entry("string", descriptorpb.FieldDescriptorProto_TYPE_STRING, false, "String"),
		Entry("repeated string", descriptorpb.FieldDescriptorProto_TYPE_STRING, true, "Vec<String>"),
		Entry("bytes", descriptorpb.FieldDescriptorProto_TYPE_BYTES, false, "Vec<u8>"),
		Entry("repeated int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, true, "Vec<i32>"),
	)

	It("rustType on message with package", func() {
		f := messageField("b", ".demo.Bar", false)
		out, err := renderWithData(`{{ rustType "pkg" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("pkg.Bar"))
	})

	It("rustType on repeated message", func() {
		f := messageField("b", ".demo.Bar", true)
		out, err := renderWithData(`{{ rustType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("Vec<Bar>"))
	})

	It("rustType on enum", func() {
		f := enumField("c", ".demo.Color", false)
		out, err := renderWithData(`{{ rustType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("Color"))
	})

	DescribeTable("cppType on scalars",
		func(t descriptorpb.FieldDescriptorProto_Type, repeated bool, want string) {
			f := scalarField("x", t, repeated)
			out, err := renderWithData(`{{ cppType "" . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("double", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false, "double"),
		Entry("float", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, false, "float"),
		Entry("int64", descriptorpb.FieldDescriptorProto_TYPE_INT64, false, "int64_t"),
		Entry("uint64", descriptorpb.FieldDescriptorProto_TYPE_UINT64, false, "uint64_t"),
		Entry("int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, false, "int32_t"),
		Entry("uint32", descriptorpb.FieldDescriptorProto_TYPE_UINT32, false, "uint32_t"),
		Entry("bool", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false, "bool"),
		Entry("string", descriptorpb.FieldDescriptorProto_TYPE_STRING, false, "std::string"),
		Entry("bytes", descriptorpb.FieldDescriptorProto_TYPE_BYTES, false, "std::vector<uint8_t>"),
		Entry("repeated int32", descriptorpb.FieldDescriptorProto_TYPE_INT32, true, "std::vector<int32_t>"),
	)

	It("cppType on message with package", func() {
		f := messageField("b", ".demo.Bar", false)
		out, err := renderWithData(`{{ cppType "pkg" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("pkg.Bar"))
	})

	It("cppType on repeated message", func() {
		f := messageField("b", ".demo.Bar", true)
		out, err := renderWithData(`{{ cppType "" . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("std::vector<Bar>"))
	})

	DescribeTable("jsType",
		func(f *descriptorpb.FieldDescriptorProto, want string) {
			out, err := renderWithData(`{{ jsType . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("string", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false), "string"),
		Entry("repeated string", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, true), "Array<string>"),
		Entry("int32", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_INT32, false), "number"),
		Entry("double", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false), "number"),
		Entry("bool", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false), "boolean"),
		Entry("bytes", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_BYTES, false), "Uint8Array"),
		Entry("message", messageField("b", ".demo.Bar", false), "demo$Bar"),
		Entry("repeated message", messageField("b", ".demo.Bar", true), "Array<demo$Bar>"),
		Entry("enum", enumField("c", ".demo.Color", false), "demo$Color"),
	)

	DescribeTable("goZeroValue",
		func(f *descriptorpb.FieldDescriptorProto, want string) {
			out, err := renderWithData(`{{ goZeroValue . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("repeated string → nil", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, true), "nil"),
		Entry("double → 0.0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, false), "0.0"),
		Entry("float → 0.0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_FLOAT, false), "0.0"),
		Entry("int64 → 0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_INT64, false), "0"),
		Entry("uint64 → 0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_UINT64, false), "0"),
		Entry("int32 → 0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_INT32, false), "0"),
		Entry("uint32 → 0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_UINT32, false), "0"),
		Entry("bool → false", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_BOOL, false), "false"),
		Entry(`string → ""`, scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false), `""`),
		Entry("message → nil", messageField("b", ".demo.Bar", false), "nil"),
		Entry("bytes → 0", scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_BYTES, false), "0"),
		Entry("enum → nil", enumField("c", ".demo.Color", false), "nil"),
	)

	// goTypeWithEmbedded is unexported; it's reached via goTypeWithGoPackage.
	Describe("goTypeWithEmbedded (via goTypeWithGoPackage)", func() {
		newFile := func(pkg, goPackage string) *descriptorpb.FileDescriptorProto {
			return &descriptorpb.FileDescriptorProto{
				Package: proto.String(pkg),
				Options: &descriptorpb.FileOptions{GoPackage: proto.String(goPackage)},
			}
		}

		It("joins embedded-message segments with underscore", func() {
			// file package "demo"; field type ".demo.GetArticleResponse.Storage"
			// → embedded name "GetArticleResponse_Storage"
			f := messageField("s", ".demo.GetArticleResponse.Storage", true)
			p := newFile("demo", "example.com/demo;demopb")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("[]*demopb.GetArticleResponse_Storage"))
		})

		It("renders embedded enums with underscore join", func() {
			f := enumField("k", ".demo.GetArticleResponse.Kind", false)
			p := newFile("demo", "example.com/demo;demopb")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*demopb.GetArticleResponse_Kind"))
		})
	})

	Describe("goTypeWithPackage", func() {
		It("prefixes message with second package segment", func() {
			f := messageField("b", ".demo.Bar", false)
			out, err := renderWithData(`{{ goTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*demo.Bar"))
		})

		It("uses the timestamp alias for google.protobuf.Timestamp", func() {
			f := messageField("t", ".google.protobuf.Timestamp", false)
			out, err := renderWithData(`{{ goTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*timestamp.Timestamp"))
		})

		It("returns an unprefixed scalar", func() {
			f := scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_INT32, false)
			out, err := renderWithData(`{{ goTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("int32"))
		})
	})

	Describe("goTypeWithGoPackage", func() {
		newFile := func(goPackage string) *descriptorpb.FileDescriptorProto {
			return &descriptorpb.FileDescriptorProto{
				Package: proto.String("demo"),
				Options: &descriptorpb.FileOptions{GoPackage: proto.String(goPackage)},
			}
		}

		It("uses the full go_package when no alias is set", func() {
			f := messageField("b", ".demo.Bar", false)
			p := newFile("example.com/demo")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*example.com/demo.Bar"))
		})

		It("uses the alias when go_package has `;alias`", func() {
			f := messageField("b", ".demo.Bar", false)
			p := newFile("example.com/demo;demopb")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*demopb.Bar"))
		})

		It("maps Timestamp to the timestamp alias", func() {
			f := messageField("t", ".google.protobuf.Timestamp", false)
			p := newFile("example.com/demo")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("*timestamp.Timestamp"))
		})

		It("passes scalars through unchanged", func() {
			f := scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false)
			p := newFile("example.com/demo")
			out, err := renderWithData(
				`{{ goTypeWithGoPackage (index . 0) (index . 1) }}`,
				[]any{p, f},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("string"))
		})
	})

	Describe("rustTypeWithPackage / cppTypeWithPackage", func() {
		It("rustTypeWithPackage prefixes with package segment", func() {
			f := messageField("b", ".demo.Bar", false)
			out, err := renderWithData(`{{ rustTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("demo.Bar"))
		})

		It("rustTypeWithPackage maps Timestamp", func() {
			f := messageField("t", ".google.protobuf.Timestamp", false)
			out, err := renderWithData(`{{ rustTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("timestamp.Timestamp"))
		})

		It("cppTypeWithPackage prefixes with package segment", func() {
			f := messageField("b", ".demo.Bar", false)
			out, err := renderWithData(`{{ cppTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("demo.Bar"))
		})

		It("passes scalars through", func() {
			f := scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_INT32, false)
			out, err := renderWithData(`{{ rustTypeWithPackage . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("i32"))
		})
	})
})
