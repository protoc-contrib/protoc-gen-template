package generator_test

import (
	"github.com/protoc-contrib/protoc-gen-template/internal/generator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options.Set", func() {
	var opts *generator.Options

	BeforeEach(func() {
		opts = &generator.Options{}
	})

	It("accepts template_dir", func() {
		Expect(opts.Set("template_dir", "/tmp/tmpl")).To(Succeed())
		Expect(opts.TemplateDir).To(Equal("/tmp/tmpl"))
	})


	It("accepts registry=true", func() {
		Expect(opts.Set("registry", "true")).To(Succeed())
		Expect(opts.Registry).To(BeTrue())
	})

	It("accepts registry=false", func() {
		Expect(opts.Set("registry", "false")).To(Succeed())
		Expect(opts.Registry).To(BeFalse())
	})

	DescribeTable("mode values",
		func(value string, expected generator.Mode) {
			Expect(opts.Set("mode", value)).To(Succeed())
			Expect(opts.Mode).To(Equal(expected))
		},
		Entry("service", "service", generator.ModeService),
		Entry("file", "file", generator.ModeFile),
		Entry("all", "all", generator.ModeAll),
	)

	It("rejects unknown mode values", func() {
		err := opts.Set("mode", "unknown")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown mode"))
	})

	It("rejects unknown options", func() {
		err := opts.Set("nope", "1")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown plugin option"))
	})

	It("rejects malformed booleans", func() {
		err := opts.Set("registry", "yeah")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid value"))
	})
})
