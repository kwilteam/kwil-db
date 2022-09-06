/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "kwil.kwil";

export interface Ddl {
  index: string;
  statement: string;
  position: number;
  final: boolean;
}

const baseDdl: object = { index: "", statement: "", position: 0, final: false };

export const Ddl = {
  encode(message: Ddl, writer: Writer = Writer.create()): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    if (message.statement !== "") {
      writer.uint32(18).string(message.statement);
    }
    if (message.position !== 0) {
      writer.uint32(24).int64(message.position);
    }
    if (message.final === true) {
      writer.uint32(32).bool(message.final);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Ddl {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseDdl } as Ddl;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        case 2:
          message.statement = reader.string();
          break;
        case 3:
          message.position = longToNumber(reader.int64() as Long);
          break;
        case 4:
          message.final = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Ddl {
    const message = { ...baseDdl } as Ddl;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    if (object.statement !== undefined && object.statement !== null) {
      message.statement = String(object.statement);
    } else {
      message.statement = "";
    }
    if (object.position !== undefined && object.position !== null) {
      message.position = Number(object.position);
    } else {
      message.position = 0;
    }
    if (object.final !== undefined && object.final !== null) {
      message.final = Boolean(object.final);
    } else {
      message.final = false;
    }
    return message;
  },

  toJSON(message: Ddl): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    message.statement !== undefined && (obj.statement = message.statement);
    message.position !== undefined && (obj.position = message.position);
    message.final !== undefined && (obj.final = message.final);
    return obj;
  },

  fromPartial(object: DeepPartial<Ddl>): Ddl {
    const message = { ...baseDdl } as Ddl;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    if (object.statement !== undefined && object.statement !== null) {
      message.statement = object.statement;
    } else {
      message.statement = "";
    }
    if (object.position !== undefined && object.position !== null) {
      message.position = object.position;
    } else {
      message.position = 0;
    }
    if (object.final !== undefined && object.final !== null) {
      message.final = object.final;
    } else {
      message.final = false;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

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

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
