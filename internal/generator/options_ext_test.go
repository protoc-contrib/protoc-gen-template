package generator_test

import (
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("options_ext extractors", func() {
	It("stringMethodOptionsExtension returns empty on nil options", func() {
		out, err := renderWithData(
			`{{ stringMethodOptionsExtension 50001 . }}`,
			&descriptorpb.MethodDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(""))
	})

	It("boolMethodOptionsExtension returns false on nil options", func() {
		out, err := renderWithData(
			`{{ boolMethodOptionsExtension 50002 . }}`,
			&descriptorpb.MethodDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("false"))
	})

	It("stringFileOptionsExtension returns empty on nil options", func() {
		out, err := renderWithData(
			`{{ stringFileOptionsExtension 50003 . }}`,
			&descriptorpb.FileDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(""))
	})

	It("stringFieldExtension returns empty on nil options", func() {
		out, err := renderWithData(
			`{{ stringFieldExtension 50004 . }}`,
			&descriptorpb.FieldDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(""))
	})

	It("int64FieldExtension returns 0 on nil options", func() {
		out, err := renderWithData(
			`{{ int64FieldExtension 50005 . }}`,
			&descriptorpb.FieldDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("0"))
	})

	It("int64MessageExtension returns 0 on nil options", func() {
		out, err := renderWithData(
			`{{ int64MessageExtension 50006 . }}`,
			&descriptorpb.DescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("0"))
	})

	It("stringMessageExtension returns empty on nil options", func() {
		out, err := renderWithData(
			`{{ stringMessageExtension 50007 . }}`,
			&descriptorpb.DescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(""))
	})

	It("boolFieldExtension returns false on nil options", func() {
		out, err := renderWithData(
			`{{ boolFieldExtension 50008 . }}`,
			&descriptorpb.FieldDescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("false"))
	})

	It("boolMessageExtension returns false on nil options", func() {
		out, err := renderWithData(
			`{{ boolMessageExtension 50009 . }}`,
			&descriptorpb.DescriptorProto{},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("false"))
	})

	It("returns zero values for fieldIDs no option declares", func() {
		out, err := renderWithData(
			`{{ stringMethodOptionsExtension 99701 . }}|{{ boolMethodOptionsExtension 99701 . }}`,
			&descriptorpb.MethodDescriptorProto{Options: &descriptorpb.MethodOptions{}},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("|false"))
	})

	Context("on a request compiled by protoc", func() {
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

		It("reads file and method options sharing a field number", func() {
			Expect(render(`{{ stringFileOptionsExtension 60000 .File }}|{{ stringMethodOptionsExtension 60000 (index .Service.Method 0) }}`)).
				To(Equal("team-books|/books/{title}"))
		})

		It("reads method options", func() {
			Expect(render(`{{ boolMethodOptionsExtension 60002 (index .Service.Method 0) }}`)).To(Equal("true"))
		})

		It("reads message options", func() {
			Expect(render(`{{ $m := index .File.MessageType 0 }}{{ int64MessageExtension 60001 $m }}|{{ stringMessageExtension 60002 $m }}|{{ boolMessageExtension 55003 (index .File.MessageType 1) }}`)).
				To(Equal("3600|A book.|true"))
		})

		It("reads field options", func() {
			Expect(render(`{{ $f := index (index .File.MessageType 0).Field 5 }}{{ stringFieldExtension 60000 $f }}|{{ int64FieldExtension 60001 $f }}|{{ boolFieldExtension 60002 $f }}`)).
				To(Equal("book_notes|280|true"))
		})

		It("returns zero values for options of another type", func() {
			Expect(render(`{{ $f := index (index .File.MessageType 0).Field 5 }}{{ stringFieldExtension 60001 $f }}|{{ int64FieldExtension 60000 $f }}|{{ boolFieldExtension 60000 $f }}`)).
				To(Equal("|0|false"))
		})
	})
})
