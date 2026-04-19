package generator_test

import (
	"google.golang.org/protobuf/types/descriptorpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Custom options extraction is only meaningful against descriptors produced
// from real proto sources with matching Go stubs. Exercising the happy path
// requires heavy fixtures, so these tests focus on the nil/absent paths that
// user templates hit most often.
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

	// When Options is set but the extension fieldID isn't registered with a
	// matching type, the registration path in getExtension runs and
	// GetExtension returns the zero value; the type assertion on a nil
	// pointer fails, so the extractor returns its own default.
	It("stringMethodOptionsExtension registers unknown fieldIDs and returns empty", func() {
		m := &descriptorpb.MethodDescriptorProto{
			Options: &descriptorpb.MethodOptions{},
		}
		// Use a high, unique fieldID so the first call registers and a second
		// call hits the cached path.
		out1, err := renderWithData(`{{ stringMethodOptionsExtension 99701 . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out1).To(Equal(""))

		out2, err := renderWithData(`{{ stringMethodOptionsExtension 99701 . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out2).To(Equal(""))
	})

	It("boolFieldExtension registers unknown fieldIDs and returns false", func() {
		f := &descriptorpb.FieldDescriptorProto{Options: &descriptorpb.FieldOptions{}}
		out, err := renderWithData(`{{ boolFieldExtension 99702 . }}`, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("false"))
	})

	It("int64MessageExtension registers unknown fieldIDs and returns 0", func() {
		m := &descriptorpb.DescriptorProto{Options: &descriptorpb.MessageOptions{}}
		out, err := renderWithData(`{{ int64MessageExtension 99703 . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("0"))
	})
})
