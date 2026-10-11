package generator

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// optionTypes resolves every extension declared in the request. A schema has
// to import the file that declares an option to use it, so the request always
// carries its definition, whether or not the plugin links its Go bindings.
var optionTypes = new(protoregistry.Types)

// decodedOptions caches each options message decoded against optionTypes, as
// templates tend to read the options of one descriptor several times.
var decodedOptions = map[proto.Message]protoreflect.Message{}

// LoadOptionTypes makes the extensions declared in files readable through the
// option, hasOption and options template helpers.
func LoadOptionTypes(files []*descriptorpb.FileDescriptorProto) error {
	reg, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: files})
	if err != nil {
		return fmt.Errorf("options: failed to load the descriptors: %w", err)
	}

	types := new(protoregistry.Types)
	reg.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		err = registerExtensions(types, fd.Extensions(), fd.Messages())
		return err == nil
	})
	if err != nil {
		return fmt.Errorf("options: failed to register the extensions: %w", err)
	}

	optionTypes = types
	decodedOptions = map[proto.Message]protoreflect.Message{}
	return nil
}

func registerExtensions(types *protoregistry.Types, exts protoreflect.ExtensionDescriptors, msgs protoreflect.MessageDescriptors) error {
	for i := range exts.Len() {
		if err := types.RegisterExtension(dynamicpb.NewExtensionType(exts.Get(i))); err != nil {
			return err
		}
	}
	for i := range msgs.Len() {
		md := msgs.Get(i)
		if err := registerExtensions(types, md.Extensions(), md.Messages()); err != nil {
			return err
		}
	}
	return nil
}

// getOption returns the value of the custom option name, e.g.
// "protoc.contrib.defaults.value", set on desc, or nil when it isn't set.
// Scalars keep their Go type, enums become their value name, messages become
// a map keyed by field name, and repeated options become a slice.
func getOption(name string, desc any) (any, error) {
	opts, err := decodeOptions(desc)
	if err != nil || opts == nil {
		return nil, err
	}

	name = strings.TrimPrefix(name, ".")
	var value any
	opts.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.IsExtension() && string(fd.FullName()) == name {
			value = optionValue(fd, v)
			return false
		}
		return true
	})
	return value, nil
}

// hasOption reports whether the custom option name is set on desc, which
// getOption can't tell apart from an option set to its zero value.
func hasOption(name string, desc any) (bool, error) {
	value, err := getOption(name, desc)
	return value != nil, err
}

// listOptions returns every custom option set on desc, keyed by full name.
func listOptions(desc any) (map[string]any, error) {
	opts, err := decodeOptions(desc)
	if err != nil {
		return nil, err
	}

	values := map[string]any{}
	if opts == nil {
		return values, nil
	}
	opts.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.IsExtension() {
			values[string(fd.FullName())] = optionValue(fd, v)
		}
		return true
	})
	return values, nil
}

// decodeOptions returns the options of desc with every custom option decoded.
// Options arrive as unknown fields unless the plugin links their Go bindings,
// so they are re-encoded and decoded against the request's own definitions.
func decodeOptions(desc any) (protoreflect.Message, error) {
	opts, err := optionsOf(desc)
	if err != nil || opts == nil || !opts.ProtoReflect().IsValid() {
		return nil, err
	}
	if decoded, ok := decodedOptions[opts]; ok {
		return decoded, nil
	}

	b, err := proto.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("options: failed to encode %T: %w", opts, err)
	}
	decoded := opts.ProtoReflect().Type().New()
	if err := (proto.UnmarshalOptions{Resolver: optionTypes}).Unmarshal(b, decoded.Interface()); err != nil {
		return nil, fmt.Errorf("options: failed to decode %T: %w", opts, err)
	}

	decodedOptions[opts] = decoded
	return decoded, nil
}

// optionsOf returns the options message of a descriptor: any descriptorpb
// type, or anything embedding one, such as the grpc-gateway registry types
// getMessageType returns.
func optionsOf(desc any) (proto.Message, error) {
	if desc == nil {
		return nil, nil
	}
	method := reflect.ValueOf(desc).MethodByName("GetOptions")
	if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
		if opts, ok := method.Call(nil)[0].Interface().(proto.Message); ok {
			return opts, nil
		}
	}
	return nil, fmt.Errorf("options: %T is not a descriptor", desc)
}

func optionValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch {
	case fd.IsList():
		list := v.List()
		values := make([]any, list.Len())
		for i := range list.Len() {
			values[i] = singularValue(fd, list.Get(i))
		}
		return values
	case fd.IsMap():
		values := map[string]any{}
		v.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
			values[k.String()] = singularValue(fd.MapValue(), v)
			return true
		})
		return values
	}
	return singularValue(fd, v)
}

func singularValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name())
		}
		return int32(v.Enum())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		values := map[string]any{}
		v.Message().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			key := string(fd.Name())
			if fd.IsExtension() {
				key = string(fd.FullName())
			}
			values[key] = optionValue(fd, v)
			return true
		})
		return values
	}
	return v.Interface()
}
