package apisvc

import (
	"context"
	"fmt"
	"kwil/x/crypto"
	"kwil/x/proto/apipb"
	"kwil/x/sqlx/schema"
	"strings"
)

func (s *Service) DeploySchema(ctx context.Context, req *apipb.DeploySchemaRequest) (*apipb.DeploySchemaResponse, error) {

	// verify the tx
	tx := crypto.Tx{}
	tx.Convert(req.Tx)
	err := tx.Verify()
	if err != nil {
		return nil, err
	}

	p, err := s.p.GetPriceForDDL(ctx)
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
	// TODO: uncomment this

	err = s.manager.Deposits.Spend(ctx, tx.Sender, tx.Fee)
	if err != nil {
		return nil, err
	}

	err = s.manager.Deployment.Deploy(ctx, tx.Sender, tx.Data)
	if err != nil {
		return nil, err
	}

	return &apipb.DeploySchemaResponse{
		Txid: tx.Id,
		Msg:  "success",
	}, nil
}

/*
func (s *Service) DropDatabase(ctx context.Context, req *apipb.DropDatabaseRequest) (*apipb.DropDatabaseResponse, error) {

		// verify the tx
		tx := crypto.Tx{}
		tx.Convert(req.Tx)
		err := tx.Verify()
		if err != nil {
			return nil, err
		}

		p, err := s.p.GetPriceForDeleteSchema(ctx)
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


		body, err := Unmarshal[DropDatabaseBody](req.Tx.Data)
		if err != nil {
			return nil, err
		}

		err = s.manager.Deployment.Delete(ctx, body.Database)
		if err != nil {
			return nil, err
		}

		return &apipb.DropDatabaseResponse{
			Txid: tx.Id,
			Msg:  "success",
		}, nil
	}
*/
func (s *Service) GetMetadata(ctx context.Context, req *apipb.GetMetadataRequest) (*apipb.GetMetadataResponse, error) {

	nm := schema.FormatOwner(strings.ToLower(req.Owner + "_" + req.Database))
	db, err := s.manager.Export(nm)
	if err != nil {
		return nil, err
	}

	return &apipb.GetMetadataResponse{
		Metadata: db.AsProtobuf(),
	}, nil
}
