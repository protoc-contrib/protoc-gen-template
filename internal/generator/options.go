package generator

import "fmt"

// Mode controls which proto files are processed and at what granularity.
type Mode string

const (
	// ModeService (default) renders the template set once per service, only
	// for files that declare at least one service.
	ModeService Mode = "service"
	// ModeFile renders the template set once per file that declares at least
	// one service. Templates see .File and .File.Service; .Service is nil.
	ModeFile Mode = "file"
	// ModeAll renders the template set once per file to generate, including
	// files that declare no services. Templates see .File; .Service is nil.
	ModeAll Mode = "all"
	// ModePackage renders the template set once per proto package among the
	// files to generate. Templates see .Package and the package's .Files;
	// .File and .Service are nil.
	ModePackage Mode = "package"
	// ModeOnce renders the template set once per run. Templates see every
	// file to generate in .Files; .Package is empty, .File and .Service nil.
	ModeOnce Mode = "once"
)

// Options controls how the plugin discovers templates and emits output.
type Options struct {
	// TemplateDir is the root directory containing .tmpl files to render.
	TemplateDir string
	// Mode selects the iteration granularity: "service" (default), "file",
	// "all", "package", or "once".
	Mode Mode
}

// Set applies a single `name=value` plugin parameter to the options. The
// signature matches what protogen.Options.ParamFunc expects.
func (o *Options) Set(name, value string) error {
	switch name {
	case "template_dir":
		o.TemplateDir = value
	case "mode":
		switch Mode(value) {
		case ModeService, ModeFile, ModeAll, ModePackage, ModeOnce:
			o.Mode = Mode(value)
		default:
			return fmt.Errorf("unknown mode %q: must be service, file, all, package, or once", value)
		}
	default:
		return fmt.Errorf("unknown plugin option %q", name)
	}
	return nil
}
