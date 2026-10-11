package generator_test

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rustPath", func() {
	var tmplDir string

	BeforeEach(func() {
		tmplDir = GinkgoT().TempDir()
	})

	// rustFile declares, in package acme.type.v1, a message Book with a
	// nested message Chapter, which nests an enum Kind, plus a message
	// HTTPRequest and a message named Self.
	rustFile := func() *descriptorpb.FileDescriptorProto {
		f := fileDescProto("acme/type/v1/book.proto", false)
		f.Package = proto.String("acme.type.v1")
		f.MessageType = []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Book"),
				NestedType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("Chapter"),
					EnumType: []*descriptorpb.EnumDescriptorProto{{
						Name:  proto.String("Kind"),
						Value: []*descriptorpb.EnumValueDescriptorProto{{Name: proto.String("KIND_UNSPECIFIED"), Number: proto.Int32(0)}},
					}},
				}},
			},
			{Name: proto.String("HTTPRequest")},
			{Name: proto.String("Self")},
		}
		return f
	}

	// render generates the last file; the others are its imports, which
	// protoc lists first.
	render := func(src string, files ...*descriptorpb.FileDescriptorProto) (string, error) {
		writeTemplate(tmplDir, "out.txt.tmpl", src)
		plugin := newPlugin([]string{files[len(files)-1].GetName()}, files...)
		if err := generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir, Mode: generator.ModeAll}); err != nil {
			return "", err
		}
		return outputByName(plugin, "out.txt"), nil
	}

	DescribeTable("maps a type name to its Rust path",
		func(root, name, want string) {
			out, err := render(`{{ rustPath "`+root+`" "`+name+`" }}`, rustFile())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(want))
		},
		Entry("a top-level message", "crate::buffa", ".acme.type.v1.Book", "crate::buffa::acme::r#type::v1::Book"),
		Entry("a nested message", "crate::buffa", ".acme.type.v1.Book.Chapter", "crate::buffa::acme::r#type::v1::book::Chapter"),
		Entry("a doubly nested enum", "crate::buffa", ".acme.type.v1.Book.Chapter.Kind", "crate::buffa::acme::r#type::v1::book::chapter::Kind"),
		Entry("an acronym", "crate", ".acme.type.v1.HTTPRequest", "crate::acme::r#type::v1::HttpRequest"),
		Entry("a name a raw identifier can't spell", "crate", ".acme.type.v1.Self", "crate::acme::r#type::v1::Self_"),
		Entry("no root", "", ".acme.type.v1.Book", "acme::r#type::v1::Book"),
		Entry("a name without the leading dot", "crate", "acme.type.v1.Book", "crate::acme::r#type::v1::Book"),
	)

	It("maps a type in a file without a package", func() {
		f := fileDescProto("bare.proto", false)
		f.Package = nil
		f.MessageType = []*descriptorpb.DescriptorProto{{Name: proto.String("Bare")}}
		Expect(render(`{{ rustPath "crate" ".Bare" }}`, f)).To(Equal("crate::Bare"))
	})

	It("maps types from imported files", func() {
		f := fileDescProto("main.proto", false)
		f.Dependency = []string{"acme/type/v1/book.proto"}
		Expect(render(`{{ rustPath "crate" ".acme.type.v1.Book" }}`, rustFile(), f)).
			To(Equal("crate::acme::r#type::v1::Book"))
	})

	It("works on method input types", func() {
		f := rustFile()
		f.Service = []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("Library"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("Get"),
				InputType:  proto.String(".acme.type.v1.HTTPRequest"),
				OutputType: proto.String(".acme.type.v1.Book"),
			}},
		}}
		Expect(render(`{{ range (index .File.Service 0).Method }}{{ rustPath "crate::buffa" .GetInputType }}{{ end }}`, f)).
			To(Equal("crate::buffa::acme::r#type::v1::HttpRequest"))
	})

	It("fails on a name that isn't a type in the request", func() {
		_, err := render(`{{ rustPath "crate" ".acme.type.v1.Missing" }}`, rustFile())
		Expect(err).To(MatchError(ContainSubstring(`".acme.type.v1.Missing" is not a type in the request`)))
	})

	It("fails on a name that isn't a message or an enum", func() {
		_, err := render(`{{ rustPath "crate" ".acme.type.v1.Book.Chapter.KIND_UNSPECIFIED" }}`, rustFile())
		Expect(err).To(MatchError(ContainSubstring("is not a message or an enum")))
	})
})

var _ = Describe("the template context", func() {
	It("exposes the file's package as a plain string", func() {
		tmplDir := GinkgoT().TempDir()
		writeTemplate(tmplDir, "out.txt.tmpl",
			`{{ .Package }} {{ eq (printf "%s.gen.rs" .Package) "demo.gen.rs" }}`)
		plugin := newPlugin([]string{"demo.proto"}, fileDescProto("demo.proto", true))
		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputByName(plugin, "out.txt")).To(Equal("demo true"))
	})
})
