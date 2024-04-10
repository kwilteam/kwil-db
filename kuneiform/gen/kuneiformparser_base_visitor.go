// Code generated from KuneiformParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // KuneiformParser
import "github.com/antlr4-go/antlr/v4"

type BaseKuneiformParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseKuneiformParserVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_mode(ctx *Stmt_modeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDatabase_declaration(ctx *Database_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUse_declaration(ctx *Use_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitTable_declaration(ctx *Table_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitColumn_def(ctx *Column_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIndex_def(ctx *Index_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_key_def(ctx *Foreign_key_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_key_action(ctx *Foreign_key_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIdentifier_list(ctx *Identifier_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitType_selector(ctx *Type_selectorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMIN(ctx *MINContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMAX(ctx *MAXContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMIN_LEN(ctx *MIN_LENContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMAX_LEN(ctx *MAX_LENContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNOT_NULL(ctx *NOT_NULLContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitPRIMARY_KEY(ctx *PRIMARY_KEYContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDEFAULT(ctx *DEFAULTContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUNIQUE(ctx *UNIQUEContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitAction_declaration(ctx *Action_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_declaration(ctx *Procedure_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_return(ctx *Stmt_returnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_typed_param_list(ctx *Stmt_typed_param_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_type_selector(ctx *Stmt_type_selectorContext) interface{} {
	return v.VisitChildren(ctx)
}
