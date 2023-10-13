package testdata

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/kwilteam/kwil-db/internal/engine/types"
)

var (
	tbl  = GetFromJson[types.Table]
	proc = GetFromJson[types.Procedure]
)

var typeFieldMap = map[reflect.Type]string{
	reflect.TypeOf(types.Table{}):     "tables",
	reflect.TypeOf(types.Procedure{}): "actions",
}

func GetFromJson[T any](name string) T {
	concrete := *new(T)

	allOfType, err := GetFieldArr[T]()
	if err != nil {
		panic(err)
	}

	for _, val := range allOfType {
		ident, err := getIdentifier(val)
		if err != nil {
			panic(err)
		}

		if strings.EqualFold(ident, name) {
			return val
		}
	}

	panic(fmt.Errorf("could not find %s in %T", name, concrete))
}

func GetFieldArr[T any]() ([]T, error) {
	detectedKind := reflect.TypeOf(*new(T))
	fieldName, ok := typeFieldMap[detectedKind]
	if !ok {
		return nil, fmt.Errorf("unknown type %v", detectedKind)
	}

	val, _, _, err := jsonparser.Get(schemaFile, fieldName)
	if err != nil {
		return nil, err
	}

	var arr []T

	dec := json.NewDecoder(bytes.NewReader(val))
	err = dec.Decode(&arr)
	if err != nil {
		return nil, err
	}

	return arr, nil
}

func getIdentifier(v any) (string, error) {
	switch typ := v.(type) {
	case types.Table:
		return typ.Name, nil
	case types.Procedure:
		return typ.Name, nil
	default:
		return "", fmt.Errorf("unknown type: %T", v)
	}
}

var (
	Table_users     = tbl("users")
	Table_posts     = tbl("posts")
	Table_likes     = tbl("likes")
	Table_followers = tbl("followers")
)

var (
	Procedure_create_user          = proc("create_user")
	Procedure_update_user          = proc("update_user")
	Procedure_create_post          = proc("create_post")
	Procedure_delete_post          = proc("delete_post")
	Procedure_like_post            = proc("like_post")
	Procedure_unlike_post          = proc("unlike_post")
	Procedure_follow               = proc("follow")
	Procedure_unfollow             = proc("unfollow")
	Procedure_get_user_by_username = proc("get_user_by_username")
	Procedure_get_user_by_wallet   = proc("get_user_by_wallet")
	Procedure_get_feed             = proc("get_feed")
	Procedure_get_celebrity_feed   = proc("get_celebrity_feed")
)

// var (
// 	UsersTable = &types.Table{
// 		Name: "users",
// 		Columns: []*types.Column{
// 			{
// 				Name: "id",
// 				Type: types.INT,
// 				Attributes: []*types.Attribute{
// 					{
// 						Type: types.PRIMARY_KEY,
// 					},
// 				},
// 			},
// 			{
// 				Name: "username",
// 				Type: types.TEXT,
// 				Attributes: []*types.Attribute{
// 					{
// 						Type: types.NOT_NULL,
// 					},
// 					{
// 						Type: types.UNIQUE,
// 					},
// 					{
// 						Type:  types.MIN_LENGTH,
// 						Value: 3,
// 					},
// 					{
// 						Type:  types.MAX_LENGTH,
// 						Value: 32,
// 					},
// 				},
// 			},
// 			{
// 				Name: "age",
// 				Type: types.INT,
// 				Attributes: []*types.Attribute{
// 					{
// 						Type: types.NOT_NULL,
// 					},
// 					{
// 						Type:  types.MIN,
// 						Value: 13,
// 					},
// 					{
// 						Type:  types.MAX,
// 						Value: 200,
// 					},
// 				},
// 			},
// 			{
// 				Name: "wallet_address",
// 				Type: types.TEXT,
// 				Attributes: []*types.Attribute{
// 					{
// 						Type: types.NOT_NULL,
// 					},
// 					{
// 						Type: types.UNIQUE,
// 					},
// 				},
// 			},
// 		},
// 		Indexes: []*types.Index{
// 			{
// 				Name: "users_age_index",
// 				Columns: []string{
// 					"age",
// 				},
// 				Type: types.BTREE,
// 			},
// 		},
// 		ForeignKeys: []*types.ForeignKey{},
// 	}
// )
