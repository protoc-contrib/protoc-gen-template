package generator

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// extensionByNumber returns the value of the custom option with the given
// field number set on desc, or nil when it isn't set. The number-based
// helpers below predate option, which reads options by full name; they read
// options the same way, against the request's own declarations.
func extensionByNumber(fieldID int32, desc any) any {
	opts, err := decodeOptions(desc)
	if err != nil || opts == nil {
		return nil
	}

	var value any
	opts.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.IsExtension() && int32(fd.Number()) == fieldID {
			value = v.Interface()
			return false
		}
		return true
	})
	return value
}

func stringExtension(fieldID int32, desc any) string {
	s, _ := extensionByNumber(fieldID, desc).(string)
	return s
}

func boolExtension(fieldID int32, desc any) bool {
	b, _ := extensionByNumber(fieldID, desc).(bool)
	return b
}

func int64Extension(fieldID int32, desc any) int64 {
	switch i := extensionByNumber(fieldID, desc).(type) {
	case int32:
		return int64(i)
	case int64:
		return i
	case uint32:
		return int64(i)
	case uint64:
		return int64(i)
	}
	return 0
}

// stringMethodOptionsExtension extracts method options of a string type.
// To define your own extensions see:
// https://developers.google.com/protocol-buffers/docs/proto#customoptions
// Typically the fieldID of private extensions should be in the range:
// 50000-99999
func stringMethodOptionsExtension(fieldID int32, f *descriptorpb.MethodDescriptorProto) string {
	return stringExtension(fieldID, f)
}

// boolMethodOptionsExtension extracts method options of a boolean type.
func boolMethodOptionsExtension(fieldID int32, f *descriptorpb.MethodDescriptorProto) bool {
	return boolExtension(fieldID, f)
}

// stringFileOptionsExtension extracts file options of a string type.
// To define your own extensions see:
// https://developers.google.com/protocol-buffers/docs/proto#customoptions
// Typically the fieldID of private extensions should be in the range:
// 50000-99999
func stringFileOptionsExtension(fieldID int32, f *descriptorpb.FileDescriptorProto) string {
	return stringExtension(fieldID, f)
}

func stringFieldExtension(fieldID int32, f *descriptorpb.FieldDescriptorProto) string {
	return stringExtension(fieldID, f)
}

func int64FieldExtension(fieldID int32, f *descriptorpb.FieldDescriptorProto) int64 {
	return int64Extension(fieldID, f)
}

func int64MessageExtension(fieldID int32, f *descriptorpb.DescriptorProto) int64 {
	return int64Extension(fieldID, f)
}

func stringMessageExtension(fieldID int32, f *descriptorpb.DescriptorProto) string {
	return stringExtension(fieldID, f)
}

func boolFieldExtension(fieldID int32, f *descriptorpb.FieldDescriptorProto) bool {
	return boolExtension(fieldID, f)
}

func boolMessageExtension(fieldID int32, f *descriptorpb.DescriptorProto) bool {
	return boolExtension(fieldID, f)
}
