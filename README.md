# protoc-gen-template

[![CI](https://github.com/protoc-contrib/protoc-gen-template/actions/workflows/ci.yml/badge.svg)](https://github.com/protoc-contrib/protoc-gen-template/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/protoc-contrib/protoc-gen-template?include_prereleases)](https://github.com/protoc-contrib/protoc-gen-template/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![protoc](https://img.shields.io/badge/protoc-compatible-blue)](https://protobuf.dev)

A [protoc](https://protobuf.dev) plugin that renders arbitrary files from
Go [`text/template`](https://pkg.go.dev/text/template) sources driven by
the parsed proto AST. The plugin walks a user-provided template directory,
parses every `.tmpl` file, and writes the rendered output into the protoc
response. The template engine is extended with helpers from
[Masterminds/sprig](https://github.com/Masterminds/sprig) plus a
proto-aware funcmap (naming, arithmetic, field/message walkers, HTTP
annotation accessors, extension readers).

This project is a continuation of
[moul/protoc-gen-gotemplate](https://github.com/moul/protoc-gen-gotemplate)
by Manfred Touron and contributors, republished under `protoc-contrib` and
renamed so the name reflects what the plugin actually does — drive code
generation from templates, regardless of the output language. The binary
entrypoint now uses `google.golang.org/protobuf/compiler/protogen`, the
layout mirrors the rest of the `protoc-contrib` plugins, and the build and
release pipeline runs on Nix + `release-please`.

## Features

- **Template-driven output** — any `.tmpl` file under `template_dir` is
  rendered; output filenames can themselves be templated, so one template
  can fan out to many files.
- **Conditional output** — a template that renders only whitespace is
  skipped, so it can opt out for a given proto file with `{{if}}`. When one
  `template_dir` holds templates for several proto packages (e.g. `v1/` and
  `v2/`), guard each template so it renders only for the proto file in its
  own directory; otherwise every template renders for every file:

  ```
  {{- if hasPrefix (printf "%s/" (dir .File.Name)) .RawFilename -}}
  ...
  {{- end }}
  ```
- **Reusable helpers** — `include` renders a named template and returns its
  output, so it can be piped, tested and called recursively, and `.tpl`
  partials share `define`s across every template. See
  [Partials and include](#partials-and-include).
- **Sprig funcmap** — every helper from
  [Masterminds/sprig](https://github.com/Masterminds/sprig) is available
  (date, string, crypto, flow, default, dictionary, and more).
- **Proto-aware funcmap** — naming (`camelCase`, `lowerCamelCase`,
  `snakeCase`, `kebabCase`, `goNormalize`, `shortType`, ...), arithmetic
  (`add`, `subtract`, `multiply`, `divide`), field/message walkers
  (`getMessageType`, `isFieldMessage`, `fieldMapKeyType`, ...), HTTP
  annotation accessors (`httpPath`, `httpVerb`, `httpBody`, ...), and
  custom option readers (`option`, `hasOption`, `options`).
- **Custom options, whoever owns them** — templates read any custom option
  set in the schema by its full name, from
  [protoc-types](https://github.com/protoc-contrib/protoc-types),
  googleapis, or a private package the plugin knows nothing about. See
  [Custom options](#custom-options).
- **Flexible iteration** — `mode=service` (default) renders once per gRPC
  service; `mode=file` renders once per file that declares a service;
  `mode=all` renders every proto file whether or not it declares a service.
  Every mode renders only the files protoc was asked to generate, not the
  files they import.
- **Cross-file lookups** — the plugin always loads a `grpc-gateway` registry
  so templates can resolve messages across imports via `getMessageType` and
  `getProtoFile`.
- **Structured logging** — the plugin emits `slog.Debug` events; enable
  them by configuring your slog handler's minimum level (e.g. `GOTOOLCHAIN` or
  a custom handler).

## Options

Pass plugin options via `--template_opt=key=value` (protoc) or the
`opt:` list under the plugin entry in `buf.gen.yaml`.

| Option         | Default      | Effect                                                                                                                           |
| -------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| `template_dir` | `./template` | Root directory containing `.tmpl` files.                                                                                         |
| `mode`         | `service`    | Iteration granularity: `service` (once per service), `file` (once per file that has a service), `all` (every file, no filter).   |

## Installation

```bash
go install github.com/protoc-contrib/protoc-gen-template/cmd/protoc-gen-template@latest
```

## Usage

### With buf

Add the plugin to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-template
    out: .
    opt:
      - template_dir=./templates
```

Then run:

```bash
buf generate
```

### With protoc

```bash
protoc \
  --template_out=. \
  --template_opt=template_dir=./templates \
  -I proto/ \
  proto/example.proto
```

## Example

Given this layout:

```
input.proto
templates/doc.txt.tmpl
templates/config.json.tmpl
```

and a `doc.txt.tmpl` like:

```
{{.File.Package}} — {{len .File.MessageType}} message(s)
{{range .File.MessageType}}- {{.Name}}
{{end}}
```

running:

```bash
protoc --template_out=. input.proto
```

produces `doc.txt` and `config.json` alongside the originals, one pair per
service declared in `input.proto`.

The top-level template context exposes the raw descriptor AST plus
per-invocation metadata:

```go
type Ast struct {
    File          *descriptorpb.FileDescriptorProto
    Package       string
    Service       *descriptorpb.ServiceDescriptorProto
    Enum          []*descriptorpb.EnumDescriptorProto
    TemplateDir   string
    RawFilename   string
    Filename      string
    PWD           string
    BuildDate     time.Time
    BuildHostname string
    BuildUser     string
}
```

The descriptor fields are pointers: `.File.Package` is a `*string`. Printing
one shows its value, but `printf`, `eq` and other functions that take any
type receive the pointer, so `printf "%s" .File.Package` formats an address
and an `eq` built on it never matches. Use the `Get*` accessors instead
(`.File.GetPackage`, `.File.GetName`, `.GetInputType`), or `.Package`.

See [`internal/generator/helpers.go`](internal/generator/helpers.go) for
the full funcmap.

## Rust paths

`rustPath root name` turns a message or enum name into the path prost and
buffa generate it at, under the module `root`:

```
{{ rustPath "crate::buffa" .GetInputType }}
```

`.acme.v1.CreateBookRequest` becomes
`crate::buffa::acme::v1::CreateBookRequest`. A nested message or enum lives
in a snake_case module named after its parent (`.acme.v1.Book.Format`
becomes `crate::buffa::acme::v1::book::Format`), type names are
UpperCamelCase (`HTTPRequest` becomes `HttpRequest`), and Rust keywords are
escaped (`type` becomes `r#type`). The name must be a type in the request;
well-known types such as `google.protobuf.Timestamp` resolve under `root`
too, not to an external crate.

## Partials and include

A `.tpl` file under `template_dir` is a partial: it is never rendered as
output, but every template can call the templates it `define`s. A `.tmpl`
file is parsed on its own, so a `define` in one is visible only to that
template; put helpers that several templates share in a partial, e.g.
`_helpers.tpl`. When two partials define the same name, the one found last
wins.

`include "name" data` renders the named template and returns its output as
a string, which `{{template}}` can't: it writes straight into the result.
That lets a helper template return a value, and lets a template call itself
to walk nested messages:

```
{{- /* _helpers.tpl */ -}}
{{- define "required" -}}
{{- $seen := append .seen .message.GetName -}}
{{- range .message.Field -}}
{{- if has "REQUIRED" (option "google.api.field_behavior" . | default list) }}{{ .Name }} {{ end -}}
{{- if and (eq (.GetType | toString) "TYPE_MESSAGE") (not (has (.GetTypeName | splitList "." | last) $seen)) -}}
{{- include "required" (dict "message" (getMessageType $.file .GetTypeName) "seen" $seen "file" $.file) -}}
{{- end -}}
{{- end -}}
{{- end -}}
```

```
{{- if include "required" (dict "message" $msg "seen" list "file" .File) | trim }}
...
{{- end }}
```

A template that includes itself more than 1000 times deep fails, which
catches a walk over a message that contains itself and doesn't track the
messages it has seen.

## Custom options

`option`, `hasOption` and `options` read the custom options set on any
descriptor: a file, message, field, oneof, enum, enum value, service or
method.

| Helper                       | Returns                                                |
| ---------------------------- | ------------------------------------------------------ |
| `option "full.name" desc`    | The option's value, or nothing when it isn't set.      |
| `hasOption "full.name" desc` | Whether the option is set, even to its zero value.     |
| `options desc`               | Every custom option set on `desc`, keyed by full name. |

Scalars keep their type, enums become their value name, messages become a
map keyed by field name, and repeated options become a list.

```proto
import "protoc/contrib/defaults/options.proto";

message ListBooksRequest {
  int32 page_size = 1 [(protoc.contrib.defaults.value).int32 = 50];
}
```

```
{{- range .Field }}
{{- if hasOption "protoc.contrib.defaults.value" . }}
{{ .Name }} defaults to {{ (option "protoc.contrib.defaults.value" .).int32 }}
{{- end }}
{{- end }}
```

The plugin needs no Go bindings for an option: a schema has to import the
file that declares an option to use it, so protoc sends that declaration to
the plugin along with the schema. `{{ options . | toPrettyJson }}` shows
what is set on a descriptor.

The number-based readers (`stringFieldExtension`, `boolMessageExtension`,
...) predate these helpers and are kept for existing templates; prefer
`option` in new ones.

## Migration from `protoc-gen-gotemplate`

- Binary renamed: `protoc-gen-gotemplate` → `protoc-gen-template`.
- Protoc flag renamed: `--gotemplate_out` → `--template_out`.
- Install path: `go install github.com/protoc-contrib/protoc-gen-template/cmd/protoc-gen-template@latest`.

## Contributing

To set up a development environment with [Nix](https://nixos.org):

```bash
nix develop
go test ./...
```

Or, without Nix, ensure `go` and `protoc` are on your `PATH`.

## License

[MIT](LICENSE.md)
