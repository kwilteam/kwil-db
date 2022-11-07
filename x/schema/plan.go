package schema

import (
	"github.com/google/uuid"
)

type PlanRepository interface {
	SavePlan(wallet, db string, data []byte) (uuid.UUID, error)
	GetPlanInfo(id uuid.UUID) (PlanInfo, error)
}

type PlanRequest struct {
	Wallet     string
	Database   string
	SchemaData []byte
}

type PlanInfo struct {
	Wallet   string
	Database string
	Data     []byte
}

type Plan struct {
	ID            uuid.UUID
	Version       string
	Name          string
	Reversible    bool
	Transactional bool
	Changes       []Change
}

type Change struct {
	Cmd     string
	Args    []any
	Comment string
	Reverse string
}

type inmemoryPlanRepo struct {
	plans map[uuid.UUID]PlanInfo
}

func NewInMemoryPlanRepository() PlanRepository {
	return &inmemoryPlanRepo{plans: make(map[uuid.UUID]PlanInfo)}
}

func (r *inmemoryPlanRepo) SavePlan(wallet, db string, data []byte) (uuid.UUID, error) {
	id := uuid.New()
	r.plans[id] = PlanInfo{Wallet: wallet, Database: db, Data: data}
	return id, nil
}

func (r *inmemoryPlanRepo) GetPlanInfo(id uuid.UUID) (PlanInfo, error) {
	if data, ok := r.plans[id]; ok {
		return data, nil
	}
	return PlanInfo{}, ErrPlanNotFound
}
