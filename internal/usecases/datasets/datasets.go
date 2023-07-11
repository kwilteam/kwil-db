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
	engine        *engine.Engine
	accountStore  AccountStore
	log           log.Logger
	extensionUrls []string
	gas_enabled   bool

	sqliteFilePath string
}

func New(ctx context.Context, opts ...DatasetUseCaseOpt) (DatasetUseCaseInterface, error) {
	u := &DatasetUseCase{
		log:            log.NewNoOp(),
		sqliteFilePath: "",
		extensionUrls:  []string{},
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
	u.gas_enabled = true
	return u, nil
}

func (u *DatasetUseCase) engineOpts() []engine.EngineOpt {
	opts := []engine.EngineOpt{
		engine.WithLogger(u.log),
	}
	if u.sqliteFilePath != "" {
		opts = append(opts, engine.WithPath(u.sqliteFilePath))
	}

	if len(u.extensionUrls) > 0 {
		exts, err := connectExtensions(context.Background(), u.extensionUrls)
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

func connectExtensions(ctx context.Context, urls []string) (map[string]*extensions.Extension, error) {
	exts := make(map[string]*extensions.Extension, len(urls))

	for _, url := range urls {
		ext := extensions.New(url)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", ext.Name(), err)
		}

		_, ok := exts[ext.Name()]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", ext.Name())
		}

		exts[ext.Name()] = ext
	}

	return exts, nil
}

type extensionInitializeFunc func(ctx context.Context, metadata map[string]string) (*extensions.Instance, error)

func (e extensionInitializeFunc) CreateInstance(ctx context.Context, metadata map[string]string) (engine.ExtensionInstance, error) {
	return e(ctx, metadata)
}

func (u *DatasetUseCase) UpdateGasCosts(gas_enabled bool) {
	u.gas_enabled = gas_enabled
}

func (u *DatasetUseCase) GasEnabled() bool {
	return u.gas_enabled
}
