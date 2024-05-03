package ftl

import (
	"context"
	"fmt"
	"reflect"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/rpc"
)

func call[Req, Resp any](ctx context.Context, callee Ref, req Req) (resp Resp, err error) {
	reqData, err := encoding.Marshal(req)
	if err != nil {
		return resp, fmt.Errorf("%s: failed to marshal request: %w", callee, err)
	}

	behavior, err := modulecontext.FromContext(ctx).BehaviorForVerb(modulecontext.Ref(callee))
	if err != nil {
		return resp, fmt.Errorf("%s: %w", callee, err)
	}
	switch behavior := behavior.(type) {
	case modulecontext.MockBehavior:
		uncheckedResp, err := behavior.Mock(ctx, req)
		if err != nil {
			return resp, fmt.Errorf("%s: %w", callee, err)
		}
		if r, ok := uncheckedResp.(Resp); ok {
			return r, nil
		}
		return resp, fmt.Errorf("%s: overridden verb had invalid response type %T, expected %v", callee, uncheckedResp, reflect.TypeFor[Resp]())
	case modulecontext.DirectBehavior:
		panic("not implemented")
	case modulecontext.StandardBehavior:
		client := rpc.ClientFromContext[ftlv1connect.VerbServiceClient](ctx)
		cresp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: callee.ToProto(), Body: reqData}))
		if err != nil {
			return resp, fmt.Errorf("%s: failed to call Verb: %w", callee, err)
		}
		switch cresp := cresp.Msg.Response.(type) {
		case *ftlv1.CallResponse_Error_:
			return resp, fmt.Errorf("%s: %s", callee, cresp.Error.Message)

		case *ftlv1.CallResponse_Body:
			err = encoding.Unmarshal(cresp.Body, &resp)
			if err != nil {
				return resp, fmt.Errorf("%s: failed to decode response: %w", callee, err)
			}
			return resp, nil

		default:
			panic(fmt.Sprintf("%s: invalid response type %T", callee, cresp))
		}
	default:
		panic(fmt.Sprintf("unknown behavior: %s", behavior))
	}
}

// Call a Verb through the FTL Controller.
func Call[Req, Resp any](ctx context.Context, verb Verb[Req, Resp], req Req) (Resp, error) {
	return call[Req, Resp](ctx, FuncRef(verb), req)
}

// CallSink calls a Sink through the FTL controller.
func CallSink[Req any](ctx context.Context, sink Sink[Req], req Req) error {
	_, err := call[Req, Unit](ctx, FuncRef(sink), req)
	return err
}

// CallSource calls a Source through the FTL controller.
func CallSource[Resp any](ctx context.Context, source Source[Resp]) (Resp, error) {
	return call[Unit, Resp](ctx, FuncRef(source), Unit{})
}

// CallEmpty calls a Verb with no request or response through the FTL controller.
func CallEmpty(ctx context.Context, empty Empty) error {
	_, err := call[Unit, Unit](ctx, FuncRef(empty), Unit{})
	return err
}
