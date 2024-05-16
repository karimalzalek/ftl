package buildengine

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestGenerateGoModule(t *testing.T) {
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "other", Decls: []schema.Decl{
				&schema.Enum{
					Comments: []string{"This is an enum.", "", "It has 3 variants."},
					Name:     "Color",
					Export:   true,
					Type:     &schema.String{},
					Variants: []*schema.EnumVariant{
						{Name: "Red", Value: &schema.StringValue{Value: "Red"}},
						{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
						{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
					},
				},
				&schema.Enum{
					Name:   "ColorInt",
					Export: true,
					Type:   &schema.Int{},
					Variants: []*schema.EnumVariant{
						{Name: "RedInt", Value: &schema.IntValue{Value: 0}},
						{Name: "BlueInt", Value: &schema.IntValue{Value: 1}},
						{Name: "GreenInt", Value: &schema.IntValue{Value: 2}},
					},
				},
				&schema.Enum{
					Comments: []string{"This is type enum."},
					Name:     "TypeEnum",
					Export:   true,
					Variants: []*schema.EnumVariant{
						{Name: "A", Value: &schema.TypeValue{Value: &schema.Int{}}},
						{Name: "B", Value: &schema.TypeValue{Value: &schema.String{}}},
					},
				},
				&schema.Data{Name: "EchoRequest", Export: true},
				&schema.Data{
					Comments: []string{"This is an echo data response."},
					Name:     "EchoResponse", Export: true},
				&schema.Verb{
					Name:     "echo",
					Export:   true,
					Request:  &schema.Ref{Name: "EchoRequest"},
					Response: &schema.Ref{Name: "EchoResponse"},
				},
				&schema.Data{Name: "SinkReq", Export: true},
				&schema.Verb{
					Comments: []string{"This is a sink verb.", "", "Here is another line for this comment!"},
					Name:     "sink",
					Export:   true,
					Request:  &schema.Ref{Name: "SinkReq"},
					Response: &schema.Unit{},
				},
				&schema.Data{Name: "SourceResp", Export: true},
				&schema.Verb{
					Name:     "source",
					Export:   true,
					Request:  &schema.Unit{},
					Response: &schema.Ref{Name: "SourceResp"},
				},
				&schema.Verb{
					Name:     "nothing",
					Export:   true,
					Request:  &schema.Unit{},
					Response: &schema.Unit{},
				},
			}},
			{Name: "test"},
		},
	}
	expected := `// Code generated by FTL. DO NOT EDIT.

package other

import (
  "context"
)

var _ = context.Background

// This is an enum.
//
// It has 3 variants.
//
//ftl:enum
type Color string
const (
  Red Color = "Red"
  Blue Color = "Blue"
  Green Color = "Green"
)

//ftl:enum
type ColorInt int
const (
  RedInt ColorInt = 0
  BlueInt ColorInt = 1
  GreenInt ColorInt = 2
)

// This is type enum.
//
//ftl:enum
type TypeEnum interface { typeEnum() }

type A int

func (A) typeEnum() {}

type B string

func (B) typeEnum() {}

type EchoRequest struct {
}

// This is an echo data response.
//
type EchoResponse struct {
}

//ftl:verb
func Echo(context.Context, EchoRequest) (EchoResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.Call()")
}

type SinkReq struct {
}

// This is a sink verb.
//
// Here is another line for this comment!
//
//ftl:verb
func Sink(context.Context, SinkReq) error {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.CallSink()")
}

type SourceResp struct {
}

//ftl:verb
func Source(context.Context) (SourceResp, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.CallSource()")
}

//ftl:verb
func Nothing(context.Context) error {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.CallEmpty()")
}
`
	bctx := buildContext{
		moduleDir: "testdata/projects/another",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuild(t, bctx, "", []assertion{
		assertGeneratedModule("go/modules/other/external_module.go", expected),
	})
}

func TestGoBuildClearsBuildDir(t *testing.T) {
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "test"},
		},
	}
	bctx := buildContext{
		moduleDir: "testdata/projects/another",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuildClearsBuildDir(t, bctx)
}

func TestMetadataImportsExcluded(t *testing.T) {
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "test", Decls: []schema.Decl{
				&schema.Data{
					Comments: []string{"Request data type."},
					Name:     "Req", Export: true},
				&schema.Data{Name: "Resp", Export: true},
				&schema.Verb{
					Comments: []string{"This is a verb."},
					Name:     "call",
					Export:   true,
					Request:  &schema.Ref{Name: "Req"},
					Response: &schema.Ref{Name: "Resp"},
					Metadata: []schema.Metadata{
						&schema.MetadataCalls{Calls: []*schema.Ref{{Name: "verb", Module: "other"}}},
					},
				},
			}},
		},
	}
	expected := `// Code generated by FTL. DO NOT EDIT.

package test

import (
  "context"
)

var _ = context.Background

// Request data type.
//
type Req struct {
}

type Resp struct {
}

// This is a verb.
//
//ftl:verb
func Call(context.Context, Req) (Resp, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.Call()")
}
`
	bctx := buildContext{
		moduleDir: "testdata/projects/another",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuild(t, bctx, "", []assertion{
		assertGeneratedModule("go/modules/test/external_module.go", expected),
	})
}

func TestExternalType(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	bctx := buildContext{
		moduleDir: "testdata/projects/external",
		buildDir:  "_ftl",
		sch:       &schema.Schema{},
	}
	testBuild(t, bctx, "unsupported external type", []assertion{
		assertBuildProtoErrors(
			"unsupported external type \"time.Month\"",
			"unsupported type \"time.Month\" for field \"Month\"",
			"unsupported response type \"ftl/external.ExternalResponse\"",
		),
	})
}

func TestGoModVersion(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "highgoversion", Decls: []schema.Decl{
				&schema.Data{Name: "EchoReq"},
				&schema.Data{Name: "EchoResp"},
				&schema.Verb{
					Name:     "echo",
					Request:  &schema.Ref{Name: "EchoRequest"},
					Response: &schema.Ref{Name: "EchoResponse"},
				},
			}},
		},
	}
	bctx := buildContext{
		moduleDir: "testdata/projects/highgoversion",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuild(t, bctx, fmt.Sprintf("go version %q is not recent enough for this module, needs minimum version \"9000.1.1\"", runtime.Version()[2:]), []assertion{})
}

func TestGeneratedTypeRegistry(t *testing.T) {
	if !testing.Short() {
		t.Skipf("skipping test in non-short mode")
	}
	sch := &schema.Schema{
		Modules: []*schema.Module{
			{Name: "another", Decls: []schema.Decl{
				&schema.Enum{
					Name:   "TypeEnum",
					Export: true,
					Variants: []*schema.EnumVariant{
						{Name: "A", Value: &schema.TypeValue{Value: &schema.Int{}}},
						{Name: "B", Value: &schema.TypeValue{Value: &schema.String{}}},
					},
				},
				&schema.Enum{
					Name:   "SecondTypeEnum",
					Export: true,
					Variants: []*schema.EnumVariant{
						{Name: "One", Value: &schema.TypeValue{Value: &schema.Int{}}},
						{Name: "Two", Value: &schema.TypeValue{Value: &schema.String{}}},
					},
				},
				&schema.Data{
					Name:   "TransitiveTypeEnum",
					Export: true,
					Fields: []*schema.Field{
						{Name: "TypeEnumRef", Type: &schema.Ref{Name: "SecondTypeEnum", Module: "another"}},
					},
				},
			}},
		},
	}
	expected, err := os.ReadFile("testdata/type_registry_main.go")
	assert.NoError(t, err)
	bctx := buildContext{
		moduleDir: "testdata/projects/other",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuild(t, bctx, "", []assertion{
		assertGeneratedMain(string(expected)),
	})
}
