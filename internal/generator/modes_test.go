package generator_test

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("package and once modes", func() {
	var tmplDir string

	BeforeEach(func() {
		tmplDir = GinkgoT().TempDir()
	})

	// commentedFile declares one message, documented with a leading comment.
	commentedFile := func(name, pkg, message string) *descriptorpb.FileDescriptorProto {
		f := fileDescProto(name, false)
		f.Package = proto.String(pkg)
		f.MessageType = []*descriptorpb.DescriptorProto{{Name: proto.String(message)}}
		f.SourceCodeInfo = &descriptorpb.SourceCodeInfo{
			Location: []*descriptorpb.SourceCodeInfo_Location{{
				Path:            []int32{4, 0},
				Span:            []int32{0, 0, 0},
				LeadingComments: proto.String(" " + message + " docs\n"),
			}},
		}
		return f
	}

	// generate renders the template set over a.proto and b.proto in demo.v1
	// and c.proto in demo.v2, which import dep.proto, not to be generated.
	generate := func(mode generator.Mode) *generatorResult {
		plugin := newPlugin(
			[]string{"a.proto", "b.proto", "c.proto"},
			commentedFile("dep.proto", "dep", "Dep"),
			commentedFile("a.proto", "demo.v1", "A"),
			commentedFile("b.proto", "demo.v1", "B"),
			commentedFile("c.proto", "demo.v2", "C"),
		)
		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir, Mode: mode})).To(Succeed())
		res := &generatorResult{names: outputNames(plugin)}
		res.outputs = map[string]string{}
		for _, name := range res.names {
			res.outputs[name] = outputByName(plugin, name)
		}
		return res
	}

	It("renders once per package with mode=package", func() {
		writeTemplate(tmplDir, "{{.Package}}.txt.tmpl",
			`{{ .Package }}:{{ range .Files }} {{ .GetName }}{{ end }} file={{ .File }} service={{ .Service }}`)

		res := generate(generator.ModePackage)
		Expect(res.names).To(ConsistOf("demo.v1.txt", "demo.v2.txt"))
		Expect(res.outputs["demo.v1.txt"]).To(Equal("demo.v1: a.proto b.proto file=<nil> service=<nil>"))
		Expect(res.outputs["demo.v2.txt"]).To(Equal("demo.v2: c.proto file=<nil> service=<nil>"))
	})

	It("renders once per run with mode=once", func() {
		writeTemplate(tmplDir, "mod.rs.tmpl", `
{{- range $pkg, $files := byPackage .Files }}
{{ $pkg }}:{{ range $files }} {{ .GetName }}{{ end }}
{{- end }}
packages={{ packages .Files | join "," }} package={{ .Package | quote }}`)

		res := generate(generator.ModeOnce)
		Expect(res.names).To(ConsistOf("mod.rs"))
		Expect(res.outputs["mod.rs"]).To(Equal(`
demo.v1: a.proto b.proto
demo.v2: c.proto
packages=demo.v1,demo.v2 package=""`))
	})

	It("reads comments from every file of the package", func() {
		writeTemplate(tmplDir, "{{.Package}}.txt.tmpl",
			`{{ range .Files }}{{ range .MessageType }}{{ leadingComment . | trim }};{{ end }}{{ end }}`)

		res := generate(generator.ModePackage)
		Expect(res.outputs["demo.v1.txt"]).To(Equal("A docs;B docs;"))
	})

	It("exposes the rendered file as .Files in the per-file modes", func() {
		writeTemplate(tmplDir, "{{.File.GetName}}.txt.tmpl",
			`{{ range .Files }}{{ .GetName }}{{ end }} {{ .Package }}`)

		res := generate(generator.ModeAll)
		Expect(res.names).To(ConsistOf("a.proto.txt", "b.proto.txt", "c.proto.txt"))
		Expect(res.outputs["b.proto.txt"]).To(Equal("b.proto demo.v1"))
	})
})
