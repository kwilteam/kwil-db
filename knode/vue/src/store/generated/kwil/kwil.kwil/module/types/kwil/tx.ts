/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";

export const protobufPackage = "kwil.kwil";

export interface MsgDatabaseWrite {
  creator: string;
  database: string;
  parQuer: string;
  data: string;
}

export interface MsgDatabaseWriteResponse {
  ret: string;
}

export interface MsgCreateDatabase {
  creator: string;
  seed: string;
}

export interface MsgCreateDatabaseResponse {
  id: string;
}

export interface MsgDDL {
  creator: string;
  dbid: string;
  ddl: string;
}

export interface MsgDDLResponse {}

export interface MsgDefineQuery {
  creator: string;
  dbId: string;
  parQuer: string;
  publicity: boolean;
}

export interface MsgDefineQueryResponse {
  id: string;
}

const baseMsgDatabaseWrite: object = {
  creator: "",
  database: "",
  parQuer: "",
  data: "",
};

export const MsgDatabaseWrite = {
  encode(message: MsgDatabaseWrite, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.database !== "") {
      writer.uint32(18).string(message.database);
    }
    if (message.parQuer !== "") {
      writer.uint32(26).string(message.parQuer);
    }
    if (message.data !== "") {
      writer.uint32(34).string(message.data);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDatabaseWrite {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDatabaseWrite } as MsgDatabaseWrite;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.database = reader.string();
          break;
        case 3:
          message.parQuer = reader.string();
          break;
        case 4:
          message.data = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDatabaseWrite {
    const message = { ...baseMsgDatabaseWrite } as MsgDatabaseWrite;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.database !== undefined && object.database !== null) {
      message.database = String(object.database);
    } else {
      message.database = "";
    }
    if (object.parQuer !== undefined && object.parQuer !== null) {
      message.parQuer = String(object.parQuer);
    } else {
      message.parQuer = "";
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = String(object.data);
    } else {
      message.data = "";
    }
    return message;
  },

  toJSON(message: MsgDatabaseWrite): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.database !== undefined && (obj.database = message.database);
    message.parQuer !== undefined && (obj.parQuer = message.parQuer);
    message.data !== undefined && (obj.data = message.data);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDatabaseWrite>): MsgDatabaseWrite {
    const message = { ...baseMsgDatabaseWrite } as MsgDatabaseWrite;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.database !== undefined && object.database !== null) {
      message.database = object.database;
    } else {
      message.database = "";
    }
    if (object.parQuer !== undefined && object.parQuer !== null) {
      message.parQuer = object.parQuer;
    } else {
      message.parQuer = "";
    }
    if (object.data !== undefined && object.data !== null) {
      message.data = object.data;
    } else {
      message.data = "";
    }
    return message;
  },
};

const baseMsgDatabaseWriteResponse: object = { ret: "" };

export const MsgDatabaseWriteResponse = {
  encode(
    message: MsgDatabaseWriteResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.ret !== "") {
      writer.uint32(10).string(message.ret);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgDatabaseWriteResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgDatabaseWriteResponse,
    } as MsgDatabaseWriteResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.ret = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDatabaseWriteResponse {
    const message = {
      ...baseMsgDatabaseWriteResponse,
    } as MsgDatabaseWriteResponse;
    if (object.ret !== undefined && object.ret !== null) {
      message.ret = String(object.ret);
    } else {
      message.ret = "";
    }
    return message;
  },

  toJSON(message: MsgDatabaseWriteResponse): unknown {
    const obj: any = {};
    message.ret !== undefined && (obj.ret = message.ret);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgDatabaseWriteResponse>
  ): MsgDatabaseWriteResponse {
    const message = {
      ...baseMsgDatabaseWriteResponse,
    } as MsgDatabaseWriteResponse;
    if (object.ret !== undefined && object.ret !== null) {
      message.ret = object.ret;
    } else {
      message.ret = "";
    }
    return message;
  },
};

const baseMsgCreateDatabase: object = { creator: "", seed: "" };

export const MsgCreateDatabase = {
  encode(message: MsgCreateDatabase, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.seed !== "") {
      writer.uint32(18).string(message.seed);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateDatabase {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateDatabase } as MsgCreateDatabase;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.seed = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateDatabase {
    const message = { ...baseMsgCreateDatabase } as MsgCreateDatabase;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.seed !== undefined && object.seed !== null) {
      message.seed = String(object.seed);
    } else {
      message.seed = "";
    }
    return message;
  },

  toJSON(message: MsgCreateDatabase): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.seed !== undefined && (obj.seed = message.seed);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateDatabase>): MsgCreateDatabase {
    const message = { ...baseMsgCreateDatabase } as MsgCreateDatabase;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.seed !== undefined && object.seed !== null) {
      message.seed = object.seed;
    } else {
      message.seed = "";
    }
    return message;
  },
};

const baseMsgCreateDatabaseResponse: object = { id: "" };

export const MsgCreateDatabaseResponse = {
  encode(
    message: MsgCreateDatabaseResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgCreateDatabaseResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgCreateDatabaseResponse,
    } as MsgCreateDatabaseResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateDatabaseResponse {
    const message = {
      ...baseMsgCreateDatabaseResponse,
    } as MsgCreateDatabaseResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = String(object.id);
    } else {
      message.id = "";
    }
    return message;
  },

  toJSON(message: MsgCreateDatabaseResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgCreateDatabaseResponse>
  ): MsgCreateDatabaseResponse {
    const message = {
      ...baseMsgCreateDatabaseResponse,
    } as MsgCreateDatabaseResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = "";
    }
    return message;
  },
};

const baseMsgDDL: object = { creator: "", dbid: "", ddl: "" };

export const MsgDDL = {
  encode(message: MsgDDL, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.dbid !== "") {
      writer.uint32(18).string(message.dbid);
    }
    if (message.ddl !== "") {
      writer.uint32(26).string(message.ddl);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDDL {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDDL } as MsgDDL;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.dbid = reader.string();
          break;
        case 3:
          message.ddl = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDDL {
    const message = { ...baseMsgDDL } as MsgDDL;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.dbid !== undefined && object.dbid !== null) {
      message.dbid = String(object.dbid);
    } else {
      message.dbid = "";
    }
    if (object.ddl !== undefined && object.ddl !== null) {
      message.ddl = String(object.ddl);
    } else {
      message.ddl = "";
    }
    return message;
  },

  toJSON(message: MsgDDL): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.dbid !== undefined && (obj.dbid = message.dbid);
    message.ddl !== undefined && (obj.ddl = message.ddl);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDDL>): MsgDDL {
    const message = { ...baseMsgDDL } as MsgDDL;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.dbid !== undefined && object.dbid !== null) {
      message.dbid = object.dbid;
    } else {
      message.dbid = "";
    }
    if (object.ddl !== undefined && object.ddl !== null) {
      message.ddl = object.ddl;
    } else {
      message.ddl = "";
    }
    return message;
  },
};

const baseMsgDDLResponse: object = {};

export const MsgDDLResponse = {
  encode(_: MsgDDLResponse, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDDLResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDDLResponse } as MsgDDLResponse;
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

  fromJSON(_: any): MsgDDLResponse {
    const message = { ...baseMsgDDLResponse } as MsgDDLResponse;
    return message;
  },

  toJSON(_: MsgDDLResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<MsgDDLResponse>): MsgDDLResponse {
    const message = { ...baseMsgDDLResponse } as MsgDDLResponse;
    return message;
  },
};

const baseMsgDefineQuery: object = {
  creator: "",
  dbId: "",
  parQuer: "",
  publicity: false,
};

export const MsgDefineQuery = {
  encode(message: MsgDefineQuery, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.dbId !== "") {
      writer.uint32(18).string(message.dbId);
    }
    if (message.parQuer !== "") {
      writer.uint32(26).string(message.parQuer);
    }
    if (message.publicity === true) {
      writer.uint32(32).bool(message.publicity);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDefineQuery {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDefineQuery } as MsgDefineQuery;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.dbId = reader.string();
          break;
        case 3:
          message.parQuer = reader.string();
          break;
        case 4:
          message.publicity = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDefineQuery {
    const message = { ...baseMsgDefineQuery } as MsgDefineQuery;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.dbId !== undefined && object.dbId !== null) {
      message.dbId = String(object.dbId);
    } else {
      message.dbId = "";
    }
    if (object.parQuer !== undefined && object.parQuer !== null) {
      message.parQuer = String(object.parQuer);
    } else {
      message.parQuer = "";
    }
    if (object.publicity !== undefined && object.publicity !== null) {
      message.publicity = Boolean(object.publicity);
    } else {
      message.publicity = false;
    }
    return message;
  },

  toJSON(message: MsgDefineQuery): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.dbId !== undefined && (obj.dbId = message.dbId);
    message.parQuer !== undefined && (obj.parQuer = message.parQuer);
    message.publicity !== undefined && (obj.publicity = message.publicity);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgDefineQuery>): MsgDefineQuery {
    const message = { ...baseMsgDefineQuery } as MsgDefineQuery;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.dbId !== undefined && object.dbId !== null) {
      message.dbId = object.dbId;
    } else {
      message.dbId = "";
    }
    if (object.parQuer !== undefined && object.parQuer !== null) {
      message.parQuer = object.parQuer;
    } else {
      message.parQuer = "";
    }
    if (object.publicity !== undefined && object.publicity !== null) {
      message.publicity = object.publicity;
    } else {
      message.publicity = false;
    }
    return message;
  },
};

const baseMsgDefineQueryResponse: object = { id: "" };

export const MsgDefineQueryResponse = {
  encode(
    message: MsgDefineQueryResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== "") {
      writer.uint32(10).string(message.id);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgDefineQueryResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgDefineQueryResponse } as MsgDefineQueryResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgDefineQueryResponse {
    const message = { ...baseMsgDefineQueryResponse } as MsgDefineQueryResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = String(object.id);
    } else {
      message.id = "";
    }
    return message;
  },

  toJSON(message: MsgDefineQueryResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgDefineQueryResponse>
  ): MsgDefineQueryResponse {
    const message = { ...baseMsgDefineQueryResponse } as MsgDefineQueryResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = "";
    }
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  DatabaseWrite(request: MsgDatabaseWrite): Promise<MsgDatabaseWriteResponse>;
  CreateDatabase(
    request: MsgCreateDatabase
  ): Promise<MsgCreateDatabaseResponse>;
  DDL(request: MsgDDL): Promise<MsgDDLResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  DefineQuery(request: MsgDefineQuery): Promise<MsgDefineQueryResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  DatabaseWrite(request: MsgDatabaseWrite): Promise<MsgDatabaseWriteResponse> {
    const data = MsgDatabaseWrite.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Msg", "DatabaseWrite", data);
    return promise.then((data) =>
      MsgDatabaseWriteResponse.decode(new Reader(data))
    );
  }

  CreateDatabase(
    request: MsgCreateDatabase
  ): Promise<MsgCreateDatabaseResponse> {
    const data = MsgCreateDatabase.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Msg", "CreateDatabase", data);
    return promise.then((data) =>
      MsgCreateDatabaseResponse.decode(new Reader(data))
    );
  }

  DDL(request: MsgDDL): Promise<MsgDDLResponse> {
    const data = MsgDDL.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Msg", "DDL", data);
    return promise.then((data) => MsgDDLResponse.decode(new Reader(data)));
  }

  DefineQuery(request: MsgDefineQuery): Promise<MsgDefineQueryResponse> {
    const data = MsgDefineQuery.encode(request).finish();
    const promise = this.rpc.request("kwil.kwil.Msg", "DefineQuery", data);
    return promise.then((data) =>
      MsgDefineQueryResponse.decode(new Reader(data))
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
