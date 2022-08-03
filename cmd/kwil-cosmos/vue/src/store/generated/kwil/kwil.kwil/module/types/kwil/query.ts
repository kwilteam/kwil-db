/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { Params } from "../kwil/params";
import { Databases } from "../kwil/databases";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";
import { Ddl } from "../kwil/ddl";
import { Ddlindex } from "../kwil/ddlindex";
import { Queryids } from "../kwil/queryids";

export const protobufPackage = "kwil.kwil";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetDatabasesRequest {
  index: string;
}

export interface QueryGetDatabasesResponse {
  databases: Databases | undefined;
}

export interface QueryAllDatabasesRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllDatabasesResponse {
  databases: Databases[];
  pagination: PageResponse | undefined;
}

export interface QueryGetDdlRequest {
  index: string;
}

export interface QueryGetDdlResponse {
  ddl: Ddl | undefined;
}

export interface QueryAllDdlRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllDdlResponse {
  ddl: Ddl[];
  pagination: PageResponse | undefined;
}

export interface QueryGetDdlindexRequest {
  index: string;
}

export interface QueryGetDdlindexResponse {
  ddlindex: Ddlindex | undefined;
}

export interface QueryAllDdlindexRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllDdlindexResponse {
  ddlindex: Ddlindex[];
  pagination: PageResponse | undefined;
}

export interface QueryGetQueryidsRequest {
  index: string;
}

export interface QueryGetQueryidsResponse {
  queryids: Queryids | undefined;
}

export interface QueryAllQueryidsRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllQueryidsResponse {
  queryids: Queryids[];
  pagination: PageResponse | undefined;
}

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryGetDatabasesRequest: object = { index: "" };

export const QueryGetDatabasesRequest = {
  encode(
    message: QueryGetDatabasesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDatabasesRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDatabasesRequest,
    } as QueryGetDatabasesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDatabasesRequest {
    const message = {
      ...baseQueryGetDatabasesRequest,
    } as QueryGetDatabasesRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    return message;
  },

  toJSON(message: QueryGetDatabasesRequest): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDatabasesRequest>
  ): QueryGetDatabasesRequest {
    const message = {
      ...baseQueryGetDatabasesRequest,
    } as QueryGetDatabasesRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    return message;
  },
};

const baseQueryGetDatabasesResponse: object = {};

export const QueryGetDatabasesResponse = {
  encode(
    message: QueryGetDatabasesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.databases !== undefined) {
      Databases.encode(message.databases, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDatabasesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDatabasesResponse,
    } as QueryGetDatabasesResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.databases = Databases.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDatabasesResponse {
    const message = {
      ...baseQueryGetDatabasesResponse,
    } as QueryGetDatabasesResponse;
    if (object.databases !== undefined && object.databases !== null) {
      message.databases = Databases.fromJSON(object.databases);
    } else {
      message.databases = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetDatabasesResponse): unknown {
    const obj: any = {};
    message.databases !== undefined &&
      (obj.databases = message.databases
        ? Databases.toJSON(message.databases)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDatabasesResponse>
  ): QueryGetDatabasesResponse {
    const message = {
      ...baseQueryGetDatabasesResponse,
    } as QueryGetDatabasesResponse;
    if (object.databases !== undefined && object.databases !== null) {
      message.databases = Databases.fromPartial(object.databases);
    } else {
      message.databases = undefined;
    }
    return message;
  },
};

const baseQueryAllDatabasesRequest: object = {};

export const QueryAllDatabasesRequest = {
  encode(
    message: QueryAllDatabasesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllDatabasesRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllDatabasesRequest,
    } as QueryAllDatabasesRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDatabasesRequest {
    const message = {
      ...baseQueryAllDatabasesRequest,
    } as QueryAllDatabasesRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDatabasesRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllDatabasesRequest>
  ): QueryAllDatabasesRequest {
    const message = {
      ...baseQueryAllDatabasesRequest,
    } as QueryAllDatabasesRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllDatabasesResponse: object = {};

export const QueryAllDatabasesResponse = {
  encode(
    message: QueryAllDatabasesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.databases) {
      Databases.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllDatabasesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllDatabasesResponse,
    } as QueryAllDatabasesResponse;
    message.databases = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.databases.push(Databases.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDatabasesResponse {
    const message = {
      ...baseQueryAllDatabasesResponse,
    } as QueryAllDatabasesResponse;
    message.databases = [];
    if (object.databases !== undefined && object.databases !== null) {
      for (const e of object.databases) {
        message.databases.push(Databases.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDatabasesResponse): unknown {
    const obj: any = {};
    if (message.databases) {
      obj.databases = message.databases.map((e) =>
        e ? Databases.toJSON(e) : undefined
      );
    } else {
      obj.databases = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllDatabasesResponse>
  ): QueryAllDatabasesResponse {
    const message = {
      ...baseQueryAllDatabasesResponse,
    } as QueryAllDatabasesResponse;
    message.databases = [];
    if (object.databases !== undefined && object.databases !== null) {
      for (const e of object.databases) {
        message.databases.push(Databases.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetDdlRequest: object = { index: "" };

export const QueryGetDdlRequest = {
  encode(
    message: QueryGetDdlRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetDdlRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetDdlRequest } as QueryGetDdlRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDdlRequest {
    const message = { ...baseQueryGetDdlRequest } as QueryGetDdlRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    return message;
  },

  toJSON(message: QueryGetDdlRequest): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryGetDdlRequest>): QueryGetDdlRequest {
    const message = { ...baseQueryGetDdlRequest } as QueryGetDdlRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    return message;
  },
};

const baseQueryGetDdlResponse: object = {};

export const QueryGetDdlResponse = {
  encode(
    message: QueryGetDdlResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.ddl !== undefined) {
      Ddl.encode(message.ddl, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetDdlResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetDdlResponse } as QueryGetDdlResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ddl = Ddl.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDdlResponse {
    const message = { ...baseQueryGetDdlResponse } as QueryGetDdlResponse;
    if (object.ddl !== undefined && object.ddl !== null) {
      message.ddl = Ddl.fromJSON(object.ddl);
    } else {
      message.ddl = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetDdlResponse): unknown {
    const obj: any = {};
    message.ddl !== undefined &&
      (obj.ddl = message.ddl ? Ddl.toJSON(message.ddl) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryGetDdlResponse>): QueryGetDdlResponse {
    const message = { ...baseQueryGetDdlResponse } as QueryGetDdlResponse;
    if (object.ddl !== undefined && object.ddl !== null) {
      message.ddl = Ddl.fromPartial(object.ddl);
    } else {
      message.ddl = undefined;
    }
    return message;
  },
};

const baseQueryAllDdlRequest: object = {};

export const QueryAllDdlRequest = {
  encode(
    message: QueryAllDdlRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllDdlRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllDdlRequest } as QueryAllDdlRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDdlRequest {
    const message = { ...baseQueryAllDdlRequest } as QueryAllDdlRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDdlRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryAllDdlRequest>): QueryAllDdlRequest {
    const message = { ...baseQueryAllDdlRequest } as QueryAllDdlRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllDdlResponse: object = {};

export const QueryAllDdlResponse = {
  encode(
    message: QueryAllDdlResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.ddl) {
      Ddl.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllDdlResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllDdlResponse } as QueryAllDdlResponse;
    message.ddl = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ddl.push(Ddl.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDdlResponse {
    const message = { ...baseQueryAllDdlResponse } as QueryAllDdlResponse;
    message.ddl = [];
    if (object.ddl !== undefined && object.ddl !== null) {
      for (const e of object.ddl) {
        message.ddl.push(Ddl.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDdlResponse): unknown {
    const obj: any = {};
    if (message.ddl) {
      obj.ddl = message.ddl.map((e) => (e ? Ddl.toJSON(e) : undefined));
    } else {
      obj.ddl = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryAllDdlResponse>): QueryAllDdlResponse {
    const message = { ...baseQueryAllDdlResponse } as QueryAllDdlResponse;
    message.ddl = [];
    if (object.ddl !== undefined && object.ddl !== null) {
      for (const e of object.ddl) {
        message.ddl.push(Ddl.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetDdlindexRequest: object = { index: "" };

export const QueryGetDdlindexRequest = {
  encode(
    message: QueryGetDdlindexRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetDdlindexRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDdlindexRequest,
    } as QueryGetDdlindexRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDdlindexRequest {
    const message = {
      ...baseQueryGetDdlindexRequest,
    } as QueryGetDdlindexRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    return message;
  },

  toJSON(message: QueryGetDdlindexRequest): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDdlindexRequest>
  ): QueryGetDdlindexRequest {
    const message = {
      ...baseQueryGetDdlindexRequest,
    } as QueryGetDdlindexRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    return message;
  },
};

const baseQueryGetDdlindexResponse: object = {};

export const QueryGetDdlindexResponse = {
  encode(
    message: QueryGetDdlindexResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.ddlindex !== undefined) {
      Ddlindex.encode(message.ddlindex, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetDdlindexResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetDdlindexResponse,
    } as QueryGetDdlindexResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ddlindex = Ddlindex.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetDdlindexResponse {
    const message = {
      ...baseQueryGetDdlindexResponse,
    } as QueryGetDdlindexResponse;
    if (object.ddlindex !== undefined && object.ddlindex !== null) {
      message.ddlindex = Ddlindex.fromJSON(object.ddlindex);
    } else {
      message.ddlindex = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetDdlindexResponse): unknown {
    const obj: any = {};
    message.ddlindex !== undefined &&
      (obj.ddlindex = message.ddlindex
        ? Ddlindex.toJSON(message.ddlindex)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetDdlindexResponse>
  ): QueryGetDdlindexResponse {
    const message = {
      ...baseQueryGetDdlindexResponse,
    } as QueryGetDdlindexResponse;
    if (object.ddlindex !== undefined && object.ddlindex !== null) {
      message.ddlindex = Ddlindex.fromPartial(object.ddlindex);
    } else {
      message.ddlindex = undefined;
    }
    return message;
  },
};

const baseQueryAllDdlindexRequest: object = {};

export const QueryAllDdlindexRequest = {
  encode(
    message: QueryAllDdlindexRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllDdlindexRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllDdlindexRequest,
    } as QueryAllDdlindexRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDdlindexRequest {
    const message = {
      ...baseQueryAllDdlindexRequest,
    } as QueryAllDdlindexRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDdlindexRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllDdlindexRequest>
  ): QueryAllDdlindexRequest {
    const message = {
      ...baseQueryAllDdlindexRequest,
    } as QueryAllDdlindexRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllDdlindexResponse: object = {};

export const QueryAllDdlindexResponse = {
  encode(
    message: QueryAllDdlindexResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.ddlindex) {
      Ddlindex.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllDdlindexResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllDdlindexResponse,
    } as QueryAllDdlindexResponse;
    message.ddlindex = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ddlindex.push(Ddlindex.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllDdlindexResponse {
    const message = {
      ...baseQueryAllDdlindexResponse,
    } as QueryAllDdlindexResponse;
    message.ddlindex = [];
    if (object.ddlindex !== undefined && object.ddlindex !== null) {
      for (const e of object.ddlindex) {
        message.ddlindex.push(Ddlindex.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllDdlindexResponse): unknown {
    const obj: any = {};
    if (message.ddlindex) {
      obj.ddlindex = message.ddlindex.map((e) =>
        e ? Ddlindex.toJSON(e) : undefined
      );
    } else {
      obj.ddlindex = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllDdlindexResponse>
  ): QueryAllDdlindexResponse {
    const message = {
      ...baseQueryAllDdlindexResponse,
    } as QueryAllDdlindexResponse;
    message.ddlindex = [];
    if (object.ddlindex !== undefined && object.ddlindex !== null) {
      for (const e of object.ddlindex) {
        message.ddlindex.push(Ddlindex.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetQueryidsRequest: object = { index: "" };

export const QueryGetQueryidsRequest = {
  encode(
    message: QueryGetQueryidsRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetQueryidsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetQueryidsRequest,
    } as QueryGetQueryidsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetQueryidsRequest {
    const message = {
      ...baseQueryGetQueryidsRequest,
    } as QueryGetQueryidsRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    return message;
  },

  toJSON(message: QueryGetQueryidsRequest): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetQueryidsRequest>
  ): QueryGetQueryidsRequest {
    const message = {
      ...baseQueryGetQueryidsRequest,
    } as QueryGetQueryidsRequest;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    return message;
  },
};

const baseQueryGetQueryidsResponse: object = {};

export const QueryGetQueryidsResponse = {
  encode(
    message: QueryGetQueryidsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.queryids !== undefined) {
      Queryids.encode(message.queryids, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetQueryidsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetQueryidsResponse,
    } as QueryGetQueryidsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.queryids = Queryids.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetQueryidsResponse {
    const message = {
      ...baseQueryGetQueryidsResponse,
    } as QueryGetQueryidsResponse;
    if (object.queryids !== undefined && object.queryids !== null) {
      message.queryids = Queryids.fromJSON(object.queryids);
    } else {
      message.queryids = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetQueryidsResponse): unknown {
    const obj: any = {};
    message.queryids !== undefined &&
      (obj.queryids = message.queryids
        ? Queryids.toJSON(message.queryids)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetQueryidsResponse>
  ): QueryGetQueryidsResponse {
    const message = {
      ...baseQueryGetQueryidsResponse,
    } as QueryGetQueryidsResponse;
    if (object.queryids !== undefined && object.queryids !== null) {
      message.queryids = Queryids.fromPartial(object.queryids);
    } else {
      message.queryids = undefined;
    }
    return message;
  },
};

const baseQueryAllQueryidsRequest: object = {};

export const QueryAllQueryidsRequest = {
  encode(
    message: QueryAllQueryidsRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllQueryidsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllQueryidsRequest,
    } as QueryAllQueryidsRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllQueryidsRequest {
    const message = {
      ...baseQueryAllQueryidsRequest,
    } as QueryAllQueryidsRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllQueryidsRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllQueryidsRequest>
  ): QueryAllQueryidsRequest {
    const message = {
      ...baseQueryAllQueryidsRequest,
    } as QueryAllQueryidsRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllQueryidsResponse: object = {};

export const QueryAllQueryidsResponse = {
  encode(
    message: QueryAllQueryidsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.queryids) {
      Queryids.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllQueryidsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllQueryidsResponse,
    } as QueryAllQueryidsResponse;
    message.queryids = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.queryids.push(Queryids.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllQueryidsResponse {
    const message = {
      ...baseQueryAllQueryidsResponse,
    } as QueryAllQueryidsResponse;
    message.queryids = [];
    if (object.queryids !== undefined && object.queryids !== null) {
      for (const e of object.queryids) {
        message.queryids.push(Queryids.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllQueryidsResponse): unknown {
    const obj: any = {};
    if (message.queryids) {
      obj.queryids = message.queryids.map((e) =>
        e ? Queryids.toJSON(e) : undefined
      );
    } else {
      obj.queryids = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllQueryidsResponse>
  ): QueryAllQueryidsResponse {
    const message = {
      ...baseQueryAllQueryidsResponse,
    } as QueryAllQueryidsResponse;
    message.queryids = [];
    if (object.queryids !== undefined && object.queryids !== null) {
      for (const e of object.queryids) {
        message.queryids.push(Queryids.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a Databases by index. */
  Databases(
    request: QueryGetDatabasesRequest
  ): Promise<QueryGetDatabasesResponse>;
  /** Queries a list of Databases items. */
  DatabasesAll(
    request: QueryAllDatabasesRequest
  ): Promise<QueryAllDatabasesResponse>;
  /** Queries a Ddl by index. */
  Ddl(request: QueryGetDdlRequest): Promise<QueryGetDdlResponse>;
  /** Queries a list of Ddl items. */
  DdlAll(request: QueryAllDdlRequest): Promise<QueryAllDdlResponse>;
  /** Queries a Ddlindex by index. */
  Ddlindex(request: QueryGetDdlindexRequest): Promise<QueryGetDdlindexResponse>;
  /** Queries a list of Ddlindex items. */
  DdlindexAll(
    request: QueryAllDdlindexRequest
  ): Promise<QueryAllDdlindexResponse>;
  /** Queries a Queryids by index. */
  Queryids(request: QueryGetQueryidsRequest): Promise<QueryGetQueryidsResponse>;
  /** Queries a list of Queryids items. */
  QueryidsAll(
    request: QueryAllQueryidsRequest
  ): Promise<QueryAllQueryidsResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "Params", data);
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  Databases(
    request: QueryGetDatabasesRequest
  ): Promise<QueryGetDatabasesResponse> {
    const data = QueryGetDatabasesRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "Databases", data);
    return promise.then((data) =>
      QueryGetDatabasesResponse.decode(new Reader(data))
    );
  }

  DatabasesAll(
    request: QueryAllDatabasesRequest
  ): Promise<QueryAllDatabasesResponse> {
    const data = QueryAllDatabasesRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "DatabasesAll", data);
    return promise.then((data) =>
      QueryAllDatabasesResponse.decode(new Reader(data))
    );
  }

  Ddl(request: QueryGetDdlRequest): Promise<QueryGetDdlResponse> {
    const data = QueryGetDdlRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "Ddl", data);
    return promise.then((data) => QueryGetDdlResponse.decode(new Reader(data)));
  }

  DdlAll(request: QueryAllDdlRequest): Promise<QueryAllDdlResponse> {
    const data = QueryAllDdlRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "DdlAll", data);
    return promise.then((data) => QueryAllDdlResponse.decode(new Reader(data)));
  }

  Ddlindex(
    request: QueryGetDdlindexRequest
  ): Promise<QueryGetDdlindexResponse> {
    const data = QueryGetDdlindexRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "Ddlindex", data);
    return promise.then((data) =>
      QueryGetDdlindexResponse.decode(new Reader(data))
    );
  }

  DdlindexAll(
    request: QueryAllDdlindexRequest
  ): Promise<QueryAllDdlindexResponse> {
    const data = QueryAllDdlindexRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "DdlindexAll", data);
    return promise.then((data) =>
      QueryAllDdlindexResponse.decode(new Reader(data))
    );
  }

  Queryids(
    request: QueryGetQueryidsRequest
  ): Promise<QueryGetQueryidsResponse> {
    const data = QueryGetQueryidsRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "Queryids", data);
    return promise.then((data) =>
      QueryGetQueryidsResponse.decode(new Reader(data))
    );
  }

  QueryidsAll(
    request: QueryAllQueryidsRequest
  ): Promise<QueryAllQueryidsResponse> {
    const data = QueryAllQueryidsRequest.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Query", "QueryidsAll", data);
    return promise.then((data) =>
      QueryAllQueryidsResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;
