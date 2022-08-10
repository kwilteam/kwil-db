import { txClient, queryClient, MissingWalletError , registry} from './module'

import { Databases } from "./module/types/kwil/databases"
import { Ddl } from "./module/types/kwil/ddl"
import { Ddlindex } from "./module/types/kwil/ddlindex"
import { Params } from "./module/types/kwil/params"
import { Queryids } from "./module/types/kwil/queryids"


export { Databases, Ddl, Ddlindex, Params, Queryids };

async function initTxClient(vuexGetters) {
	return await txClient(vuexGetters['common/wallet/signer'], {
		addr: vuexGetters['common/env/apiTendermint']
	})
}

async function initQueryClient(vuexGetters) {
	return await queryClient({
		addr: vuexGetters['common/env/apiCosmos']
	})
}

function mergeResults(value, next_values) {
	for (let prop of Object.keys(next_values)) {
		if (Array.isArray(next_values[prop])) {
			value[prop]=[...value[prop], ...next_values[prop]]
		}else{
			value[prop]=next_values[prop]
		}
	}
	return value
}

function getStructure(template) {
	let structure = { fields: [] }
	for (const [key, value] of Object.entries(template)) {
		let field: any = {}
		field.name = key
		field.type = typeof value
		structure.fields.push(field)
	}
	return structure
}

const getDefaultState = () => {
	return {
				Params: {},
				Databases: {},
				DatabasesAll: {},
				Ddl: {},
				DdlAll: {},
				Ddlindex: {},
				DdlindexAll: {},
				Queryids: {},
				QueryidsAll: {},
				
				_Structure: {
						Databases: getStructure(Databases.fromPartial({})),
						Ddl: getStructure(Ddl.fromPartial({})),
						Ddlindex: getStructure(Ddlindex.fromPartial({})),
						Params: getStructure(Params.fromPartial({})),
						Queryids: getStructure(Queryids.fromPartial({})),
						
		},
		_Registry: registry,
		_Subscriptions: new Set(),
	}
}

// initial state
const state = getDefaultState()

export default {
	namespaced: true,
	state,
	mutations: {
		RESET_STATE(state) {
			Object.assign(state, getDefaultState())
		},
		QUERY(state, { query, key, value }) {
			state[query][JSON.stringify(key)] = value
		},
		SUBSCRIBE(state, subscription) {
			state._Subscriptions.add(JSON.stringify(subscription))
		},
		UNSUBSCRIBE(state, subscription) {
			state._Subscriptions.delete(JSON.stringify(subscription))
		}
	},
	getters: {
				getParams: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Params[JSON.stringify(params)] ?? {}
		},
				getDatabases: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Databases[JSON.stringify(params)] ?? {}
		},
				getDatabasesAll: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DatabasesAll[JSON.stringify(params)] ?? {}
		},
				getDdl: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Ddl[JSON.stringify(params)] ?? {}
		},
				getDdlAll: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DdlAll[JSON.stringify(params)] ?? {}
		},
				getDdlindex: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Ddlindex[JSON.stringify(params)] ?? {}
		},
				getDdlindexAll: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.DdlindexAll[JSON.stringify(params)] ?? {}
		},
				getQueryids: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.Queryids[JSON.stringify(params)] ?? {}
		},
				getQueryidsAll: (state) => (params = { params: {}}) => {
					if (!(<any> params).query) {
						(<any> params).query=null
					}
			return state.QueryidsAll[JSON.stringify(params)] ?? {}
		},
				
		getTypeStructure: (state) => (type) => {
			return state._Structure[type].fields
		},
		getRegistry: (state) => {
			return state._Registry
		}
	},
	actions: {
		init({ dispatch, rootGetters }) {
			console.log('Vuex module: kwil.kwil initialized!')
			if (rootGetters['common/env/client']) {
				rootGetters['common/env/client'].on('newblock', () => {
					dispatch('StoreUpdate')
				})
			}
		},
		resetState({ commit }) {
			commit('RESET_STATE')
		},
		unsubscribe({ commit }, subscription) {
			commit('UNSUBSCRIBE', subscription)
		},
		async StoreUpdate({ state, dispatch }) {
			state._Subscriptions.forEach(async (subscription) => {
				try {
					const sub=JSON.parse(subscription)
					await dispatch(sub.action, sub.payload)
				}catch(e) {
					throw new Error('Subscriptions: ' + e.message)
				}
			})
		},
		
		
		
		 		
		
		
		async QueryParams({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryParams()).data
				
					
				commit('QUERY', { query: 'Params', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryParams', payload: { options: { all }, params: {...key},query }})
				return getters['getParams']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryParams API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDatabases({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDatabases( key.index)).data
				
					
				commit('QUERY', { query: 'Databases', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDatabases', payload: { options: { all }, params: {...key},query }})
				return getters['getDatabases']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDatabases API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDatabasesAll({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDatabasesAll(query)).data
				
					
				while (all && (<any> value).pagination && (<any> value).pagination.next_key!=null) {
					let next_values=(await queryClient.queryDatabasesAll({...query, 'pagination.key':(<any> value).pagination.next_key})).data
					value = mergeResults(value, next_values);
				}
				commit('QUERY', { query: 'DatabasesAll', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDatabasesAll', payload: { options: { all }, params: {...key},query }})
				return getters['getDatabasesAll']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDatabasesAll API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDdl({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDdl( key.index)).data
				
					
				commit('QUERY', { query: 'Ddl', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDdl', payload: { options: { all }, params: {...key},query }})
				return getters['getDdl']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDdl API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDdlAll({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDdlAll(query)).data
				
					
				while (all && (<any> value).pagination && (<any> value).pagination.next_key!=null) {
					let next_values=(await queryClient.queryDdlAll({...query, 'pagination.key':(<any> value).pagination.next_key})).data
					value = mergeResults(value, next_values);
				}
				commit('QUERY', { query: 'DdlAll', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDdlAll', payload: { options: { all }, params: {...key},query }})
				return getters['getDdlAll']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDdlAll API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDdlindex({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDdlindex( key.index)).data
				
					
				commit('QUERY', { query: 'Ddlindex', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDdlindex', payload: { options: { all }, params: {...key},query }})
				return getters['getDdlindex']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDdlindex API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryDdlindexAll({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryDdlindexAll(query)).data
				
					
				while (all && (<any> value).pagination && (<any> value).pagination.next_key!=null) {
					let next_values=(await queryClient.queryDdlindexAll({...query, 'pagination.key':(<any> value).pagination.next_key})).data
					value = mergeResults(value, next_values);
				}
				commit('QUERY', { query: 'DdlindexAll', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryDdlindexAll', payload: { options: { all }, params: {...key},query }})
				return getters['getDdlindexAll']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryDdlindexAll API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryQueryids({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryQueryids( key.index)).data
				
					
				commit('QUERY', { query: 'Queryids', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryQueryids', payload: { options: { all }, params: {...key},query }})
				return getters['getQueryids']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryQueryids API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		
		
		 		
		
		
		async QueryQueryidsAll({ commit, rootGetters, getters }, { options: { subscribe, all} = { subscribe:false, all:false}, params, query=null }) {
			try {
				const key = params ?? {};
				const queryClient=await initQueryClient(rootGetters)
				let value= (await queryClient.queryQueryidsAll(query)).data
				
					
				while (all && (<any> value).pagination && (<any> value).pagination.next_key!=null) {
					let next_values=(await queryClient.queryQueryidsAll({...query, 'pagination.key':(<any> value).pagination.next_key})).data
					value = mergeResults(value, next_values);
				}
				commit('QUERY', { query: 'QueryidsAll', key: { params: {...key}, query}, value })
				if (subscribe) commit('SUBSCRIBE', { action: 'QueryQueryidsAll', payload: { options: { all }, params: {...key},query }})
				return getters['getQueryidsAll']( { params: {...key}, query}) ?? {}
			} catch (e) {
				throw new Error('QueryClient:QueryQueryidsAll API Node Unavailable. Could not perform query: ' + e.message)
				
			}
		},
		
		
		async sendMsgDatabaseWrite({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDatabaseWrite(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDatabaseWrite:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgDatabaseWrite:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgCreateDatabase({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgCreateDatabase(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgCreateDatabase:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgCreateDatabase:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgDefineQuery({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDefineQuery(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDefineQuery:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgDefineQuery:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		async sendMsgDDL({ rootGetters }, { value, fee = [], memo = '' }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDDL(value)
				const result = await txClient.signAndBroadcast([msg], {fee: { amount: fee, 
	gas: "200000" }, memo})
				return result
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDDL:Init Could not initialize signing client. Wallet is required.')
				}else{
					throw new Error('TxClient:MsgDDL:Send Could not broadcast Tx: '+ e.message)
				}
			}
		},
		
		async MsgDatabaseWrite({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDatabaseWrite(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDatabaseWrite:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgDatabaseWrite:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgCreateDatabase({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgCreateDatabase(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgCreateDatabase:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgCreateDatabase:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgDefineQuery({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDefineQuery(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDefineQuery:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgDefineQuery:Create Could not create message: ' + e.message)
				}
			}
		},
		async MsgDDL({ rootGetters }, { value }) {
			try {
				const txClient=await initTxClient(rootGetters)
				const msg = await txClient.msgDDL(value)
				return msg
			} catch (e) {
				if (e == MissingWalletError) {
					throw new Error('TxClient:MsgDDL:Init Could not initialize signing client. Wallet is required.')
				} else{
					throw new Error('TxClient:MsgDDL:Create Could not create message: ' + e.message)
				}
			}
		},
		
	}
}
