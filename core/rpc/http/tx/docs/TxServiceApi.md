# {{classname}}

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**TxServiceBroadcast**](TxServiceApi.md#TxServiceBroadcast) | **Post** /api/v1/broadcast | 
[**TxServiceCall**](TxServiceApi.md#TxServiceCall) | **Post** /api/v1/call | 
[**TxServiceChainInfo**](TxServiceApi.md#TxServiceChainInfo) | **Get** /api/v1/chain_info | 
[**TxServiceEstimatePrice**](TxServiceApi.md#TxServiceEstimatePrice) | **Post** /api/v1/estimate_price | 
[**TxServiceGetAccount**](TxServiceApi.md#TxServiceGetAccount) | **Get** /api/v1/accounts/{identifier} | 
[**TxServiceGetConfig**](TxServiceApi.md#TxServiceGetConfig) | **Get** /api/v1/config | 
[**TxServiceGetSchema**](TxServiceApi.md#TxServiceGetSchema) | **Get** /api/v1/databases/{dbid}/schema | 
[**TxServiceListDatabases**](TxServiceApi.md#TxServiceListDatabases) | **Get** /api/v1/{owner}/databases | 
[**TxServicePing**](TxServiceApi.md#TxServicePing) | **Get** /api/v1/ping | 
[**TxServiceQuery**](TxServiceApi.md#TxServiceQuery) | **Post** /api/v1/query | 
[**TxServiceTxQuery**](TxServiceApi.md#TxServiceTxQuery) | **Post** /api/v1/tx_query | 

# **TxServiceBroadcast**
> TxBroadcastResponse TxServiceBroadcast(ctx, body)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**TxBroadcastRequest**](TxBroadcastRequest.md)|  | 

### Return type

[**TxBroadcastResponse**](txBroadcastResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceCall**
> TxCallResponse TxServiceCall(ctx, body)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**TxCallRequest**](TxCallRequest.md)|  | 

### Return type

[**TxCallResponse**](txCallResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceChainInfo**
> TxChainInfoResponse TxServiceChainInfo(ctx, )


### Required Parameters
This endpoint does not need any parameter.

### Return type

[**TxChainInfoResponse**](txChainInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceEstimatePrice**
> TxEstimatePriceResponse TxServiceEstimatePrice(ctx, body)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**TxEstimatePriceRequest**](TxEstimatePriceRequest.md)|  | 

### Return type

[**TxEstimatePriceResponse**](txEstimatePriceResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceGetAccount**
> TxGetAccountResponse TxServiceGetAccount(ctx, identifier, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**|  | 
 **optional** | ***TxServiceApiTxServiceGetAccountOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TxServiceApiTxServiceGetAccountOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **optional.String**| Mapped to URL query parameter &#x60;status&#x60;. | [default to latest]

### Return type

[**TxGetAccountResponse**](txGetAccountResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceGetConfig**
> TxGetConfigResponse TxServiceGetConfig(ctx, )


### Required Parameters
This endpoint does not need any parameter.

### Return type

[**TxGetConfigResponse**](txGetConfigResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceGetSchema**
> TxGetSchemaResponse TxServiceGetSchema(ctx, dbid)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **dbid** | **string**|  | 

### Return type

[**TxGetSchemaResponse**](txGetSchemaResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceListDatabases**
> TxListDatabasesResponse TxServiceListDatabases(ctx, owner)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **owner** | **string**|  | 

### Return type

[**TxListDatabasesResponse**](txListDatabasesResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServicePing**
> TxPingResponse TxServicePing(ctx, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***TxServiceApiTxServicePingOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a TxServiceApiTxServicePingOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **message** | **optional.String**|  | 

### Return type

[**TxPingResponse**](txPingResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceQuery**
> TxQueryResponse TxServiceQuery(ctx, body)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**TxQueryRequest**](TxQueryRequest.md)|  | 

### Return type

[**TxQueryResponse**](txQueryResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **TxServiceTxQuery**
> TxTxQueryResponse TxServiceTxQuery(ctx, body)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**TxTxQueryRequest**](TxTxQueryRequest.md)|  | 

### Return type

[**TxTxQueryResponse**](txTxQueryResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

