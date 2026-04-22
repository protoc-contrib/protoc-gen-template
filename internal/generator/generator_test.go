package generator_test

import (
	"os"
	"path/filepath"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeTemplate writes content to dir/relative and returns the absolute path.
func writeTemplate(dir, relative, content string) {
	abs := filepath.Join(dir, relative)
	Expect(os.MkdirAll(filepath.Dir(abs), 0o755)).To(Succeed())
	Expect(os.WriteFile(abs, []byte(content), 0o644)).To(Succeed())
}

// fileDescProto builds a minimal FileDescriptorProto with one service when
// withService is true. It's handcrafted so the tests don't have to depend
// on protoc or compiled fixtures.
func fileDescProto(name string, withService bool) *descriptorpb.FileDescriptorProto {
	f := &descriptorpb.FileDescriptorProto{
		Name:    proto.String(name),
		Package: proto.String("demo"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("example.com/demo"),
		},
	}
	if withService {
		f.Service = []*descriptorpb.ServiceDescriptorProto{
			{Name: proto.String("Greeter")},
		}
	}
	return f
}

// newPlugin wires a protogen.Plugin from a handcrafted request. Only the
// files whose names appear in filesToGenerate are scheduled for emission.
func newPlugin(filesToGenerate []string, protoFiles ...*descriptorpb.FileDescriptorProto) *protogen.Plugin {
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: filesToGenerate,
		ProtoFile:      protoFiles,
	}
	plugin, err := protogen.Options{}.New(req)
	Expect(err).NotTo(HaveOccurred())
	return plugin
}

// outputNames collects every emitted filename on the response so tests can
// assert what Generate produced.
func outputNames(plugin *protogen.Plugin) []string {
	resp := plugin.Response()
	names := make([]string, 0, len(resp.File))
	for _, f := range resp.File {
		names = append(names, f.GetName())
	}
	return names
}

func outputByName(plugin *protogen.Plugin, name string) string {
	for _, f := range plugin.Response().File {
		if f.GetName() == name {
			return f.GetContent()
		}
	}
	return ""
}

var _ = Describe("generator.Generate", func() {
	var tmplDir string

	BeforeEach(func() {
		tmplDir = GinkgoT().TempDir()
	})

	It("renders a simple template once per service (default mode)", func() {
		writeTemplate(tmplDir, "greeter.txt.tmpl",
			`pkg={{.File.Package}} service={{.Service.Name}}`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", true),
		)

		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())

		Expect(outputNames(plugin)).To(ConsistOf("greeter.txt"))
		Expect(outputByName(plugin, "greeter.txt")).To(Equal("pkg=demo service=Greeter"))
	})

	It("emits nothing for files without a service in default mode", func() {
		writeTemplate(tmplDir, "greeter.txt.tmpl", `x`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", false),
		)

		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputNames(plugin)).To(BeEmpty())
	})

	It("renders even service-less files with mode=all", func() {
		writeTemplate(tmplDir, "doc.txt.tmpl", `pkg={{.File.Package}}`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", false),
		)

		Expect(generator.Generate(plugin, &generator.Options{
			TemplateDir: tmplDir,
			Mode:        generator.ModeAll,
		})).To(Succeed())

		Expect(outputNames(plugin)).To(ConsistOf("doc.txt"))
		Expect(outputByName(plugin, "doc.txt")).To(Equal("pkg=demo"))
	})

	It("renders once per file with mode=file even when the file defines multiple services", func() {
		writeTemplate(tmplDir, "file.txt.tmpl",
			`n={{len .File.Service}}`)

		f := fileDescProto("demo.proto", true)
		f.Service = append(f.Service, &descriptorpb.ServiceDescriptorProto{Name: proto.String("Echo")})

		plugin := newPlugin([]string{"demo.proto"}, f)

		Expect(generator.Generate(plugin, &generator.Options{
			TemplateDir: tmplDir,
			Mode:        generator.ModeFile,
		})).To(Succeed())

		Expect(outputNames(plugin)).To(ConsistOf("file.txt"))
		Expect(outputByName(plugin, "file.txt")).To(Equal("n=2"))
	})

	It("skips files without services with mode=file", func() {
		writeTemplate(tmplDir, "file.txt.tmpl", `x`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", false),
		)

		Expect(generator.Generate(plugin, &generator.Options{
			TemplateDir: tmplDir,
			Mode:        generator.ModeFile,
		})).To(Succeed())
		Expect(outputNames(plugin)).To(BeEmpty())
	})

	It("templates the output filename from .File fields", func() {
		writeTemplate(tmplDir, "{{.File.Package}}/{{.Service.Name}}.txt.tmpl",
			`hello`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", true),
		)

		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputNames(plugin)).To(ConsistOf("demo/Greeter.txt"))
	})

	It("walks nested template directories", func() {
		writeTemplate(tmplDir, "a/b/c.txt.tmpl", `ok`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", true),
		)

		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputNames(plugin)).To(ConsistOf("a/b/c.txt"))
	})

	It("ignores non-.tmpl files", func() {
		writeTemplate(tmplDir, "keep.txt.tmpl", `x`)
		writeTemplate(tmplDir, "ignore.md", `y`)

		plugin := newPlugin(
			[]string{"demo.proto"},
			fileDescProto("demo.proto", true),
		)

		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputNames(plugin)).To(ConsistOf("keep.txt"))
	})
})
