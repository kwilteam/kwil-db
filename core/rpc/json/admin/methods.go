package adminjson

import (
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
)

const (
	MethodVersion       jsonrpc.Method = "admin.version"
	MethodStatus        jsonrpc.Method = "admin.status"
	MethodPeers         jsonrpc.Method = "admin.peers"
	MethodConfig        jsonrpc.Method = "admin.config"
	MethodValApprove    jsonrpc.Method = "admin.val_approve"
	MethodValJoin       jsonrpc.Method = "admin.val_join"
	MethodValRemove     jsonrpc.Method = "admin.val_remove"
	MethodValLeave      jsonrpc.Method = "admin.val_leave"
	MethodValJoinStatus jsonrpc.Method = "admin.val_join_status"
	MethodValList       jsonrpc.Method = "admin.val_list"
	MethodValListJoins  jsonrpc.Method = "admin.val_list_joins"
)
