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
- **Sprig funcmap** — every helper from
  [Masterminds/sprig](https://github.com/Masterminds/sprig) is available
  (date, string, crypto, flow, default, dictionary, and more).
- **Proto-aware funcmap** — naming (`camelCase`, `lowerCamelCase`,
  `snakeCase`, `kebabCase`, `goNormalize`, `shortType`, ...), arithmetic
  (`add`, `subtract`, `multiply`, `divide`), field/message walkers
  (`getMessageType`, `isFieldMessage`, `fieldMapKeyType`, ...), HTTP
  annotation accessors (`httpPath`, `httpVerb`, `httpBody`, ...), and
  extension readers (`stringFieldExtension`, `boolFieldExtension`, ...).
- **Flexible iteration** — `mode=service` (default) renders once per gRPC
  service; `mode=file` renders once per file that declares a service;
  `mode=all` renders every proto file whether or not it declares a service.
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

See [`internal/generator/helpers.go`](internal/generator/helpers.go) for
the full funcmap.

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
