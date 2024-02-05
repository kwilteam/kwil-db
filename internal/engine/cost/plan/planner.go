package plan

//
//type Planner interface {
//	Plan(node tree.AstNode, ctx PlannerContext) (*Plan, error)
//}
//
//type StmtPlanner struct {
//}
//
//func NewStmtPlanner() *StmtPlanner {
//	return &StmtPlanner{}
//}
//
//type PlannerContext struct {
//	currentSchema *types.Schema
//	dataset       SchemaGetter
//}
//
//func NewPlannerContext(schema *types.Schema, dataset SchemaGetter) *PlannerContext {
//	return &PlannerContext{
//		currentSchema: schema,
//		dataset:       dataset,
//	}
//}
//
//func (p *StmtPlanner) Plan(node tree.AstNode, ctx *PlannerContext) (Plan, error) {
//	switch node := node.(type) {
//	case *tree.Select:
//		return p.planSelect(node, ctx)
//	case *tree.Insert:
//		return p.planInsert(node, ctx)
//	case *tree.Delete:
//		return p.planDelete(node, ctx)
//	case *tree.Update:
//		return p.planUpdate(node, ctx)
//	}
//	return nil, nil
//}
//
//func (p *StmtPlanner) planSelect(node *tree.Select, ctx *PlannerContext) (Plan, error) {
//	// 1. build logical plan, i.e. transform AST to logical plan
//
//	//ds := map[string]*types.Schema
//	//ts := newTableCollector().collect(node)
//	//for _, t := range ts {
//	//
//	//}
//	var ts map[string]*types.Table
//	for _, t := range ctx.currentSchema.Tables {
//		ts[t.name] = t
//	}
//
//	//transformer := NewRelationTransformer(ts)
//	//logicalPlan, err := transformer.Transform(node)
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	pb := &Builder{
//		tables: ts,
//		info:   ctx.dataset,
//		ctx:    nil,
//	}
//
//	logicalPlan := pb.build(node)
//	return logicalPlan, nil
//
//	//// 2. navie optimize logical plan
//	//return nil, nil
//}
//
//// planInsert builds a plan for insert statement.
//// Relation inside the insert statement will be transformed to logical plan
//// first. The relation is created when analyzing the insert statement. We don't
//// have analysis phase now.
//func (p *StmtPlanner) planInsert(node *tree.Insert, ctx *PlannerContext) (Plan, error) {
//	return nil, nil
//}
//
//func (p *StmtPlanner) planDelete(node *tree.Delete, ctx *PlannerContext) (Plan, error) {
//	return nil, nil
//}
//
//func (p *StmtPlanner) planUpdate(node *tree.Update, ctx *PlannerContext) (Plan, error) {
//	return nil, nil
//}
