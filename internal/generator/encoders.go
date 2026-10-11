package generator

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
	plugin_go "google.golang.org/protobuf/types/pluginpb"
)

type GenericTemplateBasedEncoder struct {
	templateDir string
	service     *descriptor.ServiceDescriptorProto
	file        *descriptor.FileDescriptorProto
	enum        []*descriptor.EnumDescriptorProto
	pkg         string
	files       []*descriptor.FileDescriptorProto
}

type Ast struct {
	BuildDate      time.Time                          `json:"build-date"`
	BuildHostname  string                             `json:"build-hostname"`
	BuildUser      string                             `json:"build-user"`
	PWD            string                             `json:"pwd"`
	File           *descriptor.FileDescriptorProto    `json:"file"`
	Package        string                             `json:"package"`
	Files          []*descriptor.FileDescriptorProto  `json:"files"`
	RawFilename    string                             `json:"raw-filename"`
	Filename       string                             `json:"filename"`
	TemplateDir    string                             `json:"template-dir"`
	Service        *descriptor.ServiceDescriptorProto `json:"service"`
	Enum           []*descriptor.EnumDescriptorProto  `json:"enum"`
}

func NewGenericServiceTemplateBasedEncoder(templateDir string, service *descriptor.ServiceDescriptorProto, file *descriptor.FileDescriptorProto) (e *GenericTemplateBasedEncoder) {
	e = &GenericTemplateBasedEncoder{
		service:     service,
		file:        file,
		templateDir: templateDir,
		enum:        file.GetEnumType(),
		pkg:         file.GetPackage(),
		files:       []*descriptor.FileDescriptorProto{file},
	}
	slog.Debug("new encoder", "file", file.GetName(), "service", service.GetName(), "template-dir", templateDir)
	InitPathMap(file)

	return
}

func NewGenericTemplateBasedEncoder(templateDir string, file *descriptor.FileDescriptorProto) (e *GenericTemplateBasedEncoder) {
	e = &GenericTemplateBasedEncoder{
		service:     nil,
		file:        file,
		templateDir: templateDir,
		enum:        file.GetEnumType(),
		pkg:         file.GetPackage(),
		files:       []*descriptor.FileDescriptorProto{file},
	}
	slog.Debug("new encoder", "file", file.GetName(), "template-dir", templateDir)
	InitPathMap(file)

	return
}

// NewGenericFilesTemplateBasedEncoder renders templates once for several
// files: those of the proto package pkg, or every file to generate when pkg
// is empty. Templates see .Files and .Package; .File and .Service are nil.
func NewGenericFilesTemplateBasedEncoder(templateDir string, pkg string, files []*descriptor.FileDescriptorProto) (e *GenericTemplateBasedEncoder) {
	e = &GenericTemplateBasedEncoder{
		templateDir: templateDir,
		pkg:         pkg,
		files:       files,
	}
	slog.Debug("new encoder", "package", pkg, "files", len(files), "template-dir", templateDir)
	InitPathMap(files...)

	return
}

// templates returns the paths, relative to the template directory, of the
// templates to render, .tmpl, and of the partials every template can call
// into, .tpl.
func (e *GenericTemplateBasedEncoder) templates() (templates, partials []string, err error) {
	err = filepath.Walk(e.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".tmpl" && ext != ".tpl" {
			return nil
		}
		rel, err := filepath.Rel(e.templateDir, path)
		if err != nil {
			return err
		}
		if ext == ".tpl" {
			slog.Debug("found partial", "path", rel)
			partials = append(partials, rel)
			return nil
		}
		slog.Debug("found template", "path", rel)
		templates = append(templates, rel)
		return nil
	})
	return templates, partials, err
}

func (e *GenericTemplateBasedEncoder) genAst(templateFilename string) (*Ast, error) {
	// prepare the ast passed to the template engine
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ast := Ast{
		BuildDate:      time.Now(),
		BuildHostname:  hostname,
		BuildUser:      os.Getenv("USER"),
		PWD:            pwd,
		File:           e.file,
		Package:        e.pkg,
		Files:          e.files,
		TemplateDir:    e.templateDir,
		RawFilename:    templateFilename,
		Filename:       "",
		Service:        e.service,
		Enum:           e.enum,
	}
	buffer := new(bytes.Buffer)

	unescaped, err := url.QueryUnescape(templateFilename)
	if err != nil {
		slog.Warn("failed to unescape filepath", "path", templateFilename, "err", err)
	} else {
		templateFilename = unescaped
	}

	tmpl, err := template.New("").Funcs(ProtoHelpersFuncMap).Parse(templateFilename)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buffer, ast); err != nil {
		return nil, err
	}
	ast.Filename = buffer.String()
	return &ast, nil
}

func (e *GenericTemplateBasedEncoder) buildContent(templateFilename string, partials []string) (string, string, error) {
	// initialize template engine
	fullPath := filepath.Join(e.templateDir, templateFilename)
	paths := []string{fullPath}
	for _, partial := range partials {
		paths = append(paths, filepath.Join(e.templateDir, partial))
	}
	tmpl := template.New(filepath.Base(fullPath)).Funcs(ProtoHelpersFuncMap)
	tmpl, err := tmpl.Funcs(template.FuncMap{"include": include(tmpl)}).ParseFiles(paths...)
	if err != nil {
		return "", "", err
	}

	ast, err := e.genAst(templateFilename)
	if err != nil {
		return "", "", err
	}

	// generate the content
	buffer := new(bytes.Buffer)
	if err := tmpl.Execute(buffer, ast); err != nil {
		return "", "", err
	}

	return buffer.String(), ast.Filename, nil
}

func (e *GenericTemplateBasedEncoder) Files() ([]*plugin_go.CodeGeneratorResponse_File, error) {
	templates, partials, err := e.templates()
	if err != nil {
		return nil, fmt.Errorf("walk templates in %q: %w", e.templateDir, err)
	}

	files := make([]*plugin_go.CodeGeneratorResponse_File, 0, len(templates))
	for _, templateFilename := range templates {
		content, translatedFilename, err := e.buildContent(templateFilename, partials)
		if err != nil {
			return nil, fmt.Errorf("render %q: %w", templateFilename, err)
		}
		filename := translatedFilename[:len(translatedFilename)-len(".tmpl")]

		files = append(files, &plugin_go.CodeGeneratorResponse_File{
			Content: &content,
			Name:    &filename,
		})
	}
	return files, nil
}

// maxIncludeDepth bounds how deep include calls the same template, to fail
// cleanly on a recursive walk that never ends, e.g. over a message that
// contains itself.
const maxIncludeDepth = 1000

// include returns the include template function for tmpl: it renders the
// named template with data and returns the output, which {{template}} can't,
// so the output can be piped and tested.
func include(tmpl *template.Template) func(string, any) (string, error) {
	depth := map[string]int{}
	return func(name string, data any) (string, error) {
		if depth[name] >= maxIncludeDepth {
			return "", &includeDepthError{name: name}
		}
		depth[name]++
		defer func() { depth[name]-- }()

		var buffer strings.Builder
		if err := tmpl.ExecuteTemplate(&buffer, name, data); err != nil {
			// Pass the depth error up as is, rather than wrapped once per
			// call it went through.
			var depthErr *includeDepthError
			if errors.As(err, &depthErr) {
				return "", depthErr
			}
			return "", err
		}
		return buffer.String(), nil
	}
}

type includeDepthError struct {
	name string
}

func (e *includeDepthError) Error() string {
	return fmt.Sprintf("include: %q called itself more than %d times", e.name, maxIncludeDepth)
}
