// @generated by protoc-gen-connect-es v0.11.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/console/console.proto (package xyz.block.ftl.v1.console, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { GetCallsRequest, GetCallsResponse, GetModulesRequest, GetModulesResponse, GetRequestCallsRequest, GetRequestCallsResponse } from "./console_pb.js";

/**
 * @generated from service xyz.block.ftl.v1.console.ConsoleService
 */
export const ConsoleService = {
  typeName: "xyz.block.ftl.v1.console.ConsoleService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.console.ConsoleService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1.console.ConsoleService.GetModules
     */
    getModules: {
      name: "GetModules",
      I: GetModulesRequest,
      O: GetModulesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1.console.ConsoleService.GetCalls
     */
    getCalls: {
      name: "GetCalls",
      I: GetCallsRequest,
      O: GetCallsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.v1.console.ConsoleService.GetRequestCalls
     */
    getRequestCalls: {
      name: "GetRequestCalls",
      I: GetRequestCallsRequest,
      O: GetRequestCallsResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

