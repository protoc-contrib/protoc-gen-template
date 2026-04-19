package generator_test

import (
	"github.com/protoc-contrib/protoc-gen-template/internal/generator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("misc helpers", func() {
	Describe("goPkg / goPkgLastElement", func() {
		It("returns the raw go_package", func() {
			f := &descriptorpb.FileDescriptorProto{
				Options: &descriptorpb.FileOptions{GoPackage: proto.String("example.com/demo/pkg")},
			}
			out, err := renderWithData(`{{ goPkg . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("example.com/demo/pkg"))
		})

		It("returns only the last path element", func() {
			f := &descriptorpb.FileDescriptorProto{
				Options: &descriptorpb.FileOptions{GoPackage: proto.String("example.com/demo/pkg")},
			}
			out, err := renderWithData(`{{ goPkgLastElement . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("pkg"))
		})
	})

	Describe("isFieldMessageTimeStamp", func() {
		It("is true for .google.protobuf.Timestamp", func() {
			f := messageField("t", ".google.protobuf.Timestamp", false)
			out, err := renderWithData(`{{ isFieldMessageTimeStamp . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("true"))
		})

		It("is false for other messages", func() {
			f := messageField("b", ".demo.Bar", false)
			out, err := renderWithData(`{{ isFieldMessageTimeStamp . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})

		It("is false for scalars", func() {
			f := scalarField("x", descriptorpb.FieldDescriptorProto_TYPE_STRING, false)
			out, err := renderWithData(`{{ isFieldMessageTimeStamp . }}`, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})
	})

	Describe("getEnumValue", func() {
		It("returns the values of a matching enum (case-insensitive)", func() {
			enums := []*descriptorpb.EnumDescriptorProto{
				{
					Name: proto.String("Color"),
					Value: []*descriptorpb.EnumValueDescriptorProto{
						{Name: proto.String("RED"), Number: proto.Int32(0)},
						{Name: proto.String("GREEN"), Number: proto.Int32(1)},
					},
				},
			}
			out, err := renderWithData(
				`{{ range getEnumValue (index . 0) (index . 1) }}{{ .Name }}={{ .Number }};{{ end }}`,
				[]any{enums, "color"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("RED=0;GREEN=1;"))
		})

		It("returns nil when no enum matches", func() {
			out, err := renderWithData(
				`{{ len (getEnumValue (index . 0) (index . 1)) }}`,
				[]any{[]*descriptorpb.EnumDescriptorProto{}, "missing"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("0"))
		})
	})

	Describe("isFieldMap / fieldMapKeyType / fieldMapValueType", func() {
		// A proto map<K,V> is represented as a repeated message field pointing
		// at a synthesized nested type with fields (key=1, value=2).
		mapMessage := func(keyType, valueType descriptorpb.FieldDescriptorProto_Type) *descriptorpb.DescriptorProto {
			return &descriptorpb.DescriptorProto{
				Name: proto.String("Outer"),
				NestedType: []*descriptorpb.DescriptorProto{{
					Name: proto.String("TagsEntry"),
					Field: []*descriptorpb.FieldDescriptorProto{
						{
							Name:   proto.String("key"),
							Number: proto.Int32(1),
							Type:   keyType.Enum(),
							Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						},
						{
							Name:   proto.String("value"),
							Number: proto.Int32(2),
							Type:   valueType.Enum(),
							Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						},
					},
				}},
			}
		}

		mapField := func() *descriptorpb.FieldDescriptorProto {
			return &descriptorpb.FieldDescriptorProto{
				Name:     proto.String("tags"),
				TypeName: proto.String(".demo.Outer.TagsEntry"),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			}
		}

		It("identifies a proto map field", func() {
			f := mapField()
			m := mapMessage(
				descriptorpb.FieldDescriptorProto_TYPE_STRING,
				descriptorpb.FieldDescriptorProto_TYPE_INT32,
			)
			out, err := renderWithData(
				`{{ isFieldMap (index . 0) (index . 1) }}`,
				[]any{f, m},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("true"))
		})

		It("returns false when no matching nested type exists", func() {
			f := mapField()
			m := &descriptorpb.DescriptorProto{Name: proto.String("Outer")}
			out, err := renderWithData(
				`{{ isFieldMap (index . 0) (index . 1) }}`,
				[]any{f, m},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})

		It("returns false when TypeName is nil", func() {
			f := &descriptorpb.FieldDescriptorProto{
				Name: proto.String("x"),
				Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			}
			m := &descriptorpb.DescriptorProto{Name: proto.String("Outer")}
			out, err := renderWithData(
				`{{ isFieldMap (index . 0) (index . 1) }}`,
				[]any{f, m},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})

		It("returns the key and value field types", func() {
			f := mapField()
			m := mapMessage(
				descriptorpb.FieldDescriptorProto_TYPE_STRING,
				descriptorpb.FieldDescriptorProto_TYPE_INT32,
			)
			out, err := renderWithData(
					`{{ goType "" (fieldMapKeyType (index . 0) (index . 1)) }}/{{ goType "" (fieldMapValueType (index . 0) (index . 1)) }}`,
				[]any{f, m},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("string/int32"))
		})

		It("fieldMapKeyType/fieldMapValueType return nil when nested type is missing", func() {
			f := mapField()
			m := &descriptorpb.DescriptorProto{Name: proto.String("Outer")}
			out, err := renderWithData(
				`{{ if fieldMapKeyType (index . 0) (index . 1) }}has{{ else }}nil{{ end }}`,
				[]any{f, m},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("nil"))
		})
	})

	Describe("setStore / getStore", func() {
		It("round-trips a value through the global store", func() {
			// setStore returns "" so it's usable inline.
			out, err := renderWithData(
				`{{ setStore "k1" "hello" }}{{ getStore "k1" }}`,
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("hello"))
		})

		It("returns false for a missing key", func() {
			out, err := renderWithData(`{{ getStore "never-set" }}`, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})
	})

	Describe("InitPathMap + comments", func() {
		// Build a FileDescriptorProto that has SourceCodeInfo with locations
		// covering the first message. The protoc path convention is
		// file.message_type=4, so the first top-level message lives at [4, 0].
		newFile := func() *descriptorpb.FileDescriptorProto {
			msg := &descriptorpb.DescriptorProto{Name: proto.String("Msg")}
			return &descriptorpb.FileDescriptorProto{
				Name:        proto.String("demo.proto"),
				Package:     proto.String("demo"),
				MessageType: []*descriptorpb.DescriptorProto{msg},
				SourceCodeInfo: &descriptorpb.SourceCodeInfo{
					Location: []*descriptorpb.SourceCodeInfo_Location{
						{
							Path:                    []int32{4, 0},
							LeadingComments:         proto.String(" leading "),
							TrailingComments:        proto.String(" trailing "),
							LeadingDetachedComments: []string{" detached1 ", " detached2 "},
						},
					},
				},
			}
		}

		It("resolves leading/trailing/detached comments by descriptor identity", func() {
			file := newFile()
			generator.InitPathMap(file)

			msg := file.MessageType[0]
			out, err := renderWithData(
				`{{ leadingComment . }}|{{ trailingComment . }}|{{ index (leadingDetachedComments .) 0 }}|{{ index (leadingDetachedComments .) 1 }}`,
				msg,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(" leading | trailing | detached1 | detached2 "))
		})

		It("returns empty strings for descriptors with no recorded location", func() {
			file := newFile()
			generator.InitPathMap(file)

			unrelated := &descriptorpb.DescriptorProto{Name: proto.String("Other")}
			out, err := renderWithData(
				`{{ leadingComment . }}|{{ trailingComment . }}|{{ len (leadingDetachedComments .) }}`,
				unrelated,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("||0"))
		})
	})
})
