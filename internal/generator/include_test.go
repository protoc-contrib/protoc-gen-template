package generator_test

import (
	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("include and partials", func() {
	var tmplDir string

	BeforeEach(func() {
		tmplDir = GinkgoT().TempDir()
	})

	generate := func() (*generatorResult, error) {
		plugin := newPlugin([]string{"demo.proto"}, fileDescProto("demo.proto", true))
		err := generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})
		return &generatorResult{names: outputNames(plugin), out: outputByName(plugin, "out.txt")}, err
	}

	It("returns a template's output, so it can be piped and tested", func() {
		writeTemplate(tmplDir, "out.txt.tmpl", `
{{- define "greet" }}  hello {{ . }}  {{ end -}}
{{ include "greet" .Service.GetName | trim | upper }}|{{ if include "greet" "x" | trim }}yes{{ end }}`)
		res, err := generate()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.out).To(Equal("HELLO GREETER|yes"))
	})

	It("lets a template call itself", func() {
		writeTemplate(tmplDir, "out.txt.tmpl", `
{{- define "down" }}{{ if gt . 0 }}{{ . }} {{ include "down" (sub . 1) }}{{ end }}{{ end -}}
{{ include "down" 3 }}`)
		res, err := generate()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.out).To(Equal("3 2 1 "))
	})

	It("fails cleanly on recursion that never ends", func() {
		writeTemplate(tmplDir, "out.txt.tmpl", `
{{- define "forever" }}{{ include "forever" . }}{{ end -}}
{{ include "forever" . }}`)
		_, err := generate()
		Expect(err).To(MatchError(ContainSubstring(`include: "forever" called itself more than 1000 times`)))
		Expect(len(err.Error())).To(BeNumerically("<", 500))
	})

	It("loads .tpl partials into every template without rendering them", func() {
		writeTemplate(tmplDir, "_helpers.tpl", `{{ define "shout" }}{{ . | upper }}!{{ end }}`)
		writeTemplate(tmplDir, "nested/more.tpl", `{{ define "whisper" }}{{ . | lower }}{{ end }}`)
		writeTemplate(tmplDir, "out.txt.tmpl",
			`{{ include "shout" .Service.GetName }} {{ template "whisper" "QUIET" }}`)
		res, err := generate()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.names).To(ConsistOf("out.txt"))
		Expect(res.out).To(Equal("GREETER! quiet"))
	})

	It("still renders templates whose name starts with an underscore", func() {
		writeTemplate(tmplDir, "__init__.py.tmpl", `# {{ .Service.GetName }}`)
		res, err := generate()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.names).To(ConsistOf("__init__.py"))
	})

	It("walks messages recursively, as in the issue", func() {
		// Book reaches a REQUIRED field directly (title) and contains itself
		// (sequel), so the walk tracks the messages it has seen.
		writeTemplate(tmplDir, "_walk.tpl", `
{{- define "required" -}}
{{- $seen := append .seen .message.GetName -}}
{{- range .message.Field -}}
{{- if has "REQUIRED" (option "google.api.field_behavior" . | default list) }}{{ .Name }} {{ end -}}
{{- if and (eq (.GetType | toString) "TYPE_MESSAGE") (not (has (.GetTypeName | splitList "." | last) $seen)) -}}
{{- include "required" (dict "message" (getMessageType $.file .GetTypeName) "seen" $seen "file" $.file) -}}
{{- end -}}
{{- end -}}
{{- end -}}`)
		writeTemplate(tmplDir, "out.txt.tmpl", `
{{- range .File.MessageType -}}
{{ .Name }}: {{ include "required" (dict "message" . "seen" list "file" $.File) | trim | default "none" }}
{{ end -}}`)
		plugin := optionsPlugin()
		Expect(generator.Generate(plugin, &generator.Options{TemplateDir: tmplDir})).To(Succeed())
		Expect(outputByName(plugin, "out.txt")).To(Equal("Book: title\nHand: none\n"))
	})

	It("fails outside of rendering", func() {
		_, err := renderWithData(`{{ include "x" . }}`, nil)
		Expect(err).To(MatchError(ContainSubstring("include is only available while rendering a template")))
	})
})

type generatorResult struct {
	names   []string
	out     string
	outputs map[string]string
}
