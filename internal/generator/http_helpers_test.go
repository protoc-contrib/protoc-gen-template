package generator_test

import (
	"github.com/protoc-contrib/protoc-gen-template/internal/generator"
	"google.golang.org/genproto/googleapis/api/annotations"
	ggdescriptor "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newMethod builds a MethodDescriptorProto with the given HttpRule attached
// via the google.api.http extension. Pass nil to omit the extension.
func newMethod(name string, rule *annotations.HttpRule) *descriptorpb.MethodDescriptorProto {
	m := &descriptorpb.MethodDescriptorProto{
		Name:       proto.String(name),
		InputType:  proto.String(".demo.Req"),
		OutputType: proto.String(".demo.Resp"),
	}
	if rule != nil {
		m.Options = &descriptorpb.MethodOptions{}
		proto.SetExtension(m.Options, annotations.E_Http, rule)
	}
	return m
}

var _ = Describe("http helpers", func() {
	DescribeTable("httpVerb + httpPath",
		func(rule *annotations.HttpRule, wantVerb, wantPath string) {
			m := newMethod("M", rule)
			verb, err := renderWithData(`{{ httpVerb . }}`, m)
			Expect(err).NotTo(HaveOccurred())
			Expect(verb).To(Equal(wantVerb))

			path, err := renderWithData(`{{ httpPath . }}`, m)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(wantPath))
		},
		Entry("GET",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "/v1/foo"}},
			"GET", "/v1/foo"),
		Entry("POST",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Post{Post: "/v1/foo"}},
			"POST", "/v1/foo"),
		Entry("PUT",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Put{Put: "/v1/foo"}},
			"PUT", "/v1/foo"),
		Entry("DELETE",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Delete{Delete: "/v1/foo"}},
			"DELETE", "/v1/foo"),
		Entry("PATCH",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Patch{Patch: "/v1/foo"}},
			"PATCH", "/v1/foo"),
		Entry("Custom",
			&annotations.HttpRule{Pattern: &annotations.HttpRule_Custom{
				Custom: &annotations.CustomHttpPattern{Kind: "HEAD", Path: "/v1/head"},
			}},
			"HEAD", "/v1/head"),
	)

	It("returns empty verb/path when options are absent", func() {
		m := newMethod("M", nil)
		verb, err := renderWithData(`{{ httpVerb . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(verb).To(Equal(""))

		path, err := renderWithData(`{{ httpPath . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(""))
	})

	It("returns empty verb/path when the pattern is unset", func() {
		m := newMethod("M", &annotations.HttpRule{})
		verb, err := renderWithData(`{{ httpVerb . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(verb).To(Equal(""))

		path, err := renderWithData(`{{ httpPath . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(""))
	})

	It("renders httpBody", func() {
		m := newMethod("M", &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Post{Post: "/v1/foo"},
			Body:    "*",
		})
		out, err := renderWithData(`{{ httpBody . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("*"))
	})

	It("returns empty body when no HttpRule is attached", func() {
		m := newMethod("M", nil)
		out, err := renderWithData(`{{ httpBody . }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(""))
	})

	It("collects httpPathsAdditionalBindings across all verbs", func() {
		rule := &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Get{Get: "/primary"},
			AdditionalBindings: []*annotations.HttpRule{
				{Pattern: &annotations.HttpRule_Get{Get: "/a"}},
				{Pattern: &annotations.HttpRule_Post{Post: "/b"}},
				{Pattern: &annotations.HttpRule_Put{Put: "/c"}},
				{Pattern: &annotations.HttpRule_Delete{Delete: "/d"}},
				{Pattern: &annotations.HttpRule_Patch{Patch: "/e"}},
				{Pattern: &annotations.HttpRule_Custom{
					Custom: &annotations.CustomHttpPattern{Kind: "HEAD", Path: "/f"},
				}},
			},
		}
		m := newMethod("M", rule)
		out, err := renderWithData(`{{ range httpPathsAdditionalBindings . }}{{ . }};{{ end }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("/a;/b;/c;/d;/e;/f;"))
	})

	It("returns nil for httpPathsAdditionalBindings when no rule is attached", func() {
		m := newMethod("M", nil)
		out, err := renderWithData(`{{ len (httpPathsAdditionalBindings .) }}`, m)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal("0"))
	})

	Describe("urlHasVarsFromMessage", func() {
		mkMsg := func(fields ...*descriptorpb.FieldDescriptorProto) *ggdescriptor.Message {
			return &ggdescriptor.Message{
				DescriptorProto: &descriptorpb.DescriptorProto{
					Name:  proto.String("M"),
					Field: fields,
				},
			}
		}

		// scalarWithJSON builds a non-message field with explicit name + json_name.
		scalarWithJSON := func(name, jsonName string) *descriptorpb.FieldDescriptorProto {
			return &descriptorpb.FieldDescriptorProto{
				Name:     proto.String(name),
				JsonName: proto.String(jsonName),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			}
		}

		It("matches the field name", func() {
			msg := mkMsg(scalarWithJSON("user_id", "userId"))
			out, err := renderWithData(
				`{{ urlHasVarsFromMessage (index . 0) (index . 1) }}`,
				[]any{"/v1/users/{user_id}", msg},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("true"))
		})

		It("matches the json_name as fallback", func() {
			msg := mkMsg(scalarWithJSON("user_id", "userId"))
			out, err := renderWithData(
				`{{ urlHasVarsFromMessage (index . 0) (index . 1) }}`,
				[]any{"/v1/users/{userId}", msg},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("true"))
		})

		It("skips message-typed fields", func() {
			f := &descriptorpb.FieldDescriptorProto{
				Name:     proto.String("nested"),
				JsonName: proto.String("nested"),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				TypeName: proto.String(".demo.Nested"),
			}
			msg := mkMsg(f)
			out, err := renderWithData(
				`{{ urlHasVarsFromMessage (index . 0) (index . 1) }}`,
				[]any{"/v1/{nested}", msg},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})

		It("returns false when no variables are present", func() {
			msg := mkMsg(scalarWithJSON("user_id", "userId"))
			out, err := renderWithData(
				`{{ urlHasVarsFromMessage (index . 0) (index . 1) }}`,
				[]any{"/v1/users", msg},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("false"))
		})
	})

	It("exports urlHasVarsFromMessage, httpPath, httpVerb, httpBody, httpPathsAdditionalBindings in the funcmap", func() {
		for _, k := range []string{"httpPath", "httpVerb", "httpBody", "httpPathsAdditionalBindings", "urlHasVarsFromMessage"} {
			Expect(generator.ProtoHelpersFuncMap).To(HaveKey(k))
		}
	})
})
