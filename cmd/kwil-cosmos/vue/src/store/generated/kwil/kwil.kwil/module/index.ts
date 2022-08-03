// THIS FILE IS GENERATED AUTOMATICALLY. DO NOT MODIFY.

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient } from "@cosmjs/stargate";
import { Registry, OfflineSigner, EncodeObject, DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { Api } from "./rest";
import { MsgDDL } from "./types/kwil/tx";
import { MsgCreateDatabase } from "./types/kwil/tx";
import { MsgDefineQuery } from "./types/kwil/tx";
import { MsgDatabaseWrite } from "./types/kwil/tx";


const types = [
  ["/kwil.kwil.MsgDDL", MsgDDL],
  ["/kwil.kwil.MsgCreateDatabase", MsgCreateDatabase],
  ["/kwil.kwil.MsgDefineQuery", MsgDefineQuery],
  ["/kwil.kwil.MsgDatabaseWrite", MsgDatabaseWrite],
  
];
export const MissingWalletError = new Error("wallet is required");

export const registry = new Registry(<any>types);

const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
}

interface SignAndBroadcastOptions {
  fee: StdFee,
  memo?: string
}

const txClient = async (wallet: OfflineSigner, { addr: addr }: TxClientOptions = { addr: "http://localhost:26657" }) => {
  if (!wallet) throw MissingWalletError;
  let client;
  if (addr) {
    client = await SigningStargateClient.connectWithSigner(addr, wallet, { registry });
  }else{
    client = await SigningStargateClient.offline( wallet, { registry });
  }
  const { address } = (await wallet.getAccounts())[0];

  return {
    signAndBroadcast: (msgs: EncodeObject[], { fee, memo }: SignAndBroadcastOptions = {fee: defaultFee, memo: ""}) => client.signAndBroadcast(address, msgs, fee,memo),
    msgDDL: (data: MsgDDL): EncodeObject => ({ typeUrl: "/kwil.kwil.MsgDDL", value: MsgDDL.fromPartial( data ) }),
    msgCreateDatabase: (data: MsgCreateDatabase): EncodeObject => ({ typeUrl: "/kwil.kwil.MsgCreateDatabase", value: MsgCreateDatabase.fromPartial( data ) }),
    msgDefineQuery: (data: MsgDefineQuery): EncodeObject => ({ typeUrl: "/kwil.kwil.MsgDefineQuery", value: MsgDefineQuery.fromPartial( data ) }),
    msgDatabaseWrite: (data: MsgDatabaseWrite): EncodeObject => ({ typeUrl: "/kwil.kwil.MsgDatabaseWrite", value: MsgDatabaseWrite.fromPartial( data ) }),
    
  };
};

interface QueryClientOptions {
  addr: string
}

const queryClient = async ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseUrl: addr });
};

export {
  txClient,
  queryClient,
};
