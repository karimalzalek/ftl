// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package sql

import (
	"context"

	"github.com/TBD54566975/ftl/backend/controller/internal/sqltypes"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	AssociateArtefactWithDeployment(ctx context.Context, arg AssociateArtefactWithDeploymentParams) error
	// Create a new artefact and return the artefact ID.
	CreateArtefact(ctx context.Context, digest []byte, content []byte) (int64, error)
	CreateDeployment(ctx context.Context, key sqltypes.Key, moduleName string, schema []byte) error
	CreateIngressRequest(ctx context.Context, key sqltypes.Key, sourceAddr string) error
	CreateIngressRoute(ctx context.Context, arg CreateIngressRouteParams) error
	DeregisterRunner(ctx context.Context, key sqltypes.Key) (int64, error)
	ExpireRunnerReservations(ctx context.Context) (int64, error)
	GetActiveRunners(ctx context.Context, all bool) ([]GetActiveRunnersRow, error)
	GetAllIngressRoutes(ctx context.Context, all bool) ([]GetAllIngressRoutesRow, error)
	GetArtefactContentRange(ctx context.Context, start int32, count int32, iD int64) ([]byte, error)
	// Return the digests that exist in the database.
	GetArtefactDigests(ctx context.Context, digests [][]byte) ([]GetArtefactDigestsRow, error)
	GetControllers(ctx context.Context, all bool) ([]Controller, error)
	GetDeployment(ctx context.Context, key sqltypes.Key) (GetDeploymentRow, error)
	// Get all artefacts matching the given digests.
	GetDeploymentArtefacts(ctx context.Context, deploymentID int64) ([]GetDeploymentArtefactsRow, error)
	GetDeploymentLogs(ctx context.Context, deploymentKey sqltypes.NullKey, afterTimestamp pgtype.Timestamptz, afterID int64) ([]GetDeploymentLogsRow, error)
	GetDeployments(ctx context.Context, all bool) ([]GetDeploymentsRow, error)
	GetDeploymentsByID(ctx context.Context, ids []int64) ([]Deployment, error)
	// Get deployments that have a mismatch between the number of assigned and required replicas.
	GetDeploymentsNeedingReconciliation(ctx context.Context) ([]GetDeploymentsNeedingReconciliationRow, error)
	// Get all deployments that have artefacts matching the given digests.
	GetDeploymentsWithArtefacts(ctx context.Context, digests [][]byte, count interface{}) ([]GetDeploymentsWithArtefactsRow, error)
	GetExistingDeploymentForModule(ctx context.Context, name string) (Deployment, error)
	GetIdleRunners(ctx context.Context, labels []byte, limit int32) ([]Runner, error)
	// Get the runner endpoints corresponding to the given ingress route.
	GetIngressRoutes(ctx context.Context, method string, path string) ([]GetIngressRoutesRow, error)
	GetModuleCalls(ctx context.Context, modules []string) ([]GetModuleCallsRow, error)
	GetModulesByID(ctx context.Context, ids []int64) ([]Module, error)
	GetRequestCalls(ctx context.Context, key sqltypes.Key) ([]GetRequestCallsRow, error)
	GetRoutingTable(ctx context.Context, name string) ([]GetRoutingTableRow, error)
	GetRunner(ctx context.Context, key sqltypes.Key) (GetRunnerRow, error)
	GetRunnerState(ctx context.Context, key sqltypes.Key) (RunnerState, error)
	GetRunnersForDeployment(ctx context.Context, key sqltypes.Key) ([]Runner, error)
	InsertCallEntry(ctx context.Context, arg InsertCallEntryParams) error
	InsertDeploymentLogEntry(ctx context.Context, arg InsertDeploymentLogEntryParams) error
	// Mark any controller entries that haven't been updated recently as dead.
	KillStaleControllers(ctx context.Context, dollar_1 pgtype.Interval) (int64, error)
	KillStaleRunners(ctx context.Context, dollar_1 pgtype.Interval) (int64, error)
	ReplaceDeployment(ctx context.Context, oldDeployment sqltypes.Key, newDeployment sqltypes.Key, minReplicas int32) (int64, error)
	// Find an idle runner and reserve it for the given deployment.
	ReserveRunner(ctx context.Context, reservationTimeout pgtype.Timestamptz, deploymentKey sqltypes.Key, labels []byte) (Runner, error)
	SetDeploymentDesiredReplicas(ctx context.Context, key sqltypes.Key, minReplicas int32) error
	UpsertController(ctx context.Context, key sqltypes.Key, endpoint string) (int64, error)
	UpsertModule(ctx context.Context, language string, name string) (int64, error)
	// Upsert a runner and return the deployment ID that it is assigned to, if any.
	// If the deployment key is null, then deployment_rel.id will be null,
	// otherwise we try to retrieve the deployments.id using the key. If
	// there is no corresponding deployment, then the deployment ID is -1
	// and the parent statement will fail due to a foreign key constraint.
	UpsertRunner(ctx context.Context, arg UpsertRunnerParams) (pgtype.Int8, error)
}

var _ Querier = (*Queries)(nil)
