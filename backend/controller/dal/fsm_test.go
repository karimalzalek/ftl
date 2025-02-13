package dal

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/either"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/db/dalerrs"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestSendFSMEvent(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)

	_, err = dal.AcquireAsyncCall(ctx)
	assert.IsError(t, err, dalerrs.ErrNotFound)

	ref := schema.RefKey{Module: "module", Name: "verb"}
	err = dal.StartFSMTransition(ctx, schema.RefKey{Module: "test", Name: "test"}, "invoiceID", ref, []byte(`{}`), schema.RetryParams{})
	assert.NoError(t, err)

	err = dal.StartFSMTransition(ctx, schema.RefKey{Module: "test", Name: "test"}, "invoiceID", ref, []byte(`{}`), schema.RetryParams{})
	assert.IsError(t, err, dalerrs.ErrConflict)
	assert.EqualError(t, err, "transition already executing: conflict")

	call, err := dal.AcquireAsyncCall(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() {
		err := call.Lease.Release()
		assert.NoError(t, err)
	})

	assert.HasPrefix(t, call.Lease.String(), "/system/async_call/1:")
	expectedCall := &AsyncCall{
		ID:   1,
		Verb: ref,
		Origin: AsyncOriginFSM{
			FSM: schema.RefKey{Module: "test", Name: "test"},
			Key: "invoiceID",
		},
		Request: []byte(`{}`),
	}
	assert.Equal(t, expectedCall, call, assert.Exclude[*Lease](), assert.Exclude[time.Time]())

	err = dal.CompleteAsyncCall(ctx, call, either.LeftOf[string]([]byte(`{}`)), func(tx *Tx) error { return nil })
	assert.NoError(t, err)

	actual, err := dal.LoadAsyncCall(ctx, call.ID)
	assert.NoError(t, err)
	assert.Equal(t, call, actual, assert.Exclude[*Lease](), assert.Exclude[time.Time]())
}
