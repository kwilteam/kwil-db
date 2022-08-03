/* eslint-disable */
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "kwil.kwil";

export interface Queryids {
  index: string;
  queryid: string;
  query: string;
  dbid: string;
  publicity: string;
}

const baseQueryids: object = {
  index: "",
  queryid: "",
  query: "",
  dbid: "",
  publicity: "",
};

export const Queryids = {
  encode(message: Queryids, writer: Writer = Writer.create()): Writer {
    if (message.index !== "") {
      writer.uint32(10).string(message.index);
    }
    if (message.queryid !== "") {
      writer.uint32(18).string(message.queryid);
    }
    if (message.query !== "") {
      writer.uint32(26).string(message.query);
    }
    if (message.dbid !== "") {
      writer.uint32(34).string(message.dbid);
    }
    if (message.publicity !== "") {
      writer.uint32(42).string(message.publicity);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Queryids {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryids } as Queryids;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.index = reader.string();
          break;
        case 2:
          message.queryid = reader.string();
          break;
        case 3:
          message.query = reader.string();
          break;
        case 4:
          message.dbid = reader.string();
          break;
        case 5:
          message.publicity = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Queryids {
    const message = { ...baseQueryids } as Queryids;
    if (object.index !== undefined && object.index !== null) {
      message.index = String(object.index);
    } else {
      message.index = "";
    }
    if (object.queryid !== undefined && object.queryid !== null) {
      message.queryid = String(object.queryid);
    } else {
      message.queryid = "";
    }
    if (object.query !== undefined && object.query !== null) {
      message.query = String(object.query);
    } else {
      message.query = "";
    }
    if (object.dbid !== undefined && object.dbid !== null) {
      message.dbid = String(object.dbid);
    } else {
      message.dbid = "";
    }
    if (object.publicity !== undefined && object.publicity !== null) {
      message.publicity = String(object.publicity);
    } else {
      message.publicity = "";
    }
    return message;
  },

  toJSON(message: Queryids): unknown {
    const obj: any = {};
    message.index !== undefined && (obj.index = message.index);
    message.queryid !== undefined && (obj.queryid = message.queryid);
    message.query !== undefined && (obj.query = message.query);
    message.dbid !== undefined && (obj.dbid = message.dbid);
    message.publicity !== undefined && (obj.publicity = message.publicity);
    return obj;
  },

  fromPartial(object: DeepPartial<Queryids>): Queryids {
    const message = { ...baseQueryids } as Queryids;
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = "";
    }
    if (object.queryid !== undefined && object.queryid !== null) {
      message.queryid = object.queryid;
    } else {
      message.queryid = "";
    }
    if (object.query !== undefined && object.query !== null) {
      message.query = object.query;
    } else {
      message.query = "";
    }
    if (object.dbid !== undefined && object.dbid !== null) {
      message.dbid = object.dbid;
    } else {
      message.dbid = "";
    }
    if (object.publicity !== undefined && object.publicity !== null) {
      message.publicity = object.publicity;
    } else {
      message.publicity = "";
    }
    return message;
  },
};

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
