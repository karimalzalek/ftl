package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jpillora/backoff"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scaling/localscaling"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type boxRunCmd struct {
	Recreate          bool          `help:"Recreate the database."`
	DSN               string        `help:"DSN for the database." default:"postgres://postgres:secret@localhost:5432/ftl?sslmode=disable" env:"FTL_CONTROLLER_DSN"`
	IngressBind       *url.URL      `help:"Bind address for the ingress server." default:"http://0.0.0.0:8891" env:"FTL_INGRESS_BIND"`
	Bind              *url.URL      `help:"Bind address for the FTL controller." default:"http://0.0.0.0:8892" env:"FTL_BIND"`
	RunnerBase        *url.URL      `help:"Base bind address for FTL runners." default:"http://127.0.0.1:8893" env:"FTL_RUNNER_BIND"`
	Dir               string        `arg:"" help:"Directory to scan for precompiled modules." default:"."`
	ControllerTimeout time.Duration `help:"Timeout for Controller start." default:"30s"`
}

func (b *boxRunCmd) Run(ctx context.Context) error {
	conn, err := databasetesting.CreateForDevel(ctx, b.DSN, b.Recreate)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	dal, err := dal.New(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to create DAL: %w", err)
	}
	config := controller.Config{
		Bind:        b.Bind,
		IngressBind: b.IngressBind,
		Key:         model.NewLocalControllerKey(0),
		DSN:         b.DSN,
	}
	if err := kong.ApplyDefaults(&config); err != nil {
		return err
	}

	// Start the controller.
	runnerPortAllocator, err := bind.NewBindAllocator(b.RunnerBase)
	if err != nil {
		return fmt.Errorf("failed to create runner port allocator: %w", err)
	}
	runnerScaling, err := localscaling.NewLocalScaling(runnerPortAllocator, []*url.URL{b.Bind})
	if err != nil {
		return fmt.Errorf("failed to create runner autoscaler: %w", err)
	}
	wg := errgroup.Group{}
	wg.Go(func() error {
		return controller.Start(ctx, config, runnerScaling, dal)
	})

	// Wait for the controller to come up.
	client := ftlv1connect.NewControllerServiceClient(rpc.GetHTTPClient(b.Bind.String()), b.Bind.String())
	waitCtx, cancel := context.WithTimeout(ctx, b.ControllerTimeout)
	defer cancel()
	if err := rpc.Wait(waitCtx, backoff.Backoff{}, client); err != nil {
		return fmt.Errorf("controller failed to start: %w", err)
	}

	engine, err := buildengine.New(ctx, client, []string{b.Dir})
	if err != nil {
		return fmt.Errorf("failed to create build engine: %w", err)
	}

	if err := engine.Deploy(ctx, 1, true); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}
	return wg.Wait()
}
