// Code generated from KuneiformParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // KuneiformParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by KuneiformParser.
type KuneiformParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by KuneiformParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_mode.
	VisitStmt_mode(ctx *Stmt_modeContext) interface{}

	// Visit a parse tree produced by KuneiformParser#database_declaration.
	VisitDatabase_declaration(ctx *Database_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#use_declaration.
	VisitUse_declaration(ctx *Use_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#table_declaration.
	VisitTable_declaration(ctx *Table_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#column_def.
	VisitColumn_def(ctx *Column_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#index_def.
	VisitIndex_def(ctx *Index_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#foreign_key_def.
	VisitForeign_key_def(ctx *Foreign_key_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#foreign_key_action.
	VisitForeign_key_action(ctx *Foreign_key_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#identifier_list.
	VisitIdentifier_list(ctx *Identifier_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by KuneiformParser#type_selector.
	VisitType_selector(ctx *Type_selectorContext) interface{}

	// Visit a parse tree produced by KuneiformParser#MIN.
	VisitMIN(ctx *MINContext) interface{}

	// Visit a parse tree produced by KuneiformParser#MAX.
	VisitMAX(ctx *MAXContext) interface{}

	// Visit a parse tree produced by KuneiformParser#MIN_LEN.
	VisitMIN_LEN(ctx *MIN_LENContext) interface{}

	// Visit a parse tree produced by KuneiformParser#MAX_LEN.
	VisitMAX_LEN(ctx *MAX_LENContext) interface{}

	// Visit a parse tree produced by KuneiformParser#NOT_NULL.
	VisitNOT_NULL(ctx *NOT_NULLContext) interface{}

	// Visit a parse tree produced by KuneiformParser#PRIMARY_KEY.
	VisitPRIMARY_KEY(ctx *PRIMARY_KEYContext) interface{}

	// Visit a parse tree produced by KuneiformParser#DEFAULT.
	VisitDEFAULT(ctx *DEFAULTContext) interface{}

	// Visit a parse tree produced by KuneiformParser#UNIQUE.
	VisitUNIQUE(ctx *UNIQUEContext) interface{}

	// Visit a parse tree produced by KuneiformParser#action_declaration.
	VisitAction_declaration(ctx *Action_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_declaration.
	VisitProcedure_declaration(ctx *Procedure_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_return.
	VisitStmt_return(ctx *Stmt_returnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_typed_param_list.
	VisitStmt_typed_param_list(ctx *Stmt_typed_param_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_type_selector.
	VisitStmt_type_selector(ctx *Stmt_type_selectorContext) interface{}
}
