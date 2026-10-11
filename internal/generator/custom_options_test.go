package generator_test

import (
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate protoc --include_imports -I testdata/options -o testdata/options/image.binpb testdata/options/demo/v1/demo.proto

// optionsPlugin loads testdata/options/image.binpb, which protoc compiled from
// the protos next to it. Decoding it leaves the custom options as unknown
// fields, except google.api.field_behavior, whose Go bindings the plugin
// links: the same state a real request reaches the plugin in.
func optionsPlugin() *protogen.Plugin {
	b, err := os.ReadFile("testdata/options/image.binpb")
	Expect(err).NotTo(HaveOccurred())
	set := &descriptorpb.FileDescriptorSet{}
	Expect(proto.Unmarshal(b, set)).To(Succeed())
	return newPlugin([]string{"demo/v1/demo.proto"}, set.GetFile()...)
}

var _ = Describe("custom option helpers", func() {
	var tmplDir string

	BeforeEach(func() {
		tmplDir = GinkgoT().TempDir()
	})

	render := func(src string) string {
		writeTemplate(tmplDir, "out.txt.tmpl", src)
		plugin := optionsPlugin()
		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		return outputByName(plugin, "out.txt")
	}

	It("reads file options", func() {
		Expect(render(`{{ option "acme.owner" .File }}`)).To(Equal("team-books"))
	})

	It("reads message-valued options as maps", func() {
		Expect(render(`{{ option "acme.table" (index .File.MessageType 0) | toJson }}`)).
			To(Equal(`{"indexes":["title","author"],"labels":{"tier":"gold"},"name":"books"}`))
	})

	It("reads field options declared by protoc-types", func() {
		out := render(`
{{- range (index .File.MessageType 0).Field }}
{{ .Name }}={{ option "protoc.contrib.defaults.value" . | toJson }}
{{- end }}`)
		Expect(out).To(Equal(`
title=null
page_size={"int32":50}
format={"enum_name":"FORMAT_PAPERBACK"}
draft={"bool":false}
sequel={"message":{"initialize":true}}
notes=null
isbn=null
url=null`))
	})

	It("gives template access to fields of message-valued options", func() {
		Expect(render(`{{ (option "protoc.contrib.defaults.value" (index (index .File.MessageType 0).Field 1)).int32 }}`)).
			To(Equal("50"))
	})

	It("tells an unset option from one set to its zero value", func() {
		out := render(`
{{- range (index .File.MessageType 0).Field }}
{{ .Name }}={{ hasOption "protoc.contrib.defaults.value" . }}
{{- end }}`)
		Expect(out).To(ContainSubstring("draft=true"))
		Expect(out).To(ContainSubstring("notes=false"))
	})

	It("reads options whose Go bindings the plugin links", func() {
		Expect(render(`{{ option "google.api.field_behavior" (index (index .File.MessageType 0).Field 0) | toJson }}`)).
			To(Equal(`["REQUIRED"]`))
	})

	It("reads oneof options", func() {
		Expect(render(`{{ option "protoc.contrib.defaults.oneof" (index (index .File.MessageType 0).OneofDecl 0) }}`)).
			To(Equal("isbn"))
	})

	It("reads enum and enum value options", func() {
		Expect(render(`{{ $e := index .File.EnumType 0 }}{{ option "protoc.contrib.json.enum_skip" $e }} {{ option "acme.label" (index $e.Value 1) }}`)).
			To(Equal("true Paperback"))
	})

	It("reads service and method options", func() {
		Expect(render(`{{ option "acme.internal" .Service }} {{ option "acme.timeout_ms" (index .Service.Method 0) }}`)).
			To(Equal("true 1500"))
	})

	It("reads options through the registry types getMessageType returns", func() {
		Expect(render(`{{ (option "acme.table" (getMessageType .File ".demo.v1.Book")).name }}`)).
			To(Equal("books"))
	})

	It("accepts a leading dot in the option name", func() {
		Expect(render(`{{ option ".acme.owner" .File }}`)).To(Equal("team-books"))
	})

	It("lists every custom option, keeping options of the same name apart", func() {
		Expect(render(`{{ options (index .File.MessageType 1) | toJson }}`)).
			To(Equal(`{"protoc.contrib.defaults.skip":true,"protoc.contrib.json.skip":true}`))
	})

	It("returns nothing for a descriptor without options", func() {
		Expect(render(`{{ option "acme.owner" (index (index .File.MessageType 0).Field 6) }}|{{ options (index (index .File.MessageType 0).Field 6) | toJson }}`)).
			To(Equal("<no value>|{}"))
	})

	It("fails on a value that isn't a descriptor", func() {
		writeTemplate(tmplDir, "out.txt.tmpl", `{{ option "acme.owner" .File.Name }}`)
		err := generator.Generate(optionsPlugin(), &generator.Options{TemplateDir: tmplDir})
		Expect(err).To(MatchError(ContainSubstring("*string is not a descriptor")))
	})
})
