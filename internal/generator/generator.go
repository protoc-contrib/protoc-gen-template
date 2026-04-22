// Package generator renders arbitrary files from Go text/template sources
// driven by a protoc CodeGeneratorRequest.
package generator

import (
	"fmt"
	"sort"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"google.golang.org/protobuf/compiler/protogen"
)

// Generate walks the request the plugin was invoked with, applies the
// template set rooted at opts.TemplateDir once per service (or per file,
// depending on the mode flags), and attaches the rendered output to the
// plugin response. Files that share a name have their content concatenated
// in arrival order.
func Generate(plugin *protogen.Plugin, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}

	req := plugin.Request

	if opts.Registry {
		reg := descriptor.NewRegistry()
		SetRegistry(reg)
		if err := reg.Load(req); err != nil {
			return fmt.Errorf("registry: failed to load the request: %w", err)
		}
	}

	emitted := map[string]*generatedFile{}
	emit := func(name, content string) {
		if gf, ok := emitted[name]; ok {
			gf.content += content
			return
		}
		emitted[name] = &generatedFile{name: name, content: content}
	}

	emitAll := func(enc *GenericTemplateBasedEncoder) error {
		tmpls, err := enc.Files()
		if err != nil {
			return err
		}
		for _, tmpl := range tmpls {
			emit(tmpl.GetName(), tmpl.GetContent())
		}
		return nil
	}

	for _, file := range req.GetProtoFile() {
		switch opts.Mode {
		case ModeAll:
			if err := emitAll(NewGenericTemplateBasedEncoder(opts.TemplateDir, file)); err != nil {
				return err
			}
		case ModeFile:
			if len(file.GetService()) == 0 {
				continue
			}
			if err := emitAll(NewGenericTemplateBasedEncoder(opts.TemplateDir, file)); err != nil {
				return err
			}
		default: // ModeService or ""
			for _, service := range file.GetService() {
				if err := emitAll(NewGenericServiceTemplateBasedEncoder(opts.TemplateDir, service, file)); err != nil {
					return err
				}
			}
		}
	}

	names := make([]string, 0, len(emitted))
	for name := range emitted {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		gf := emitted[name]
		out := plugin.NewGeneratedFile(gf.name, "")
		if _, err := out.Write([]byte(gf.content)); err != nil {
			return fmt.Errorf("write %q: %w", gf.name, err)
		}
	}
	return nil
}

type generatedFile struct {
	name    string
	content string
}
