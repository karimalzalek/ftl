package other

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.

	"ftl/another"
)

//ftl:enum
type D interface {
	tag()
}

type Bool bool

func (Bool) tag() {}

type Bytes []byte

func (Bytes) tag() {}

type Float float64

func (Float) tag() {}

type Int int

func (Int) tag() {}

//type Time time.Time
//
//func (Time) tag() {}

type U ftl.Unit

func (U) tag() {}

type List []string

func (List) tag() {}

type Map map[string]string

func (Map) tag() {}

type String string

func (String) tag() {}

type Struct struct{}

func (Struct) tag() {}

//ftl:enum
type SecondSumType interface {
	tag2()
}

type A string

func (A) tag2() {}

type B EchoRequest

func (B) tag2() {}

type EchoRequest struct {
	Name            ftl.Option[string] `json:"name"`
	ExternalSumType another.TypeEnum
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
