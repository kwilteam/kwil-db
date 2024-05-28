package specifications

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExecuteTypesSpecification(ctx context.Context, t *testing.T, execute ProcedureDSL, testNil bool) {
	db := SchemaLoader.Load(t, TypesDB)

	res, err := execute.DeployDatabase(ctx, db)
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	// procedures
	t.Log("Testing types procedures")

	wg := sync.WaitGroup{}

	typeFuncs := []typeFunc{
		testUUID,
		testDecimal,
		testUint256,
		testText,
		testBool,
		testBlob,
		testInts,
	}
	if testNil {
		typeFuncs = append(typeFuncs, testNils)
	}

	// concurrent testing here significantly speeds up the test

	for _, proc := range typeFuncs {
		wg.Add(1)
		go func(fn typeFunc) {
			defer wg.Done()
			fn(ctx, t, execute, execute.DBID(db.Name), "proc")
		}(proc)
	}

	wg.Wait()

	// delete
	res, err = execute.Execute(ctx, execute.DBID(db.Name), "delete_all", []any{})
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	// actions
	t.Log("Testing types actions")

	for _, proc := range typeFuncs {
		wg.Add(1)
		go func(fn typeFunc) {
			defer wg.Done()
			fn(ctx, t, execute, execute.DBID(db.Name), "act")
		}(proc)
	}

	wg.Wait()
}

// signMu prevents concurrent signing
var signMu sync.Mutex

type typeFunc func(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string)

func testUUID(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store
	uuid := types.NewUUIDV5([]byte("1"))
	uuidArr := []any{types.NewUUIDV5([]byte("2")), types.NewUUIDV5([]byte("3"))}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_uuids", []any{uuid, uuidArr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_uuids", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		assert.Equal(t, uuid.String(), fmt.Sprint(row["id"]))

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, uuidArr[i].(*types.UUID).String(), fmt.Sprint(id))
		}
	}
	assert.Equal(t, 1, count)
}

func testDecimal(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	dec, err := decimal.NewFromString("123.456")
	require.NoError(t, err)

	dec2, err := decimal.NewFromString("789.012")
	require.NoError(t, err)
	dec3, err := decimal.NewFromString("345.678")
	require.NoError(t, err)

	decArr := []any{dec2, dec3}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_decimals", []any{dec, decArr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_decimals", []any{})
	require.NoError(t, err)

	count := 0

	for result.Next() {
		count++
		row := result.Record()
		id := row["id"].(string)
		// trim off extra precision
		// this is not ideal, but unfortunately necessary since we json marshal results and have several drivers.
		assert.Equal(t, dec.String(), strings.TrimRight(fmt.Sprint(id), "0"))

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, decArr[i].(*decimal.Decimal).String(), strings.TrimRight(fmt.Sprint(id), "0")) // trim off extra precision
		}
	}
}

func testUint256(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store
	uint256, err := types.Uint256FromString("115792089237316195423570985008687907853269984665640564039457584007913129639935") // max uint256
	require.NoError(t, err)
	uint256Arr := []any{types.Uint256FromInt(456), types.Uint256FromInt(789)}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_uint256s", []any{uint256, uint256Arr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_uint256s", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		// uint256 values should be convertable to int64s here
		assert.Equal(t, uint256.String(), fmt.Sprint(row["id"]))

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, uint256Arr[i].(*types.Uint256).String(), fmt.Sprint(id))
		}
	}
	assert.Equal(t, 1, count)
}

func testText(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store
	text := "hello"
	textArr := []any{"world", "foo"}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_texts", []any{text, textArr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_texts", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		assert.Equal(t, text, row["id"])

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, textArr[i].(string), id)
		}
	}
	assert.Equal(t, 1, count)
}

func testBool(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store
	boolean := true
	booleanArr := []any{false, true}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_bools", []any{boolean, booleanArr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_bools", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		assert.Equal(t, fmt.Sprint(boolean), fmt.Sprint(row["id"]))

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, fmt.Sprint(booleanArr[i]), fmt.Sprint(id))
		}
	}
	assert.Equal(t, 1, count)
}

func testBlob(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store
	blob := []byte{1, 2, 3}
	blobArr := []any{[]byte{4, 5, 6}, []byte{7, 8, 9}}

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_blobs", []any{blob, blobArr})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_blobs", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		// we use base64 since we are reading this after querying with jsonrpc
		assert.Equal(t, base64.StdEncoding.EncodeToString(blob), row["id"])

		ids := row["arr"]
		for i, id := range ids.([]any) {
			assert.Equal(t, base64.StdEncoding.EncodeToString(blobArr[i].([]byte)), id)
		}
	}
	assert.Equal(t, 1, count)
}

func testInts(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	// execute store

	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_ints", []any{int32(1), []any{int64(2), uint32(2)}})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_ints", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		assert.Equal(t, fmt.Sprint(1), fmt.Sprint(row["id"]))

		ids := row["arr"]
		for _, id := range ids.([]any) {
			assert.Equal(t, fmt.Sprint(2), fmt.Sprint(id))
		}
	}
	assert.Equal(t, 1, count)
}

func testNils(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid string, callType string) {
	/*
		the store method takes the following arguments:

		$text_s text
		$text_a text[]
		$int_s int
		$int_a int[]
		$bool_s bool
		$bool_a bool[]
		$blob_s blob
		$blob_a blob[]
		$decimal_s decimal(10,5)
		$decimal_a decimal(10,5)[]
		$uint256_s uint256
		$uint256_a uint256[]
		$uuid_s uuid
		$uuid_a uuid[]
	*/

	// execute store, one nil for each arg
	signMu.Lock()
	res, err := execute.Execute(ctx, dbid, callType+"_store_nils", []any{
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	})
	signMu.Unlock()
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	result, err := execute.Call(ctx, dbid, callType+"_get_nils", []any{})
	require.NoError(t, err)

	count := 0
	for result.Next() {
		count++
		row := result.Record()
		assert.Nil(t, row["text_s"])
		assert.Nil(t, row["text_a"])
		assert.Nil(t, row["int_s"])
		assert.Nil(t, row["int_a"])
		assert.Nil(t, row["bool_s"])
		assert.Nil(t, row["bool_a"])
		assert.Nil(t, row["blob_s"])
		assert.Nil(t, row["blob_a"])
		assert.Nil(t, row["decimal_s"])
		assert.Nil(t, row["decimal_a"])
		assert.Nil(t, row["uint256_s"])
		assert.Nil(t, row["uint256_a"])
		assert.Nil(t, row["uuid_s"])
		assert.Nil(t, row["uuid_a"])
	}
	assert.Equal(t, 1, count)
}
