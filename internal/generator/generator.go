// Package generator renders arbitrary files from Go text/template sources
// driven by a protoc CodeGeneratorRequest.
package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Generate walks the files protoc asked to generate, applies the template set
// rooted at opts.TemplateDir once per service (or per file, per package or
// once, depending on the mode), and attaches the rendered output to the
// plugin response. Files that share a name have their content concatenated
// in arrival order, and a template that renders only whitespace is skipped.
func Generate(plugin *protogen.Plugin, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}

	req := plugin.Request

	reg := descriptor.NewRegistry()
	SetRegistry(reg)
	if err := reg.Load(req); err != nil {
		return fmt.Errorf("registry: failed to load the request: %w", err)
	}
	if err := LoadDescriptors(req.GetProtoFile()); err != nil {
		return err
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
			// A template that renders only whitespace opts out for this
			// file, e.g. one guarded by `{{if}}` to apply to a single proto
			// package; emitting it would clash with the file it is meant for.
			if strings.TrimSpace(tmpl.GetContent()) == "" {
				continue
			}
			emit(tmpl.GetName(), tmpl.GetContent())
		}
		return nil
	}

	// The request also carries every file the schema imports, which only
	// resolve types and options; protoc asked for the others.
	var files []*descriptorpb.FileDescriptorProto
	for _, f := range plugin.Files {
		if f.Generate {
			files = append(files, f.Proto)
		}
	}

	switch opts.Mode {
	case ModeOnce:
		if err := emitAll(NewGenericFilesTemplateBasedEncoder(opts.TemplateDir, "", files)); err != nil {
			return err
		}
	case ModePackage:
		groups := byPackage(files)
		for _, pkg := range packages(files) {
			if err := emitAll(NewGenericFilesTemplateBasedEncoder(opts.TemplateDir, pkg, groups[pkg])); err != nil {
				return err
			}
		}
	default:
		for _, file := range files {
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
