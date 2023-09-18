// @generated by protoc-gen-connect-es v0.13.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/ftl.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CallRequest, CallResponse, CreateDeploymentRequest, CreateDeploymentResponse, DeployRequest, DeployResponse, GetArtefactDiffsRequest, GetArtefactDiffsResponse, GetDeploymentArtefactsRequest, GetDeploymentArtefactsResponse, GetDeploymentRequest, GetDeploymentResponse, GetSchemaRequest, GetSchemaResponse, PingRequest, PingResponse, ProcessListRequest, ProcessListResponse, PullSchemaRequest, PullSchemaResponse, RegisterRunnerRequest, RegisterRunnerResponse, ReplaceDeployRequest, ReplaceDeployResponse, ReserveRequest, ReserveResponse, StatusRequest, StatusResponse, StreamDeploymentLogsRequest, StreamDeploymentLogsResponse, TerminateRequest, UpdateDeployRequest, UpdateDeployResponse, UploadArtefactRequest, UploadArtefactResponse } from "./ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";

/**
 * VerbService is a common interface shared by multiple services for calling Verbs.
 *
 * @generated from service xyz.block.ftl.v1.VerbService
 */
export const VerbService = {
  typeName: "xyz.block.ftl.v1.VerbService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.VerbService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Issue a synchronous call to a Verb.
     *
     * @generated from rpc xyz.block.ftl.v1.VerbService.Call
     */
    call: {
      name: "Call",
      I: CallRequest,
      O: CallResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * @generated from service xyz.block.ftl.v1.ControllerService
 */
export const ControllerService = {
  typeName: "xyz.block.ftl.v1.ControllerService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * List "processes" running on the cluster.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.ProcessList
     */
    processList: {
      name: "ProcessList",
      I: ProcessListRequest,
      O: ProcessListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1.ControllerService.Status
     */
    status: {
      name: "Status",
      I: StatusRequest,
      O: StatusResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get list of artefacts that differ between the server and client.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetArtefactDiffs
     */
    getArtefactDiffs: {
      name: "GetArtefactDiffs",
      I: GetArtefactDiffsRequest,
      O: GetArtefactDiffsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Upload an artefact to the server.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.UploadArtefact
     */
    uploadArtefact: {
      name: "UploadArtefact",
      I: UploadArtefactRequest,
      O: UploadArtefactResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Create a deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.CreateDeployment
     */
    createDeployment: {
      name: "CreateDeployment",
      I: CreateDeploymentRequest,
      O: CreateDeploymentResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get the schema and artefact metadata for a deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetDeployment
     */
    getDeployment: {
      name: "GetDeployment",
      I: GetDeploymentRequest,
      O: GetDeploymentResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream deployment artefacts from the server.
     *
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts
     */
    getDeploymentArtefacts: {
      name: "GetDeploymentArtefacts",
      I: GetDeploymentArtefactsRequest,
      O: GetDeploymentArtefactsResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Register a Runner with the Controller.
     *
     * Each runner issue a RegisterRunnerRequest to the ControllerService
     * every 10 seconds to maintain its heartbeat.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.RegisterRunner
     */
    registerRunner: {
      name: "RegisterRunner",
      I: RegisterRunnerRequest,
      O: RegisterRunnerResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Update an existing deployment.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.UpdateDeploy
     */
    updateDeploy: {
      name: "UpdateDeploy",
      I: UpdateDeployRequest,
      O: UpdateDeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Gradually replace an existing deployment with a new one.
     *
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.ReplaceDeploy
     */
    replaceDeploy: {
      name: "ReplaceDeploy",
      I: ReplaceDeployRequest,
      O: ReplaceDeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream logs from a deployment
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.StreamDeploymentLogs
     */
    streamDeploymentLogs: {
      name: "StreamDeploymentLogs",
      I: StreamDeploymentLogsRequest,
      O: StreamDeploymentLogsResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Get the full schema.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.GetSchema
     */
    getSchema: {
      name: "GetSchema",
      I: GetSchemaRequest,
      O: GetSchemaResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Pull schema changes from the Controller.
     *
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     *
     * @generated from rpc xyz.block.ftl.v1.ControllerService.PullSchema
     */
    pullSchema: {
      name: "PullSchema",
      I: PullSchemaRequest,
      O: PullSchemaResponse,
      kind: MethodKind.ServerStreaming,
    },
  }
} as const;

/**
 * RunnerService is the service that executes Deployments.
 *
 * The Controller will scale the Runner horizontally as required. The Runner will
 * register itself automatically with the ControllerService, which will then
 * assign modules to it.
 *
 * @generated from service xyz.block.ftl.v1.RunnerService
 */
export const RunnerService = {
  typeName: "xyz.block.ftl.v1.RunnerService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Reserve synchronously reserves a Runner for a deployment but does nothing else.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Reserve
     */
    reserve: {
      name: "Reserve",
      I: ReserveRequest,
      O: ReserveResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Initiate a deployment on this Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Deploy
     */
    deploy: {
      name: "Deploy",
      I: DeployRequest,
      O: DeployResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Terminate the deployment on this Runner.
     *
     * @generated from rpc xyz.block.ftl.v1.RunnerService.Terminate
     */
    terminate: {
      name: "Terminate",
      I: TerminateRequest,
      O: RegisterRunnerRequest,
      kind: MethodKind.Unary,
    },
  }
} as const;

