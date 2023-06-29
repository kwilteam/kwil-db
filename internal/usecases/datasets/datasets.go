package datasets

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type DatasetUseCase struct {
	engine           *engine.Engine
	accountStore     AccountStore
	log              log.Logger
	extensionConfigs []*extensions.ExtensionConfig

	sqliteFilePath string
}

func New(ctx context.Context, opts ...DatasetUseCaseOpt) (DatasetUseCaseInterface, error) {
	u := &DatasetUseCase{
		log:            log.NewNoOp(),
		sqliteFilePath: "",
	}

	for _, opt := range opts {
		opt(u)
	}

	var err error
	if u.engine == nil {
		u.engine, err = engine.Open(ctx,
			u.engineOpts()...,
		)
		if err != nil {
			return nil, err
		}
	}

	if u.accountStore == nil {
		u.accountStore, err = balances.NewAccountStore(
			balances.WithLogger(u.log),
		)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (u *DatasetUseCase) engineOpts() []engine.EngineOpt {
	opts := []engine.EngineOpt{
		engine.WithLogger(u.log),
	}
	if u.sqliteFilePath != "" {
		opts = append(opts, engine.WithPath(u.sqliteFilePath))
	}

	if len(u.extensionConfigs) > 0 {
		exts, err := connectExtensions(context.Background(), u.extensionConfigs)
		if err != nil {
			panic(err)
		}

		var initializers = make(map[string]engine.ExtensionInitializer)
		for name, ext := range exts {
			initializers[name] = extensionInitializeFunc(ext.CreateInstance)
		}

		opts = append(opts, engine.WithExtensions(initializers))
	}

	return opts
}

func (u *DatasetUseCase) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	return u.engine.ListDatasets(ctx, owner)
}

func (u *DatasetUseCase) GetSchema(ctx context.Context, dbid string) (*entity.Schema, error) {
	db, err := u.engine.GetDataset(ctx, dbid)
	if err != nil {
		return nil, err
	}

	actions := db.ListProcedures()
	tables, err := db.ListTables(ctx)
	if err != nil {
		return nil, err
	}

	dbName, dbOwner := db.Metadata()

	return &entity.Schema{
		Owner:   dbOwner,
		Name:    dbName,
		Actions: convertActions(actions),
		Tables:  convertTables(tables),
	}, nil
}

func (u *DatasetUseCase) Close() error {
	var errs []error

	err := u.accountStore.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err2 := u.engine.Close()
	if err2 != nil {
		errs = append(errs, err2)
	}

	return errors.Join(errs...)
}

func connectExtensions(ctx context.Context, confs []*extensions.ExtensionConfig) (map[string]*extensions.Extension, error) {
	exts := make(map[string]*extensions.Extension, len(confs))

	for _, conf := range confs {
		ext := extensions.New(conf)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", conf.Name, err)
		}

		exts[conf.Name] = ext
	}

	return exts, nil
}

type extensionInitializeFunc func(ctx context.Context, metadata map[string]string) (*extensions.Instance, error)

func (e extensionInitializeFunc) CreateInstance(ctx context.Context, metadata map[string]string) (engine.ExtensionInstance, error) {
	return e(ctx, metadata)
}
