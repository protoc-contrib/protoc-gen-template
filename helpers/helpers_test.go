package pgghelpers

import (
	"testing"
	"text/template"
)

// invoke looks up a helper in ProtoHelpersFuncMap and invokes it against args
// via a throwaway text/template. This exercises the same code path users hit.
func invoke(t *testing.T, name string, args ...interface{}) string {
	t.Helper()

	placeholders := ""
	for i := range args {
		if i == 0 {
			placeholders = "(index . 0)"
			continue
		}
		placeholders += " (index . " + itoa(i) + ")"
	}
	src := "{{ " + name + " " + placeholders + " }}"

	tmpl, err := template.New("t").Funcs(ProtoHelpersFuncMap).Parse(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}

	buf := &stringBuilder{}
	if err := tmpl.Execute(buf, args); err != nil {
		t.Fatalf("exec %q: %v", src, err)
	}
	return buf.String()
}

type stringBuilder struct{ s string }

func (b *stringBuilder) Write(p []byte) (int, error) { b.s += string(p); return len(p), nil }
func (b *stringBuilder) String() string              { return b.s }

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := ""
	if i < 0 {
		neg = "-"
		i = -i
	}
	out := ""
	for i > 0 {
		out = string(rune('0'+i%10)) + out
		i /= 10
	}
	return neg + out
}

func TestCamelCase(t *testing.T) {
	// xstrings.ToCamelCase produces lowerCamelCase; the single-character
	// branch upper-cases. These expectations encode that behavior.
	cases := map[string]string{
		"hello_world":  "helloWorld",
		"foo_bar_baz":  "fooBarBaz",
		"a":            "A",
		"already_good": "alreadyGood",
	}
	for in, want := range cases {
		if got := invoke(t, "camelCase", in); got != want {
			t.Errorf("camelCase(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestLowerCamelCase(t *testing.T) {
	cases := map[string]string{
		"hello_world": "helloWorld",
		"foo_bar":     "fooBar",
		"a":           "a",
	}
	for in, want := range cases {
		if got := invoke(t, "lowerCamelCase", in); got != want {
			t.Errorf("lowerCamelCase(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"HelloWorld": "hello_world",
		"FooBarBaz":  "foo_bar_baz",
		"a":          "a",
	}
	for in, want := range cases {
		if got := invoke(t, "snakeCase", in); got != want {
			t.Errorf("snakeCase(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestKebabCase(t *testing.T) {
	cases := map[string]string{
		"HelloWorld": "hello-world",
		"foo_bar":    "foo-bar",
	}
	for in, want := range cases {
		if got := invoke(t, "kebabCase", in); got != want {
			t.Errorf("kebabCase(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestUpperLowerFirst(t *testing.T) {
	if got := invoke(t, "upperFirst", "hello"); got != "Hello" {
		t.Errorf("upperFirst: got %q", got)
	}
	if got := invoke(t, "lowerFirst", "HELLO"); got != "hELLO" {
		t.Errorf("lowerFirst: got %q", got)
	}
}

func TestArithmetic(t *testing.T) {
	if got := invoke(t, "add", 2, 3); got != "5" {
		t.Errorf("add: got %q", got)
	}
	if got := invoke(t, "subtract", 10, 4); got != "6" {
		t.Errorf("subtract: got %q", got)
	}
	if got := invoke(t, "multiply", 6, 7); got != "42" {
		t.Errorf("multiply: got %q", got)
	}
	if got := invoke(t, "divide", 20, 5); got != "4" {
		t.Errorf("divide: got %q", got)
	}
}

func TestDivideByZero(t *testing.T) {
	// divide panics on zero; text/template converts the panic into an
	// execution error. Assert that templates calling divide(_, 0) fail
	// rather than returning a bogus number.
	src := "{{ divide (index . 0) (index . 1) }}"
	tmpl, err := template.New("t").Funcs(ProtoHelpersFuncMap).Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	buf := &stringBuilder{}
	if err := tmpl.Execute(buf, []interface{}{1, 0}); err == nil {
		t.Fatalf("expected error on division by zero, got %q", buf.String())
	}
}

func TestContains(t *testing.T) {
	if got := invoke(t, "contains", "ell", "hello"); got != "true" {
		t.Errorf("contains present: got %q", got)
	}
	if got := invoke(t, "contains", "xyz", "hello"); got != "false" {
		t.Errorf("contains absent: got %q", got)
	}
}

func TestTrimStr(t *testing.T) {
	if got := invoke(t, "trimstr", "/", "/foo/bar/"); got != "foo/bar" {
		t.Errorf("trimstr: got %q", got)
	}
}

func TestGoNormalize(t *testing.T) {
	// Matches xstrings.ToCamelCase behavior (lowerCamelCase). The ID
	// rewrite kicks in only for inputs that look like id fields.
	if got := goNormalize("foo_bar_baz"); got != "fooBarBaz" {
		t.Errorf("goNormalize: got %q", got)
	}
	if got := goNormalize("user_id"); got != "userID" {
		t.Errorf("goNormalize id-rewrite: got %q", got)
	}
}

func TestLowerGoNormalize(t *testing.T) {
	if got := lowerGoNormalize("foo_bar_baz"); got != "fooBarBaz" {
		t.Errorf("lowerGoNormalize: got %q", got)
	}
}

func TestShortType(t *testing.T) {
	if got := shortType(".foo.bar.Baz"); got != "Baz" {
		t.Errorf("shortType: got %q", got)
	}
	if got := shortType("Baz"); got != "Baz" {
		t.Errorf("shortType (no dots): got %q", got)
	}
}

func TestNamespacedFlowType(t *testing.T) {
	if got := namespacedFlowType(".foo.bar.Baz"); got != "foo$bar$Baz" {
		t.Errorf("namespacedFlowType: got %q", got)
	}
}

func TestFuncMapRegistered(t *testing.T) {
	// Guard against accidental removal — these are the high-visibility helpers.
	wantKeys := []string{
		"camelCase", "lowerCamelCase", "snakeCase", "kebabCase",
		"upperFirst", "lowerFirst",
		"add", "subtract", "multiply", "divide",
		"contains", "trimstr", "json", "prettyjson",
		"goType", "goPkg", "jsType", "haskellType",
		"httpVerb", "httpPath", "httpBody",
	}
	for _, k := range wantKeys {
		if _, ok := ProtoHelpersFuncMap[k]; !ok {
			t.Errorf("funcmap missing key %q", k)
		}
	}
}
