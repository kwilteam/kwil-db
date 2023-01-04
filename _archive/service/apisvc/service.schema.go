package apisvc

import (
	"context"
	"fmt"
	"kwil/x/crypto"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
	"kwil/x/sqlx/cache"
	"kwil/x/sqlx/models"
	"strings"

	"github.com/cstockton/go-conv"
)

func (s *Service) DeploySchema(ctx context.Context, req *apipb.DeploySchemaRequest) (*apipb.DeploySchemaResponse, error) {

	// verify the tx
	tx := crypto.Tx{}
	tx.Convert(req.Tx)
	err := tx.Verify()
	if err != nil {
		return nil, err
	}

	p, err := s.p.GetPrice(pricing.Deploy)
	if err != nil {
		return nil, err
	}

	// parse fee
	fee, ok := parseBigInt(req.Tx.Fee)
	if !ok {
		return nil, fmt.Errorf("invalid fee")
	}

	// check price is enough
	if fee.Cmp(p) < 0 {
		return nil, fmt.Errorf("price is not enough")
	}

	// spend funds and then write data
	err = s.manager.Deposits.Spend(ctx, tx.Sender, tx.Fee)
	if err != nil {
		return nil, err
	}

	msgBody, err := Unmarshal[models.CreateDatabase](tx.Payload)
	if err != nil {
		return nil, err
	}

	db := &models.Database{}
	err = db.FromJSON(msgBody.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ddl: %w", err)
	}

	db.Clean()

	if db.Owner != strings.ToLower(tx.Sender) {
		return nil, fmt.Errorf("db owner must also be tx signer: %s != %s", db.Owner, tx.Sender)
	}

	err = s.manager.Deployment.Deploy(ctx, db)
	if err != nil {
		return nil, err
	}

	return &apipb.DeploySchemaResponse{
		Txid: tx.Id,
		Msg:  "success",
	}, nil
}

func (s *Service) DropSchema(ctx context.Context, req *apipb.DropSchemaRequest) (*apipb.DropSchemaResponse, error) {

	// verify the tx
	tx := crypto.Tx{}
	tx.Convert(req.Tx)
	err := tx.Verify()
	if err != nil {
		return nil, err
	}

	p, err := s.p.GetPrice(pricing.Delete)
	if err != nil {
		return nil, err
	}

	// parse fee
	fee, ok := parseBigInt(req.Tx.Fee)
	if !ok {
		return nil, fmt.Errorf("invalid fee")
	}

	// check price is enough
	if fee.Cmp(p) < 0 {
		return nil, fmt.Errorf("price is not enough")
	}

	// spend funds and then write data!

	err = s.manager.Deposits.Spend(ctx, tx.Sender, tx.Fee)
	if err != nil {
		return nil, err
	}

	body, err := Unmarshal[models.DropDatabase](req.Tx.Payload)
	if err != nil {
		return nil, err
	}

	// check if the owner is the same as the tx sender
	if !strings.EqualFold(body.Owner, tx.Sender) {
		return nil, fmt.Errorf("db owner must also be tx signer: %s != %s", body.Owner, tx.Sender)
	}

	dbName := strings.ToLower(body.Name + "_" + body.Owner)

	err = s.manager.Deployment.Delete(ctx, dbName)
	if err != nil {
		return nil, err
	}

	return &apipb.DropSchemaResponse{
		Txid: tx.Id,
		Msg:  "success",
	}, nil
}

func (s *Service) GetMetadata(ctx context.Context, req *apipb.GetMetadataRequest) (*apipb.GetMetadataResponse, error) {

	nm := strings.ToLower(req.Database + "_" + req.Owner)
	db, err := s.manager.GetDatabase(ctx, nm)
	if err != nil {
		return nil, err
	}

	return &apipb.GetMetadataResponse{
		Database: convertDb(db),
	}, nil
}

func convertDb(db *cache.Database) *apipb.Database {
	var tables []*apipb.Table
	for _, t := range db.Tables {
		tables = append(tables, convertTable(t))
	}

	var indexes []*apipb.Index
	for _, i := range db.Indexes {
		indexes = append(indexes, convertIndex(i))
	}

	var queries []*apipb.Query
	for _, q := range db.Queries {
		queries = append(queries, convertQuery(q))
	}

	var roles []*apipb.Role
	for _, r := range db.Roles {
		roles = append(roles, convertRole(r))
	}

	return &apipb.Database{
		Name:        db.Name,
		Owner:       db.Owner,
		DefaultRole: db.DefaultRole,
		Tables:      tables,
		Indexes:     indexes,
		Queries:     queries,
		Roles:       roles,
	}
}

func convertTable(t *cache.Table) *apipb.Table {
	cols := make([]*apipb.Column, len(t.Columns))
	for i, c := range t.Columns {
		cols[i] = convertColumn(c)
	}

	return &apipb.Table{
		Name:    t.Name,
		Columns: cols,
	}
}

func convertColumn(c *cache.Column) *apipb.Column {
	attrs := make([]*apipb.Attribute, len(c.Attributes))
	for i, a := range c.Attributes {
		attrs[i] = convertAttribute(a)
	}

	return &apipb.Column{
		Name:       c.Name,
		Type:       apipb.DataType(c.Type.Int()),
		Attributes: attrs,
	}
}

func convertAttribute(a *cache.Attribute) *apipb.Attribute {
	str, err := conv.String(a.Value)
	if err != nil {
		fmt.Println("WARNING failed to convert attribute") // delete this once im sure this works fine
		return nil
	}

	return &apipb.Attribute{
		Type:  apipb.AttributeType(a.Type.Int()),
		Value: str,
	}
}

func convertIndex(i *cache.Index) *apipb.Index {
	return &apipb.Index{
		Name:    i.Name,
		Table:   i.Table,
		Columns: i.Columns,
		Using:   apipb.IndexType(i.Using.Int()),
	}
}

func convertQuery(q *cache.Executable) *apipb.Query {
	var args []*apipb.Arg
	for _, a := range q.Args {
		args = append(args, convertArg(a))
	}

	return &apipb.Query{
		Name:      q.Name,
		Statement: q.Statement,
		Table:     q.Table,
		Type:      apipb.QueryType(q.Type.Int()),
		Args:      args,
	}
}

func convertArg(a *cache.Arg) *apipb.Arg {
	str, err := conv.String(a.Value)
	if err != nil {
		fmt.Println("WARNING failed to convert arg") // delete this once im sure this works fine
		return nil
	}

	return &apipb.Arg{
		Position:      int32(a.Position),
		Static:        a.Static,
		InputPosition: int32(a.InputPosition),
		DefaultValue:  str,
		Type:          apipb.DataType(a.Type.Int()),
		Modifier:      apipb.ModifierType(a.Modifier.Int()),
	}
}

func convertRole(r *cache.Role) *apipb.Role {
	var perms []string
	for p := range r.Permissions {
		perms = append(perms, p)
	}

	return &apipb.Role{
		Name:        r.Name,
		Permissions: perms,
	}
}

func (s *Service) ListDatabases(ctx context.Context, req *apipb.ListDatabasesRequest) (*apipb.ListDatabasesResponse, error) {
	dbs, err := s.manager.Metadata.ListDatabases(ctx, req.Owner)
	if err != nil {
		return nil, err
	}

	return &apipb.ListDatabasesResponse{
		Databases: dbs,
	}, nil
}
