package apisvc

//const DATABASE_EMITTER_ALIAS = "service.database.emitter"
//
//// DBRequest DB Request represents the plan to be applied to the DB
//// (this is not in stone and needs to be further updated as more is known)
//type DBRequest struct {
//	/*
//		Discussed 10/17 during code walk-through
//		1) This needs to be the changes/plan itself
//		2) This needs to be a proto
//		3) This needs to a single plan, not a list of plans
//		4) The plan will contain a list of changes
//		5) The plan changes will have previously been sent to "pre-apply"
//	*/
//
//	IdempotentKey string // key to use for de-duplicating commands in the target Db and for retrieving the request info
//	RoutingKey    string // represents a command path to ensure messages are processed in order
//	DBCommand     string // actual DDL or INSERT/UPDATE/DELETE SQL
//}

// Change to plan-changes --> DBRequest
/*
func getCreateDbRequest(req *apipb.CreateDatabaseRequest) *composer.Message {
	panic("not implemented")
}

// Will remove and just use above once changed
func getUpdateDbRequest(req *apipb.UpdateDatabaseRequest) *composer.Message {
	panic("not implemented")
}

func doSubmitRequest[T any](ctx context.Context, req T, fn func(T) *composer.Message) (string, error) {
	emitter := x.Resolve[composer.EmitTrack](ctx, composer.SERVICE_ALIAS)
	if emitter == nil {
		return "", fmt.Errorf("failed to resolve request manager %s", composer.SERVICE_ALIAS)
	}

	emitter.Submit(ctx, fn(req))

	db_req := fn(req)
	a := emitter.Submit(ctx, db_req)
	<-a.DoneCh() // blocking call

	return string(a.Get()), nil
}
*/
