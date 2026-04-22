package generator

import (
	"fmt"
	"strconv"
)

// Mode controls which proto files are processed and at what granularity.
type Mode string

const (
	// ModeService (default) renders the template set once per service, only
	// for files that declare at least one service.
	ModeService Mode = "service"
	// ModeFile renders the template set once per file that declares at least
	// one service. Templates see .File and .File.Service; .Service is nil.
	ModeFile Mode = "file"
	// ModeAll renders the template set once per file, including files that
	// declare no services. Templates see .File; .Service is nil.
	ModeAll Mode = "all"
)

// Options controls how the plugin discovers templates and emits output.
type Options struct {
	// TemplateDir is the root directory containing .tmpl files to render.
	TemplateDir string
	// Mode selects the iteration granularity: "service" (default), "file", or "all".
	Mode Mode
	// Registry loads the full request into a grpc-gateway registry so templates
	// can walk cross-file message references.
	Registry bool
}

// Set applies a single `name=value` plugin parameter to the options. The
// signature matches what protogen.Options.ParamFunc expects.
func (o *Options) Set(name, value string) error {
	switch name {
	case "template_dir":
		o.TemplateDir = value
	case "registry":
		return setBool(&o.Registry, name, value)
	case "mode":
		switch Mode(value) {
		case ModeService, ModeFile, ModeAll:
			o.Mode = Mode(value)
		default:
			return fmt.Errorf("unknown mode %q: must be service, file, or all", value)
		}
	default:
		return fmt.Errorf("unknown plugin option %q", name)
	}
	return nil
}

func setBool(dst *bool, name, value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid value for %q: %w", name, err)
	}
	*dst = v
	return nil
}
