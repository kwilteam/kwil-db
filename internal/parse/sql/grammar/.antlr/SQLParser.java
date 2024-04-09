// Generated from /Users/brennanlamey/kwil-db/internal/parse/sql/grammar/SQLParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class SQLParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		SCOL=1, DOT=2, OPEN_PAR=3, CLOSE_PAR=4, L_BRACKET=5, R_BRACKET=6, COMMA=7, 
		ASSIGN=8, STAR=9, PLUS=10, MINUS=11, DIV=12, MOD=13, LT=14, LT_EQ=15, 
		GT=16, GT_EQ=17, NOT_EQ1=18, NOT_EQ2=19, TYPE_CAST=20, ADD_=21, ALL_=22, 
		AND_=23, ASC_=24, AS_=25, BETWEEN_=26, BY_=27, CASE_=28, COLLATE_=29, 
		COMMIT_=30, CONFLICT_=31, CREATE_=32, CROSS_=33, DEFAULT_=34, DELETE_=35, 
		DESC_=36, DISTINCT_=37, DO_=38, ELSE_=39, END_=40, ESCAPE_=41, EXCEPT_=42, 
		EXISTS_=43, FILTER_=44, FIRST_=45, FROM_=46, FULL_=47, GROUPS_=48, GROUP_=49, 
		HAVING_=50, INNER_=51, INSERT_=52, INTERSECT_=53, INTO_=54, IN_=55, ISNULL_=56, 
		IS_=57, JOIN_=58, LAST_=59, LEFT_=60, LIKE_=61, LIMIT_=62, NOTHING_=63, 
		NOTNULL_=64, NOT_=65, NULLS_=66, OFFSET_=67, OF_=68, ON_=69, ORDER_=70, 
		OR_=71, OUTER_=72, RAISE_=73, REPLACE_=74, RETURNING_=75, RIGHT_=76, SELECT_=77, 
		SET_=78, THEN_=79, UNION_=80, UPDATE_=81, USING_=82, VALUES_=83, WHEN_=84, 
		WHERE_=85, WITH_=86, BOOLEAN_LITERAL=87, NUMERIC_LITERAL=88, BLOB_LITERAL=89, 
		TEXT_LITERAL=90, NULL_LITERAL=91, IDENTIFIER=92, BIND_PARAMETER=93, SINGLE_LINE_COMMENT=94, 
		MULTILINE_COMMENT=95, SPACES=96, UNEXPECTED_CHAR=97;
	public static final int
		RULE_statements = 0, RULE_sql_stmt_list = 1, RULE_sql_stmt = 2, RULE_indexed_column = 3, 
		RULE_cte_table_name = 4, RULE_common_table_expression = 5, RULE_common_table_stmt = 6, 
		RULE_delete_core = 7, RULE_delete_stmt = 8, RULE_variable = 9, RULE_function_call = 10, 
		RULE_column_ref = 11, RULE_when_clause = 12, RULE_expr = 13, RULE_subquery = 14, 
		RULE_expr_list = 15, RULE_comparisonOperator = 16, RULE_cast_type = 17, 
		RULE_type_cast = 18, RULE_value_row = 19, RULE_values_clause = 20, RULE_insert_core = 21, 
		RULE_insert_stmt = 22, RULE_returning_clause = 23, RULE_upsert_update = 24, 
		RULE_upsert_clause = 25, RULE_select_core = 26, RULE_select_stmt = 27, 
		RULE_join_relation = 28, RULE_relation = 29, RULE_simple_select = 30, 
		RULE_table_or_subquery = 31, RULE_result_column = 32, RULE_returning_clause_result_column = 33, 
		RULE_join_operator = 34, RULE_join_constraint = 35, RULE_compound_operator = 36, 
		RULE_update_set_subclause = 37, RULE_update_core = 38, RULE_update_stmt = 39, 
		RULE_column_name_list = 40, RULE_qualified_table_name = 41, RULE_order_by_stmt = 42, 
		RULE_limit_stmt = 43, RULE_ordering_term = 44, RULE_asc_desc = 45, RULE_function_keyword = 46, 
		RULE_function_name = 47, RULE_table_name = 48, RULE_table_alias = 49, 
		RULE_column_name = 50, RULE_column_alias = 51, RULE_collation_name = 52, 
		RULE_index_name = 53;
	private static String[] makeRuleNames() {
		return new String[] {
			"statements", "sql_stmt_list", "sql_stmt", "indexed_column", "cte_table_name", 
			"common_table_expression", "common_table_stmt", "delete_core", "delete_stmt", 
			"variable", "function_call", "column_ref", "when_clause", "expr", "subquery", 
			"expr_list", "comparisonOperator", "cast_type", "type_cast", "value_row", 
			"values_clause", "insert_core", "insert_stmt", "returning_clause", "upsert_update", 
			"upsert_clause", "select_core", "select_stmt", "join_relation", "relation", 
			"simple_select", "table_or_subquery", "result_column", "returning_clause_result_column", 
			"join_operator", "join_constraint", "compound_operator", "update_set_subclause", 
			"update_core", "update_stmt", "column_name_list", "qualified_table_name", 
			"order_by_stmt", "limit_stmt", "ordering_term", "asc_desc", "function_keyword", 
			"function_name", "table_name", "table_alias", "column_name", "column_alias", 
			"collation_name", "index_name"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "';'", "'.'", "'('", "')'", "'['", "']'", "','", "'='", "'*'", 
			"'+'", "'-'", "'/'", "'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'<>'", 
			"'::'", "'ADD'", "'ALL'", "'AND'", "'ASC'", "'AS'", "'BETWEEN'", "'BY'", 
			"'CASE'", "'COLLATE'", "'COMMIT'", "'CONFLICT'", "'CREATE'", "'CROSS'", 
			"'DEFAULT'", "'DELETE'", "'DESC'", "'DISTINCT'", "'DO'", "'ELSE'", "'END'", 
			"'ESCAPE'", "'EXCEPT'", "'EXISTS'", "'FILTER'", "'FIRST'", "'FROM'", 
			"'FULL'", "'GROUPS'", "'GROUP'", "'HAVING'", "'INNER'", "'INSERT'", "'INTERSECT'", 
			"'INTO'", "'IN'", "'ISNULL'", "'IS'", "'JOIN'", "'LAST'", "'LEFT'", "'LIKE'", 
			"'LIMIT'", "'NOTHING'", "'NOTNULL'", "'NOT'", "'NULLS'", "'OFFSET'", 
			"'OF'", "'ON'", "'ORDER'", "'OR'", "'OUTER'", "'RAISE'", "'REPLACE'", 
			"'RETURNING'", "'RIGHT'", "'SELECT'", "'SET'", "'THEN'", "'UNION'", "'UPDATE'", 
			"'USING'", "'VALUES'", "'WHEN'", "'WHERE'", "'WITH'", null, null, null, 
			null, "'null'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "SCOL", "DOT", "OPEN_PAR", "CLOSE_PAR", "L_BRACKET", "R_BRACKET", 
			"COMMA", "ASSIGN", "STAR", "PLUS", "MINUS", "DIV", "MOD", "LT", "LT_EQ", 
			"GT", "GT_EQ", "NOT_EQ1", "NOT_EQ2", "TYPE_CAST", "ADD_", "ALL_", "AND_", 
			"ASC_", "AS_", "BETWEEN_", "BY_", "CASE_", "COLLATE_", "COMMIT_", "CONFLICT_", 
			"CREATE_", "CROSS_", "DEFAULT_", "DELETE_", "DESC_", "DISTINCT_", "DO_", 
			"ELSE_", "END_", "ESCAPE_", "EXCEPT_", "EXISTS_", "FILTER_", "FIRST_", 
			"FROM_", "FULL_", "GROUPS_", "GROUP_", "HAVING_", "INNER_", "INSERT_", 
			"INTERSECT_", "INTO_", "IN_", "ISNULL_", "IS_", "JOIN_", "LAST_", "LEFT_", 
			"LIKE_", "LIMIT_", "NOTHING_", "NOTNULL_", "NOT_", "NULLS_", "OFFSET_", 
			"OF_", "ON_", "ORDER_", "OR_", "OUTER_", "RAISE_", "REPLACE_", "RETURNING_", 
			"RIGHT_", "SELECT_", "SET_", "THEN_", "UNION_", "UPDATE_", "USING_", 
			"VALUES_", "WHEN_", "WHERE_", "WITH_", "BOOLEAN_LITERAL", "NUMERIC_LITERAL", 
			"BLOB_LITERAL", "TEXT_LITERAL", "NULL_LITERAL", "IDENTIFIER", "BIND_PARAMETER", 
			"SINGLE_LINE_COMMENT", "MULTILINE_COMMENT", "SPACES", "UNEXPECTED_CHAR"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "SQLParser.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public SQLParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StatementsContext extends ParserRuleContext {
		public TerminalNode EOF() { return getToken(SQLParser.EOF, 0); }
		public List<Sql_stmt_listContext> sql_stmt_list() {
			return getRuleContexts(Sql_stmt_listContext.class);
		}
		public Sql_stmt_listContext sql_stmt_list(int i) {
			return getRuleContext(Sql_stmt_listContext.class,i);
		}
		public StatementsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statements; }
	}

	public final StatementsContext statements() throws RecognitionException {
		StatementsContext _localctx = new StatementsContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_statements);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(111);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4503633987108866L) != 0) || ((((_la - 77)) & ~0x3f) == 0 && ((1L << (_la - 77)) & 529L) != 0)) {
				{
				{
				setState(108);
				sql_stmt_list();
				}
				}
				setState(113);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(114);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Sql_stmt_listContext extends ParserRuleContext {
		public List<Sql_stmtContext> sql_stmt() {
			return getRuleContexts(Sql_stmtContext.class);
		}
		public Sql_stmtContext sql_stmt(int i) {
			return getRuleContext(Sql_stmtContext.class,i);
		}
		public List<TerminalNode> SCOL() { return getTokens(SQLParser.SCOL); }
		public TerminalNode SCOL(int i) {
			return getToken(SQLParser.SCOL, i);
		}
		public Sql_stmt_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sql_stmt_list; }
	}

	public final Sql_stmt_listContext sql_stmt_list() throws RecognitionException {
		Sql_stmt_listContext _localctx = new Sql_stmt_listContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_sql_stmt_list);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(119);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==SCOL) {
				{
				{
				setState(116);
				match(SCOL);
				}
				}
				setState(121);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(122);
			sql_stmt();
			setState(131);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					{
					{
					setState(124); 
					_errHandler.sync(this);
					_la = _input.LA(1);
					do {
						{
						{
						setState(123);
						match(SCOL);
						}
						}
						setState(126); 
						_errHandler.sync(this);
						_la = _input.LA(1);
					} while ( _la==SCOL );
					setState(128);
					sql_stmt();
					}
					} 
				}
				setState(133);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
			}
			setState(137);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					{
					{
					setState(134);
					match(SCOL);
					}
					} 
				}
				setState(139);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Sql_stmtContext extends ParserRuleContext {
		public Delete_stmtContext delete_stmt() {
			return getRuleContext(Delete_stmtContext.class,0);
		}
		public Insert_stmtContext insert_stmt() {
			return getRuleContext(Insert_stmtContext.class,0);
		}
		public Select_stmtContext select_stmt() {
			return getRuleContext(Select_stmtContext.class,0);
		}
		public Update_stmtContext update_stmt() {
			return getRuleContext(Update_stmtContext.class,0);
		}
		public Sql_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sql_stmt; }
	}

	public final Sql_stmtContext sql_stmt() throws RecognitionException {
		Sql_stmtContext _localctx = new Sql_stmtContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_sql_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(144);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,5,_ctx) ) {
			case 1:
				{
				setState(140);
				delete_stmt();
				}
				break;
			case 2:
				{
				setState(141);
				insert_stmt();
				}
				break;
			case 3:
				{
				setState(142);
				select_stmt();
				}
				break;
			case 4:
				{
				setState(143);
				update_stmt();
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Indexed_columnContext extends ParserRuleContext {
		public Column_nameContext column_name() {
			return getRuleContext(Column_nameContext.class,0);
		}
		public Indexed_columnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexed_column; }
	}

	public final Indexed_columnContext indexed_column() throws RecognitionException {
		Indexed_columnContext _localctx = new Indexed_columnContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_indexed_column);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(146);
			column_name();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Cte_table_nameContext extends ParserRuleContext {
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public List<Column_nameContext> column_name() {
			return getRuleContexts(Column_nameContext.class);
		}
		public Column_nameContext column_name(int i) {
			return getRuleContext(Column_nameContext.class,i);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Cte_table_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_cte_table_name; }
	}

	public final Cte_table_nameContext cte_table_name() throws RecognitionException {
		Cte_table_nameContext _localctx = new Cte_table_nameContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_cte_table_name);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(148);
			table_name();
			setState(160);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OPEN_PAR) {
				{
				setState(149);
				match(OPEN_PAR);
				setState(150);
				column_name();
				setState(155);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(151);
					match(COMMA);
					setState(152);
					column_name();
					}
					}
					setState(157);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(158);
				match(CLOSE_PAR);
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Common_table_expressionContext extends ParserRuleContext {
		public Cte_table_nameContext cte_table_name() {
			return getRuleContext(Cte_table_nameContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public Select_coreContext select_core() {
			return getRuleContext(Select_coreContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public Common_table_expressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_common_table_expression; }
	}

	public final Common_table_expressionContext common_table_expression() throws RecognitionException {
		Common_table_expressionContext _localctx = new Common_table_expressionContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_common_table_expression);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(162);
			cte_table_name();
			setState(163);
			match(AS_);
			setState(164);
			match(OPEN_PAR);
			setState(165);
			select_core();
			setState(166);
			match(CLOSE_PAR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Common_table_stmtContext extends ParserRuleContext {
		public TerminalNode WITH_() { return getToken(SQLParser.WITH_, 0); }
		public List<Common_table_expressionContext> common_table_expression() {
			return getRuleContexts(Common_table_expressionContext.class);
		}
		public Common_table_expressionContext common_table_expression(int i) {
			return getRuleContext(Common_table_expressionContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Common_table_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_common_table_stmt; }
	}

	public final Common_table_stmtContext common_table_stmt() throws RecognitionException {
		Common_table_stmtContext _localctx = new Common_table_stmtContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_common_table_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(168);
			match(WITH_);
			setState(169);
			common_table_expression();
			setState(174);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(170);
				match(COMMA);
				setState(171);
				common_table_expression();
				}
				}
				setState(176);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Delete_coreContext extends ParserRuleContext {
		public TerminalNode DELETE_() { return getToken(SQLParser.DELETE_, 0); }
		public TerminalNode FROM_() { return getToken(SQLParser.FROM_, 0); }
		public Qualified_table_nameContext qualified_table_name() {
			return getRuleContext(Qualified_table_nameContext.class,0);
		}
		public TerminalNode WHERE_() { return getToken(SQLParser.WHERE_, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Returning_clauseContext returning_clause() {
			return getRuleContext(Returning_clauseContext.class,0);
		}
		public Delete_coreContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_delete_core; }
	}

	public final Delete_coreContext delete_core() throws RecognitionException {
		Delete_coreContext _localctx = new Delete_coreContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_delete_core);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(177);
			match(DELETE_);
			setState(178);
			match(FROM_);
			setState(179);
			qualified_table_name();
			setState(182);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WHERE_) {
				{
				setState(180);
				match(WHERE_);
				setState(181);
				expr(0);
				}
			}

			setState(185);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==RETURNING_) {
				{
				setState(184);
				returning_clause();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Delete_stmtContext extends ParserRuleContext {
		public Delete_coreContext delete_core() {
			return getRuleContext(Delete_coreContext.class,0);
		}
		public Common_table_stmtContext common_table_stmt() {
			return getRuleContext(Common_table_stmtContext.class,0);
		}
		public Delete_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_delete_stmt; }
	}

	public final Delete_stmtContext delete_stmt() throws RecognitionException {
		Delete_stmtContext _localctx = new Delete_stmtContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_delete_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(188);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WITH_) {
				{
				setState(187);
				common_table_stmt();
				}
			}

			setState(190);
			delete_core();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class VariableContext extends ParserRuleContext {
		public TerminalNode BIND_PARAMETER() { return getToken(SQLParser.BIND_PARAMETER, 0); }
		public VariableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_variable; }
	}

	public final VariableContext variable() throws RecognitionException {
		VariableContext _localctx = new VariableContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_variable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(192);
			match(BIND_PARAMETER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Function_callContext extends ParserRuleContext {
		public Function_nameContext function_name() {
			return getRuleContext(Function_nameContext.class,0);
		}
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public TerminalNode STAR() { return getToken(SQLParser.STAR, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode DISTINCT_() { return getToken(SQLParser.DISTINCT_, 0); }
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Function_callContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_function_call; }
	}

	public final Function_callContext function_call() throws RecognitionException {
		Function_callContext _localctx = new Function_callContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_function_call);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(194);
			function_name();
			setState(195);
			match(OPEN_PAR);
			setState(208);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OPEN_PAR:
			case PLUS:
			case MINUS:
			case CASE_:
			case DISTINCT_:
			case EXISTS_:
			case LIKE_:
			case NOT_:
			case REPLACE_:
			case BOOLEAN_LITERAL:
			case NUMERIC_LITERAL:
			case BLOB_LITERAL:
			case TEXT_LITERAL:
			case NULL_LITERAL:
			case IDENTIFIER:
			case BIND_PARAMETER:
				{
				{
				setState(197);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==DISTINCT_) {
					{
					setState(196);
					match(DISTINCT_);
					}
				}

				setState(199);
				expr(0);
				setState(204);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(200);
					match(COMMA);
					setState(201);
					expr(0);
					}
					}
					setState(206);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
				}
				break;
			case STAR:
				{
				setState(207);
				match(STAR);
				}
				break;
			case CLOSE_PAR:
				break;
			default:
				break;
			}
			setState(210);
			match(CLOSE_PAR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Column_refContext extends ParserRuleContext {
		public Column_nameContext column_name() {
			return getRuleContext(Column_nameContext.class,0);
		}
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public TerminalNode DOT() { return getToken(SQLParser.DOT, 0); }
		public Column_refContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column_ref; }
	}

	public final Column_refContext column_ref() throws RecognitionException {
		Column_refContext _localctx = new Column_refContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_column_ref);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(215);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,15,_ctx) ) {
			case 1:
				{
				setState(212);
				table_name();
				setState(213);
				match(DOT);
				}
				break;
			}
			setState(217);
			column_name();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class When_clauseContext extends ParserRuleContext {
		public ExprContext condition;
		public ExprContext result;
		public TerminalNode WHEN_() { return getToken(SQLParser.WHEN_, 0); }
		public TerminalNode THEN_() { return getToken(SQLParser.THEN_, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public When_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_when_clause; }
	}

	public final When_clauseContext when_clause() throws RecognitionException {
		When_clauseContext _localctx = new When_clauseContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_when_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(219);
			match(WHEN_);
			setState(220);
			((When_clauseContext)_localctx).condition = expr(0);
			setState(221);
			match(THEN_);
			setState(222);
			((When_clauseContext)_localctx).result = expr(0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ExprContext extends ParserRuleContext {
		public ExprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expr; }
	 
		public ExprContext() { }
		public void copyFrom(ExprContext ctx) {
			super.copyFrom(ctx);
		}
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Subquery_exprContext extends ExprContext {
		public SubqueryContext subquery() {
			return getRuleContext(SubqueryContext.class,0);
		}
		public TerminalNode EXISTS_() { return getToken(SQLParser.EXISTS_, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public Subquery_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Logical_not_exprContext extends ExprContext {
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Logical_not_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Boolean_literal_exprContext extends ExprContext {
		public TerminalNode BOOLEAN_LITERAL() { return getToken(SQLParser.BOOLEAN_LITERAL, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Boolean_literal_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Comparison_exprContext extends ExprContext {
		public ExprContext left;
		public ExprContext right;
		public ComparisonOperatorContext comparisonOperator() {
			return getRuleContext(ComparisonOperatorContext.class,0);
		}
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public Comparison_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Like_exprContext extends ExprContext {
		public ExprContext elem;
		public Token operator;
		public ExprContext pattern;
		public ExprContext escape;
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode LIKE_() { return getToken(SQLParser.LIKE_, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public TerminalNode ESCAPE_() { return getToken(SQLParser.ESCAPE_, 0); }
		public Like_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Null_exprContext extends ExprContext {
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode ISNULL_() { return getToken(SQLParser.ISNULL_, 0); }
		public TerminalNode NOTNULL_() { return getToken(SQLParser.NOTNULL_, 0); }
		public Null_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Column_exprContext extends ExprContext {
		public Column_refContext column_ref() {
			return getRuleContext(Column_refContext.class,0);
		}
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Column_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class In_subquery_exprContext extends ExprContext {
		public ExprContext elem;
		public Token operator;
		public SubqueryContext subquery() {
			return getRuleContext(SubqueryContext.class,0);
		}
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode IN_() { return getToken(SQLParser.IN_, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public In_subquery_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Arithmetic_exprContext extends ExprContext {
		public ExprContext left;
		public Token operator;
		public ExprContext right;
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode STAR() { return getToken(SQLParser.STAR, 0); }
		public TerminalNode DIV() { return getToken(SQLParser.DIV, 0); }
		public TerminalNode MOD() { return getToken(SQLParser.MOD, 0); }
		public TerminalNode PLUS() { return getToken(SQLParser.PLUS, 0); }
		public TerminalNode MINUS() { return getToken(SQLParser.MINUS, 0); }
		public Arithmetic_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Logical_binary_exprContext extends ExprContext {
		public ExprContext left;
		public Token operator;
		public ExprContext right;
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode AND_() { return getToken(SQLParser.AND_, 0); }
		public TerminalNode OR_() { return getToken(SQLParser.OR_, 0); }
		public Logical_binary_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Variable_exprContext extends ExprContext {
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Variable_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Text_literal_exprContext extends ExprContext {
		public TerminalNode TEXT_LITERAL() { return getToken(SQLParser.TEXT_LITERAL, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Text_literal_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Unary_exprContext extends ExprContext {
		public Token operator;
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode MINUS() { return getToken(SQLParser.MINUS, 0); }
		public TerminalNode PLUS() { return getToken(SQLParser.PLUS, 0); }
		public Unary_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Collate_exprContext extends ExprContext {
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode COLLATE_() { return getToken(SQLParser.COLLATE_, 0); }
		public Collation_nameContext collation_name() {
			return getRuleContext(Collation_nameContext.class,0);
		}
		public Collate_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Parenthesized_exprContext extends ExprContext {
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Parenthesized_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Between_exprContext extends ExprContext {
		public ExprContext elem;
		public Token operator;
		public ExprContext low;
		public ExprContext high;
		public TerminalNode AND_() { return getToken(SQLParser.AND_, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode BETWEEN_() { return getToken(SQLParser.BETWEEN_, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public Between_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_list_exprContext extends ExprContext {
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public Expr_listContext expr_list() {
			return getRuleContext(Expr_listContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public Expr_list_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Numeric_literal_exprContext extends ExprContext {
		public TerminalNode NUMERIC_LITERAL() { return getToken(SQLParser.NUMERIC_LITERAL, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Numeric_literal_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Null_literal_exprContext extends ExprContext {
		public TerminalNode NULL_LITERAL() { return getToken(SQLParser.NULL_LITERAL, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Null_literal_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class In_list_exprContext extends ExprContext {
		public ExprContext elem;
		public Token operator;
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public Expr_listContext expr_list() {
			return getRuleContext(Expr_listContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode IN_() { return getToken(SQLParser.IN_, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public In_list_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Is_exprContext extends ExprContext {
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode IS_() { return getToken(SQLParser.IS_, 0); }
		public TerminalNode BOOLEAN_LITERAL() { return getToken(SQLParser.BOOLEAN_LITERAL, 0); }
		public TerminalNode NULL_LITERAL() { return getToken(SQLParser.NULL_LITERAL, 0); }
		public TerminalNode NOT_() { return getToken(SQLParser.NOT_, 0); }
		public TerminalNode DISTINCT_() { return getToken(SQLParser.DISTINCT_, 0); }
		public TerminalNode FROM_() { return getToken(SQLParser.FROM_, 0); }
		public Is_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Case_exprContext extends ExprContext {
		public ExprContext case_clause;
		public ExprContext else_clause;
		public TerminalNode CASE_() { return getToken(SQLParser.CASE_, 0); }
		public TerminalNode END_() { return getToken(SQLParser.END_, 0); }
		public List<When_clauseContext> when_clause() {
			return getRuleContexts(When_clauseContext.class);
		}
		public When_clauseContext when_clause(int i) {
			return getRuleContext(When_clauseContext.class,i);
		}
		public TerminalNode ELSE_() { return getToken(SQLParser.ELSE_, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public Case_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Function_exprContext extends ExprContext {
		public Function_callContext function_call() {
			return getRuleContext(Function_callContext.class,0);
		}
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Function_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Blob_literal_exprContext extends ExprContext {
		public TerminalNode BLOB_LITERAL() { return getToken(SQLParser.BLOB_LITERAL, 0); }
		public Type_castContext type_cast() {
			return getRuleContext(Type_castContext.class,0);
		}
		public Blob_literal_exprContext(ExprContext ctx) { copyFrom(ctx); }
	}

	public final ExprContext expr() throws RecognitionException {
		return expr(0);
	}

	private ExprContext expr(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		ExprContext _localctx = new ExprContext(_ctx, _parentState);
		ExprContext _prevctx = _localctx;
		int _startState = 26;
		enterRecursionRule(_localctx, 26, RULE_expr, _p);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(293);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,30,_ctx) ) {
			case 1:
				{
				_localctx = new Text_literal_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(225);
				match(TEXT_LITERAL);
				setState(227);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,16,_ctx) ) {
				case 1:
					{
					setState(226);
					type_cast();
					}
					break;
				}
				}
				break;
			case 2:
				{
				_localctx = new Boolean_literal_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(229);
				match(BOOLEAN_LITERAL);
				setState(231);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,17,_ctx) ) {
				case 1:
					{
					setState(230);
					type_cast();
					}
					break;
				}
				}
				break;
			case 3:
				{
				_localctx = new Numeric_literal_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(233);
				match(NUMERIC_LITERAL);
				setState(235);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,18,_ctx) ) {
				case 1:
					{
					setState(234);
					type_cast();
					}
					break;
				}
				}
				break;
			case 4:
				{
				_localctx = new Null_literal_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(237);
				match(NULL_LITERAL);
				setState(239);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,19,_ctx) ) {
				case 1:
					{
					setState(238);
					type_cast();
					}
					break;
				}
				}
				break;
			case 5:
				{
				_localctx = new Blob_literal_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(241);
				match(BLOB_LITERAL);
				setState(243);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,20,_ctx) ) {
				case 1:
					{
					setState(242);
					type_cast();
					}
					break;
				}
				}
				break;
			case 6:
				{
				_localctx = new Variable_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(245);
				variable();
				setState(247);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,21,_ctx) ) {
				case 1:
					{
					setState(246);
					type_cast();
					}
					break;
				}
				}
				break;
			case 7:
				{
				_localctx = new Column_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(249);
				column_ref();
				setState(251);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,22,_ctx) ) {
				case 1:
					{
					setState(250);
					type_cast();
					}
					break;
				}
				}
				break;
			case 8:
				{
				_localctx = new Unary_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(253);
				((Unary_exprContext)_localctx).operator = _input.LT(1);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
					((Unary_exprContext)_localctx).operator = (Token)_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(254);
				expr(19);
				}
				break;
			case 9:
				{
				_localctx = new Parenthesized_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(255);
				match(OPEN_PAR);
				setState(256);
				expr(0);
				setState(257);
				match(CLOSE_PAR);
				setState(259);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,23,_ctx) ) {
				case 1:
					{
					setState(258);
					type_cast();
					}
					break;
				}
				}
				break;
			case 10:
				{
				_localctx = new Subquery_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(265);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==EXISTS_ || _la==NOT_) {
					{
					setState(262);
					_errHandler.sync(this);
					_la = _input.LA(1);
					if (_la==NOT_) {
						{
						setState(261);
						match(NOT_);
						}
					}

					setState(264);
					match(EXISTS_);
					}
				}

				setState(267);
				subquery();
				}
				break;
			case 11:
				{
				_localctx = new Case_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(268);
				match(CASE_);
				setState(270);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 2305851805575154696L) != 0) || ((((_la - 65)) & ~0x3f) == 0 && ((1L << (_la - 65)) & 532677121L) != 0)) {
					{
					setState(269);
					((Case_exprContext)_localctx).case_clause = expr(0);
					}
				}

				setState(273); 
				_errHandler.sync(this);
				_la = _input.LA(1);
				do {
					{
					{
					setState(272);
					when_clause();
					}
					}
					setState(275); 
					_errHandler.sync(this);
					_la = _input.LA(1);
				} while ( _la==WHEN_ );
				setState(279);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==ELSE_) {
					{
					setState(277);
					match(ELSE_);
					setState(278);
					((Case_exprContext)_localctx).else_clause = expr(0);
					}
				}

				setState(281);
				match(END_);
				}
				break;
			case 12:
				{
				_localctx = new Expr_list_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(283);
				match(OPEN_PAR);
				setState(284);
				expr_list();
				setState(285);
				match(CLOSE_PAR);
				}
				break;
			case 13:
				{
				_localctx = new Function_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(287);
				function_call();
				setState(289);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,29,_ctx) ) {
				case 1:
					{
					setState(288);
					type_cast();
					}
					break;
				}
				}
				break;
			case 14:
				{
				_localctx = new Logical_not_exprContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(291);
				match(NOT_);
				setState(292);
				expr(3);
				}
				break;
			}
			_ctx.stop = _input.LT(-1);
			setState(364);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,39,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(362);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,38,_ctx) ) {
					case 1:
						{
						_localctx = new Arithmetic_exprContext(new ExprContext(_parentctx, _parentState));
						((Arithmetic_exprContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(295);
						if (!(precpred(_ctx, 12))) throw new FailedPredicateException(this, "precpred(_ctx, 12)");
						setState(296);
						((Arithmetic_exprContext)_localctx).operator = _input.LT(1);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 12800L) != 0)) ) {
							((Arithmetic_exprContext)_localctx).operator = (Token)_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(297);
						((Arithmetic_exprContext)_localctx).right = expr(13);
						}
						break;
					case 2:
						{
						_localctx = new Arithmetic_exprContext(new ExprContext(_parentctx, _parentState));
						((Arithmetic_exprContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(298);
						if (!(precpred(_ctx, 11))) throw new FailedPredicateException(this, "precpred(_ctx, 11)");
						setState(299);
						((Arithmetic_exprContext)_localctx).operator = _input.LT(1);
						_la = _input.LA(1);
						if ( !(_la==PLUS || _la==MINUS) ) {
							((Arithmetic_exprContext)_localctx).operator = (Token)_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(300);
						((Arithmetic_exprContext)_localctx).right = expr(12);
						}
						break;
					case 3:
						{
						_localctx = new Between_exprContext(new ExprContext(_parentctx, _parentState));
						((Between_exprContext)_localctx).elem = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(301);
						if (!(precpred(_ctx, 8))) throw new FailedPredicateException(this, "precpred(_ctx, 8)");
						setState(303);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==NOT_) {
							{
							setState(302);
							match(NOT_);
							}
						}

						setState(305);
						((Between_exprContext)_localctx).operator = match(BETWEEN_);
						setState(306);
						((Between_exprContext)_localctx).low = expr(0);
						setState(307);
						match(AND_);
						setState(308);
						((Between_exprContext)_localctx).high = expr(9);
						}
						break;
					case 4:
						{
						_localctx = new Comparison_exprContext(new ExprContext(_parentctx, _parentState));
						((Comparison_exprContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(310);
						if (!(precpred(_ctx, 6))) throw new FailedPredicateException(this, "precpred(_ctx, 6)");
						setState(311);
						comparisonOperator();
						setState(312);
						((Comparison_exprContext)_localctx).right = expr(7);
						}
						break;
					case 5:
						{
						_localctx = new Logical_binary_exprContext(new ExprContext(_parentctx, _parentState));
						((Logical_binary_exprContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(314);
						if (!(precpred(_ctx, 2))) throw new FailedPredicateException(this, "precpred(_ctx, 2)");
						setState(315);
						((Logical_binary_exprContext)_localctx).operator = match(AND_);
						setState(316);
						((Logical_binary_exprContext)_localctx).right = expr(3);
						}
						break;
					case 6:
						{
						_localctx = new Logical_binary_exprContext(new ExprContext(_parentctx, _parentState));
						((Logical_binary_exprContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(317);
						if (!(precpred(_ctx, 1))) throw new FailedPredicateException(this, "precpred(_ctx, 1)");
						setState(318);
						((Logical_binary_exprContext)_localctx).operator = match(OR_);
						setState(319);
						((Logical_binary_exprContext)_localctx).right = expr(2);
						}
						break;
					case 7:
						{
						_localctx = new Collate_exprContext(new ExprContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(320);
						if (!(precpred(_ctx, 18))) throw new FailedPredicateException(this, "precpred(_ctx, 18)");
						setState(321);
						match(COLLATE_);
						setState(322);
						collation_name();
						}
						break;
					case 8:
						{
						_localctx = new In_subquery_exprContext(new ExprContext(_parentctx, _parentState));
						((In_subquery_exprContext)_localctx).elem = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(323);
						if (!(precpred(_ctx, 10))) throw new FailedPredicateException(this, "precpred(_ctx, 10)");
						setState(325);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==NOT_) {
							{
							setState(324);
							match(NOT_);
							}
						}

						setState(327);
						((In_subquery_exprContext)_localctx).operator = match(IN_);
						setState(328);
						subquery();
						}
						break;
					case 9:
						{
						_localctx = new In_list_exprContext(new ExprContext(_parentctx, _parentState));
						((In_list_exprContext)_localctx).elem = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(329);
						if (!(precpred(_ctx, 9))) throw new FailedPredicateException(this, "precpred(_ctx, 9)");
						setState(331);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==NOT_) {
							{
							setState(330);
							match(NOT_);
							}
						}

						setState(333);
						((In_list_exprContext)_localctx).operator = match(IN_);
						setState(334);
						match(OPEN_PAR);
						setState(335);
						expr_list();
						setState(336);
						match(CLOSE_PAR);
						}
						break;
					case 10:
						{
						_localctx = new Like_exprContext(new ExprContext(_parentctx, _parentState));
						((Like_exprContext)_localctx).elem = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(338);
						if (!(precpred(_ctx, 7))) throw new FailedPredicateException(this, "precpred(_ctx, 7)");
						setState(340);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==NOT_) {
							{
							setState(339);
							match(NOT_);
							}
						}

						setState(342);
						((Like_exprContext)_localctx).operator = match(LIKE_);
						setState(343);
						((Like_exprContext)_localctx).pattern = expr(0);
						setState(346);
						_errHandler.sync(this);
						switch ( getInterpreter().adaptivePredict(_input,35,_ctx) ) {
						case 1:
							{
							setState(344);
							match(ESCAPE_);
							setState(345);
							((Like_exprContext)_localctx).escape = expr(0);
							}
							break;
						}
						}
						break;
					case 11:
						{
						_localctx = new Is_exprContext(new ExprContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(348);
						if (!(precpred(_ctx, 5))) throw new FailedPredicateException(this, "precpred(_ctx, 5)");
						setState(349);
						match(IS_);
						setState(351);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==NOT_) {
							{
							setState(350);
							match(NOT_);
							}
						}

						setState(358);
						_errHandler.sync(this);
						switch (_input.LA(1)) {
						case DISTINCT_:
							{
							{
							setState(353);
							match(DISTINCT_);
							setState(354);
							match(FROM_);
							setState(355);
							expr(0);
							}
							}
							break;
						case BOOLEAN_LITERAL:
							{
							setState(356);
							match(BOOLEAN_LITERAL);
							}
							break;
						case NULL_LITERAL:
							{
							setState(357);
							match(NULL_LITERAL);
							}
							break;
						default:
							throw new NoViableAltException(this);
						}
						}
						break;
					case 12:
						{
						_localctx = new Null_exprContext(new ExprContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expr);
						setState(360);
						if (!(precpred(_ctx, 4))) throw new FailedPredicateException(this, "precpred(_ctx, 4)");
						setState(361);
						_la = _input.LA(1);
						if ( !(_la==ISNULL_ || _la==NOTNULL_) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						}
						break;
					}
					} 
				}
				setState(366);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,39,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			unrollRecursionContexts(_parentctx);
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SubqueryContext extends ParserRuleContext {
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public Select_coreContext select_core() {
			return getRuleContext(Select_coreContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public SubqueryContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_subquery; }
	}

	public final SubqueryContext subquery() throws RecognitionException {
		SubqueryContext _localctx = new SubqueryContext(_ctx, getState());
		enterRule(_localctx, 28, RULE_subquery);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(367);
			match(OPEN_PAR);
			setState(368);
			select_core();
			setState(369);
			match(CLOSE_PAR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Expr_listContext extends ParserRuleContext {
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Expr_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expr_list; }
	}

	public final Expr_listContext expr_list() throws RecognitionException {
		Expr_listContext _localctx = new Expr_listContext(_ctx, getState());
		enterRule(_localctx, 30, RULE_expr_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(371);
			expr(0);
			setState(376);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(372);
				match(COMMA);
				setState(373);
				expr(0);
				}
				}
				setState(378);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ComparisonOperatorContext extends ParserRuleContext {
		public TerminalNode LT() { return getToken(SQLParser.LT, 0); }
		public TerminalNode LT_EQ() { return getToken(SQLParser.LT_EQ, 0); }
		public TerminalNode GT() { return getToken(SQLParser.GT, 0); }
		public TerminalNode GT_EQ() { return getToken(SQLParser.GT_EQ, 0); }
		public TerminalNode ASSIGN() { return getToken(SQLParser.ASSIGN, 0); }
		public TerminalNode NOT_EQ1() { return getToken(SQLParser.NOT_EQ1, 0); }
		public TerminalNode NOT_EQ2() { return getToken(SQLParser.NOT_EQ2, 0); }
		public ComparisonOperatorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_comparisonOperator; }
	}

	public final ComparisonOperatorContext comparisonOperator() throws RecognitionException {
		ComparisonOperatorContext _localctx = new ComparisonOperatorContext(_ctx, getState());
		enterRule(_localctx, 32, RULE_comparisonOperator);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(379);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 1032448L) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Cast_typeContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public TerminalNode L_BRACKET() { return getToken(SQLParser.L_BRACKET, 0); }
		public TerminalNode R_BRACKET() { return getToken(SQLParser.R_BRACKET, 0); }
		public Cast_typeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_cast_type; }
	}

	public final Cast_typeContext cast_type() throws RecognitionException {
		Cast_typeContext _localctx = new Cast_typeContext(_ctx, getState());
		enterRule(_localctx, 34, RULE_cast_type);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(381);
			match(IDENTIFIER);
			setState(384);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,41,_ctx) ) {
			case 1:
				{
				setState(382);
				match(L_BRACKET);
				setState(383);
				match(R_BRACKET);
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Type_castContext extends ParserRuleContext {
		public TerminalNode TYPE_CAST() { return getToken(SQLParser.TYPE_CAST, 0); }
		public Cast_typeContext cast_type() {
			return getRuleContext(Cast_typeContext.class,0);
		}
		public Type_castContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_type_cast; }
	}

	public final Type_castContext type_cast() throws RecognitionException {
		Type_castContext _localctx = new Type_castContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_type_cast);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(386);
			match(TYPE_CAST);
			setState(387);
			cast_type();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Value_rowContext extends ParserRuleContext {
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Value_rowContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_value_row; }
	}

	public final Value_rowContext value_row() throws RecognitionException {
		Value_rowContext _localctx = new Value_rowContext(_ctx, getState());
		enterRule(_localctx, 38, RULE_value_row);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(389);
			match(OPEN_PAR);
			setState(390);
			expr(0);
			setState(395);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(391);
				match(COMMA);
				setState(392);
				expr(0);
				}
				}
				setState(397);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(398);
			match(CLOSE_PAR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Values_clauseContext extends ParserRuleContext {
		public TerminalNode VALUES_() { return getToken(SQLParser.VALUES_, 0); }
		public List<Value_rowContext> value_row() {
			return getRuleContexts(Value_rowContext.class);
		}
		public Value_rowContext value_row(int i) {
			return getRuleContext(Value_rowContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Values_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_values_clause; }
	}

	public final Values_clauseContext values_clause() throws RecognitionException {
		Values_clauseContext _localctx = new Values_clauseContext(_ctx, getState());
		enterRule(_localctx, 40, RULE_values_clause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(400);
			match(VALUES_);
			setState(401);
			value_row();
			setState(406);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(402);
				match(COMMA);
				setState(403);
				value_row();
				}
				}
				setState(408);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Insert_coreContext extends ParserRuleContext {
		public TerminalNode INSERT_() { return getToken(SQLParser.INSERT_, 0); }
		public TerminalNode INTO_() { return getToken(SQLParser.INTO_, 0); }
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public Values_clauseContext values_clause() {
			return getRuleContext(Values_clauseContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public Table_aliasContext table_alias() {
			return getRuleContext(Table_aliasContext.class,0);
		}
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public List<Column_nameContext> column_name() {
			return getRuleContexts(Column_nameContext.class);
		}
		public Column_nameContext column_name(int i) {
			return getRuleContext(Column_nameContext.class,i);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public Upsert_clauseContext upsert_clause() {
			return getRuleContext(Upsert_clauseContext.class,0);
		}
		public Returning_clauseContext returning_clause() {
			return getRuleContext(Returning_clauseContext.class,0);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Insert_coreContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_insert_core; }
	}

	public final Insert_coreContext insert_core() throws RecognitionException {
		Insert_coreContext _localctx = new Insert_coreContext(_ctx, getState());
		enterRule(_localctx, 42, RULE_insert_core);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(409);
			match(INSERT_);
			setState(410);
			match(INTO_);
			setState(411);
			table_name();
			setState(414);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==AS_) {
				{
				setState(412);
				match(AS_);
				setState(413);
				table_alias();
				}
			}

			setState(427);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OPEN_PAR) {
				{
				setState(416);
				match(OPEN_PAR);
				setState(417);
				column_name();
				setState(422);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(418);
					match(COMMA);
					setState(419);
					column_name();
					}
					}
					setState(424);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(425);
				match(CLOSE_PAR);
				}
			}

			setState(429);
			values_clause();
			setState(431);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==ON_) {
				{
				setState(430);
				upsert_clause();
				}
			}

			setState(434);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==RETURNING_) {
				{
				setState(433);
				returning_clause();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Insert_stmtContext extends ParserRuleContext {
		public Insert_coreContext insert_core() {
			return getRuleContext(Insert_coreContext.class,0);
		}
		public Common_table_stmtContext common_table_stmt() {
			return getRuleContext(Common_table_stmtContext.class,0);
		}
		public Insert_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_insert_stmt; }
	}

	public final Insert_stmtContext insert_stmt() throws RecognitionException {
		Insert_stmtContext _localctx = new Insert_stmtContext(_ctx, getState());
		enterRule(_localctx, 44, RULE_insert_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(437);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WITH_) {
				{
				setState(436);
				common_table_stmt();
				}
			}

			setState(439);
			insert_core();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Returning_clauseContext extends ParserRuleContext {
		public TerminalNode RETURNING_() { return getToken(SQLParser.RETURNING_, 0); }
		public List<Returning_clause_result_columnContext> returning_clause_result_column() {
			return getRuleContexts(Returning_clause_result_columnContext.class);
		}
		public Returning_clause_result_columnContext returning_clause_result_column(int i) {
			return getRuleContext(Returning_clause_result_columnContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Returning_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returning_clause; }
	}

	public final Returning_clauseContext returning_clause() throws RecognitionException {
		Returning_clauseContext _localctx = new Returning_clauseContext(_ctx, getState());
		enterRule(_localctx, 46, RULE_returning_clause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(441);
			match(RETURNING_);
			setState(442);
			returning_clause_result_column();
			setState(447);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(443);
				match(COMMA);
				setState(444);
				returning_clause_result_column();
				}
				}
				setState(449);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Upsert_updateContext extends ParserRuleContext {
		public TerminalNode ASSIGN() { return getToken(SQLParser.ASSIGN, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Column_nameContext column_name() {
			return getRuleContext(Column_nameContext.class,0);
		}
		public Column_name_listContext column_name_list() {
			return getRuleContext(Column_name_listContext.class,0);
		}
		public Upsert_updateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_upsert_update; }
	}

	public final Upsert_updateContext upsert_update() throws RecognitionException {
		Upsert_updateContext _localctx = new Upsert_updateContext(_ctx, getState());
		enterRule(_localctx, 48, RULE_upsert_update);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(452);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				{
				setState(450);
				column_name();
				}
				break;
			case OPEN_PAR:
				{
				setState(451);
				column_name_list();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(454);
			match(ASSIGN);
			setState(455);
			expr(0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Upsert_clauseContext extends ParserRuleContext {
		public ExprContext target_expr;
		public ExprContext update_expr;
		public TerminalNode ON_() { return getToken(SQLParser.ON_, 0); }
		public TerminalNode CONFLICT_() { return getToken(SQLParser.CONFLICT_, 0); }
		public TerminalNode DO_() { return getToken(SQLParser.DO_, 0); }
		public TerminalNode NOTHING_() { return getToken(SQLParser.NOTHING_, 0); }
		public TerminalNode UPDATE_() { return getToken(SQLParser.UPDATE_, 0); }
		public TerminalNode SET_() { return getToken(SQLParser.SET_, 0); }
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public List<Indexed_columnContext> indexed_column() {
			return getRuleContexts(Indexed_columnContext.class);
		}
		public Indexed_columnContext indexed_column(int i) {
			return getRuleContext(Indexed_columnContext.class,i);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public List<Upsert_updateContext> upsert_update() {
			return getRuleContexts(Upsert_updateContext.class);
		}
		public Upsert_updateContext upsert_update(int i) {
			return getRuleContext(Upsert_updateContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public List<TerminalNode> WHERE_() { return getTokens(SQLParser.WHERE_); }
		public TerminalNode WHERE_(int i) {
			return getToken(SQLParser.WHERE_, i);
		}
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public Upsert_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_upsert_clause; }
	}

	public final Upsert_clauseContext upsert_clause() throws RecognitionException {
		Upsert_clauseContext _localctx = new Upsert_clauseContext(_ctx, getState());
		enterRule(_localctx, 50, RULE_upsert_clause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(457);
			match(ON_);
			setState(458);
			match(CONFLICT_);
			setState(473);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OPEN_PAR) {
				{
				setState(459);
				match(OPEN_PAR);
				setState(460);
				indexed_column();
				setState(465);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(461);
					match(COMMA);
					setState(462);
					indexed_column();
					}
					}
					setState(467);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(468);
				match(CLOSE_PAR);
				setState(471);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==WHERE_) {
					{
					setState(469);
					match(WHERE_);
					setState(470);
					((Upsert_clauseContext)_localctx).target_expr = expr(0);
					}
				}

				}
			}

			setState(475);
			match(DO_);
			setState(491);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case NOTHING_:
				{
				setState(476);
				match(NOTHING_);
				}
				break;
			case UPDATE_:
				{
				setState(477);
				match(UPDATE_);
				setState(478);
				match(SET_);
				{
				setState(479);
				upsert_update();
				setState(484);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(480);
					match(COMMA);
					setState(481);
					upsert_update();
					}
					}
					setState(486);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(489);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==WHERE_) {
					{
					setState(487);
					match(WHERE_);
					setState(488);
					((Upsert_clauseContext)_localctx).update_expr = expr(0);
					}
				}

				}
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Select_coreContext extends ParserRuleContext {
		public List<Simple_selectContext> simple_select() {
			return getRuleContexts(Simple_selectContext.class);
		}
		public Simple_selectContext simple_select(int i) {
			return getRuleContext(Simple_selectContext.class,i);
		}
		public List<Compound_operatorContext> compound_operator() {
			return getRuleContexts(Compound_operatorContext.class);
		}
		public Compound_operatorContext compound_operator(int i) {
			return getRuleContext(Compound_operatorContext.class,i);
		}
		public Order_by_stmtContext order_by_stmt() {
			return getRuleContext(Order_by_stmtContext.class,0);
		}
		public Limit_stmtContext limit_stmt() {
			return getRuleContext(Limit_stmtContext.class,0);
		}
		public Select_coreContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_select_core; }
	}

	public final Select_coreContext select_core() throws RecognitionException {
		Select_coreContext _localctx = new Select_coreContext(_ctx, getState());
		enterRule(_localctx, 52, RULE_select_core);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(493);
			simple_select();
			setState(499);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 42)) & ~0x3f) == 0 && ((1L << (_la - 42)) & 274877908993L) != 0)) {
				{
				{
				setState(494);
				compound_operator();
				setState(495);
				simple_select();
				}
				}
				setState(501);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(503);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==ORDER_) {
				{
				setState(502);
				order_by_stmt();
				}
			}

			setState(506);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LIMIT_) {
				{
				setState(505);
				limit_stmt();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Select_stmtContext extends ParserRuleContext {
		public Select_coreContext select_core() {
			return getRuleContext(Select_coreContext.class,0);
		}
		public Common_table_stmtContext common_table_stmt() {
			return getRuleContext(Common_table_stmtContext.class,0);
		}
		public Select_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_select_stmt; }
	}

	public final Select_stmtContext select_stmt() throws RecognitionException {
		Select_stmtContext _localctx = new Select_stmtContext(_ctx, getState());
		enterRule(_localctx, 54, RULE_select_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(509);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WITH_) {
				{
				setState(508);
				common_table_stmt();
				}
			}

			setState(511);
			select_core();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Join_relationContext extends ParserRuleContext {
		public Table_or_subqueryContext right_relation;
		public Join_operatorContext join_operator() {
			return getRuleContext(Join_operatorContext.class,0);
		}
		public Join_constraintContext join_constraint() {
			return getRuleContext(Join_constraintContext.class,0);
		}
		public Table_or_subqueryContext table_or_subquery() {
			return getRuleContext(Table_or_subqueryContext.class,0);
		}
		public Join_relationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_join_relation; }
	}

	public final Join_relationContext join_relation() throws RecognitionException {
		Join_relationContext _localctx = new Join_relationContext(_ctx, getState());
		enterRule(_localctx, 56, RULE_join_relation);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(513);
			join_operator();
			setState(514);
			((Join_relationContext)_localctx).right_relation = table_or_subquery();
			setState(515);
			join_constraint();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelationContext extends ParserRuleContext {
		public Table_or_subqueryContext table_or_subquery() {
			return getRuleContext(Table_or_subqueryContext.class,0);
		}
		public List<Join_relationContext> join_relation() {
			return getRuleContexts(Join_relationContext.class);
		}
		public Join_relationContext join_relation(int i) {
			return getRuleContext(Join_relationContext.class,i);
		}
		public RelationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relation; }
	}

	public final RelationContext relation() throws RecognitionException {
		RelationContext _localctx = new RelationContext(_ctx, getState());
		enterRule(_localctx, 58, RULE_relation);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(517);
			table_or_subquery();
			setState(521);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 47)) & ~0x3f) == 0 && ((1L << (_la - 47)) & 536881169L) != 0)) {
				{
				{
				setState(518);
				join_relation();
				}
				}
				setState(523);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Simple_selectContext extends ParserRuleContext {
		public ExprContext whereExpr;
		public ExprContext expr;
		public List<ExprContext> groupByExpr = new ArrayList<ExprContext>();
		public ExprContext havingExpr;
		public TerminalNode SELECT_() { return getToken(SQLParser.SELECT_, 0); }
		public List<Result_columnContext> result_column() {
			return getRuleContexts(Result_columnContext.class);
		}
		public Result_columnContext result_column(int i) {
			return getRuleContext(Result_columnContext.class,i);
		}
		public TerminalNode DISTINCT_() { return getToken(SQLParser.DISTINCT_, 0); }
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public TerminalNode FROM_() { return getToken(SQLParser.FROM_, 0); }
		public RelationContext relation() {
			return getRuleContext(RelationContext.class,0);
		}
		public TerminalNode WHERE_() { return getToken(SQLParser.WHERE_, 0); }
		public TerminalNode GROUP_() { return getToken(SQLParser.GROUP_, 0); }
		public TerminalNode BY_() { return getToken(SQLParser.BY_, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode HAVING_() { return getToken(SQLParser.HAVING_, 0); }
		public Simple_selectContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_simple_select; }
	}

	public final Simple_selectContext simple_select() throws RecognitionException {
		Simple_selectContext _localctx = new Simple_selectContext(_ctx, getState());
		enterRule(_localctx, 60, RULE_simple_select);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(524);
			match(SELECT_);
			setState(526);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==DISTINCT_) {
				{
				setState(525);
				match(DISTINCT_);
				}
			}

			setState(528);
			result_column();
			setState(533);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(529);
				match(COMMA);
				setState(530);
				result_column();
				}
				}
				setState(535);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(538);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==FROM_) {
				{
				setState(536);
				match(FROM_);
				setState(537);
				relation();
				}
			}

			setState(542);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WHERE_) {
				{
				setState(540);
				match(WHERE_);
				setState(541);
				((Simple_selectContext)_localctx).whereExpr = expr(0);
				}
			}

			setState(558);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==GROUP_) {
				{
				setState(544);
				match(GROUP_);
				setState(545);
				match(BY_);
				setState(546);
				((Simple_selectContext)_localctx).expr = expr(0);
				((Simple_selectContext)_localctx).groupByExpr.add(((Simple_selectContext)_localctx).expr);
				setState(551);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(547);
					match(COMMA);
					setState(548);
					((Simple_selectContext)_localctx).expr = expr(0);
					((Simple_selectContext)_localctx).groupByExpr.add(((Simple_selectContext)_localctx).expr);
					}
					}
					setState(553);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(556);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==HAVING_) {
					{
					setState(554);
					match(HAVING_);
					setState(555);
					((Simple_selectContext)_localctx).havingExpr = expr(0);
					}
				}

				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Table_or_subqueryContext extends ParserRuleContext {
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public Table_aliasContext table_alias() {
			return getRuleContext(Table_aliasContext.class,0);
		}
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public Select_coreContext select_core() {
			return getRuleContext(Select_coreContext.class,0);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public Table_or_subqueryContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table_or_subquery; }
	}

	public final Table_or_subqueryContext table_or_subquery() throws RecognitionException {
		Table_or_subqueryContext _localctx = new Table_or_subqueryContext(_ctx, getState());
		enterRule(_localctx, 62, RULE_table_or_subquery);
		int _la;
		try {
			setState(572);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(560);
				table_name();
				setState(563);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS_) {
					{
					setState(561);
					match(AS_);
					setState(562);
					table_alias();
					}
				}

				}
				break;
			case OPEN_PAR:
				enterOuterAlt(_localctx, 2);
				{
				setState(565);
				match(OPEN_PAR);
				setState(566);
				select_core();
				setState(567);
				match(CLOSE_PAR);
				setState(570);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS_) {
					{
					setState(568);
					match(AS_);
					setState(569);
					table_alias();
					}
				}

				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Result_columnContext extends ParserRuleContext {
		public TerminalNode STAR() { return getToken(SQLParser.STAR, 0); }
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public TerminalNode DOT() { return getToken(SQLParser.DOT, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public Column_aliasContext column_alias() {
			return getRuleContext(Column_aliasContext.class,0);
		}
		public Result_columnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_result_column; }
	}

	public final Result_columnContext result_column() throws RecognitionException {
		Result_columnContext _localctx = new Result_columnContext(_ctx, getState());
		enterRule(_localctx, 64, RULE_result_column);
		int _la;
		try {
			setState(584);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,74,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(574);
				match(STAR);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(575);
				table_name();
				setState(576);
				match(DOT);
				setState(577);
				match(STAR);
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(579);
				expr(0);
				setState(582);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS_) {
					{
					setState(580);
					match(AS_);
					setState(581);
					column_alias();
					}
				}

				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Returning_clause_result_columnContext extends ParserRuleContext {
		public TerminalNode STAR() { return getToken(SQLParser.STAR, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public Column_aliasContext column_alias() {
			return getRuleContext(Column_aliasContext.class,0);
		}
		public Returning_clause_result_columnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returning_clause_result_column; }
	}

	public final Returning_clause_result_columnContext returning_clause_result_column() throws RecognitionException {
		Returning_clause_result_columnContext _localctx = new Returning_clause_result_columnContext(_ctx, getState());
		enterRule(_localctx, 66, RULE_returning_clause_result_column);
		int _la;
		try {
			setState(592);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STAR:
				enterOuterAlt(_localctx, 1);
				{
				setState(586);
				match(STAR);
				}
				break;
			case OPEN_PAR:
			case PLUS:
			case MINUS:
			case CASE_:
			case EXISTS_:
			case LIKE_:
			case NOT_:
			case REPLACE_:
			case BOOLEAN_LITERAL:
			case NUMERIC_LITERAL:
			case BLOB_LITERAL:
			case TEXT_LITERAL:
			case NULL_LITERAL:
			case IDENTIFIER:
			case BIND_PARAMETER:
				enterOuterAlt(_localctx, 2);
				{
				setState(587);
				expr(0);
				setState(590);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS_) {
					{
					setState(588);
					match(AS_);
					setState(589);
					column_alias();
					}
				}

				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Join_operatorContext extends ParserRuleContext {
		public TerminalNode JOIN_() { return getToken(SQLParser.JOIN_, 0); }
		public TerminalNode INNER_() { return getToken(SQLParser.INNER_, 0); }
		public TerminalNode LEFT_() { return getToken(SQLParser.LEFT_, 0); }
		public TerminalNode RIGHT_() { return getToken(SQLParser.RIGHT_, 0); }
		public TerminalNode FULL_() { return getToken(SQLParser.FULL_, 0); }
		public TerminalNode OUTER_() { return getToken(SQLParser.OUTER_, 0); }
		public Join_operatorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_join_operator; }
	}

	public final Join_operatorContext join_operator() throws RecognitionException {
		Join_operatorContext _localctx = new Join_operatorContext(_ctx, getState());
		enterRule(_localctx, 68, RULE_join_operator);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(599);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case FULL_:
			case LEFT_:
			case RIGHT_:
				{
				setState(594);
				_la = _input.LA(1);
				if ( !(((((_la - 47)) & ~0x3f) == 0 && ((1L << (_la - 47)) & 536879105L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(596);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==OUTER_) {
					{
					setState(595);
					match(OUTER_);
					}
				}

				}
				break;
			case INNER_:
				{
				setState(598);
				match(INNER_);
				}
				break;
			case JOIN_:
				break;
			default:
				break;
			}
			setState(601);
			match(JOIN_);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Join_constraintContext extends ParserRuleContext {
		public TerminalNode ON_() { return getToken(SQLParser.ON_, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Join_constraintContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_join_constraint; }
	}

	public final Join_constraintContext join_constraint() throws RecognitionException {
		Join_constraintContext _localctx = new Join_constraintContext(_ctx, getState());
		enterRule(_localctx, 70, RULE_join_constraint);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(603);
			match(ON_);
			setState(604);
			expr(0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Compound_operatorContext extends ParserRuleContext {
		public TerminalNode UNION_() { return getToken(SQLParser.UNION_, 0); }
		public TerminalNode ALL_() { return getToken(SQLParser.ALL_, 0); }
		public TerminalNode INTERSECT_() { return getToken(SQLParser.INTERSECT_, 0); }
		public TerminalNode EXCEPT_() { return getToken(SQLParser.EXCEPT_, 0); }
		public Compound_operatorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_compound_operator; }
	}

	public final Compound_operatorContext compound_operator() throws RecognitionException {
		Compound_operatorContext _localctx = new Compound_operatorContext(_ctx, getState());
		enterRule(_localctx, 72, RULE_compound_operator);
		int _la;
		try {
			setState(612);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case UNION_:
				enterOuterAlt(_localctx, 1);
				{
				setState(606);
				match(UNION_);
				setState(608);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==ALL_) {
					{
					setState(607);
					match(ALL_);
					}
				}

				}
				break;
			case INTERSECT_:
				enterOuterAlt(_localctx, 2);
				{
				setState(610);
				match(INTERSECT_);
				}
				break;
			case EXCEPT_:
				enterOuterAlt(_localctx, 3);
				{
				setState(611);
				match(EXCEPT_);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Update_set_subclauseContext extends ParserRuleContext {
		public TerminalNode ASSIGN() { return getToken(SQLParser.ASSIGN, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Column_nameContext column_name() {
			return getRuleContext(Column_nameContext.class,0);
		}
		public Column_name_listContext column_name_list() {
			return getRuleContext(Column_name_listContext.class,0);
		}
		public Update_set_subclauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_update_set_subclause; }
	}

	public final Update_set_subclauseContext update_set_subclause() throws RecognitionException {
		Update_set_subclauseContext _localctx = new Update_set_subclauseContext(_ctx, getState());
		enterRule(_localctx, 74, RULE_update_set_subclause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(616);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				{
				setState(614);
				column_name();
				}
				break;
			case OPEN_PAR:
				{
				setState(615);
				column_name_list();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(618);
			match(ASSIGN);
			setState(619);
			expr(0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Update_coreContext extends ParserRuleContext {
		public TerminalNode UPDATE_() { return getToken(SQLParser.UPDATE_, 0); }
		public Qualified_table_nameContext qualified_table_name() {
			return getRuleContext(Qualified_table_nameContext.class,0);
		}
		public TerminalNode SET_() { return getToken(SQLParser.SET_, 0); }
		public List<Update_set_subclauseContext> update_set_subclause() {
			return getRuleContexts(Update_set_subclauseContext.class);
		}
		public Update_set_subclauseContext update_set_subclause(int i) {
			return getRuleContext(Update_set_subclauseContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public TerminalNode FROM_() { return getToken(SQLParser.FROM_, 0); }
		public RelationContext relation() {
			return getRuleContext(RelationContext.class,0);
		}
		public TerminalNode WHERE_() { return getToken(SQLParser.WHERE_, 0); }
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Returning_clauseContext returning_clause() {
			return getRuleContext(Returning_clauseContext.class,0);
		}
		public Update_coreContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_update_core; }
	}

	public final Update_coreContext update_core() throws RecognitionException {
		Update_coreContext _localctx = new Update_coreContext(_ctx, getState());
		enterRule(_localctx, 76, RULE_update_core);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(621);
			match(UPDATE_);
			setState(622);
			qualified_table_name();
			setState(623);
			match(SET_);
			setState(624);
			update_set_subclause();
			setState(629);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(625);
				match(COMMA);
				setState(626);
				update_set_subclause();
				}
				}
				setState(631);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(634);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==FROM_) {
				{
				setState(632);
				match(FROM_);
				setState(633);
				relation();
				}
			}

			setState(638);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WHERE_) {
				{
				setState(636);
				match(WHERE_);
				setState(637);
				expr(0);
				}
			}

			setState(641);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==RETURNING_) {
				{
				setState(640);
				returning_clause();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Update_stmtContext extends ParserRuleContext {
		public Update_coreContext update_core() {
			return getRuleContext(Update_coreContext.class,0);
		}
		public Common_table_stmtContext common_table_stmt() {
			return getRuleContext(Common_table_stmtContext.class,0);
		}
		public Update_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_update_stmt; }
	}

	public final Update_stmtContext update_stmt() throws RecognitionException {
		Update_stmtContext _localctx = new Update_stmtContext(_ctx, getState());
		enterRule(_localctx, 78, RULE_update_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(644);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WITH_) {
				{
				setState(643);
				common_table_stmt();
				}
			}

			setState(646);
			update_core();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Column_name_listContext extends ParserRuleContext {
		public TerminalNode OPEN_PAR() { return getToken(SQLParser.OPEN_PAR, 0); }
		public List<Column_nameContext> column_name() {
			return getRuleContexts(Column_nameContext.class);
		}
		public Column_nameContext column_name(int i) {
			return getRuleContext(Column_nameContext.class,i);
		}
		public TerminalNode CLOSE_PAR() { return getToken(SQLParser.CLOSE_PAR, 0); }
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Column_name_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column_name_list; }
	}

	public final Column_name_listContext column_name_list() throws RecognitionException {
		Column_name_listContext _localctx = new Column_name_listContext(_ctx, getState());
		enterRule(_localctx, 80, RULE_column_name_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(648);
			match(OPEN_PAR);
			setState(649);
			column_name();
			setState(654);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(650);
				match(COMMA);
				setState(651);
				column_name();
				}
				}
				setState(656);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(657);
			match(CLOSE_PAR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Qualified_table_nameContext extends ParserRuleContext {
		public Table_nameContext table_name() {
			return getRuleContext(Table_nameContext.class,0);
		}
		public TerminalNode AS_() { return getToken(SQLParser.AS_, 0); }
		public Table_aliasContext table_alias() {
			return getRuleContext(Table_aliasContext.class,0);
		}
		public Qualified_table_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_qualified_table_name; }
	}

	public final Qualified_table_nameContext qualified_table_name() throws RecognitionException {
		Qualified_table_nameContext _localctx = new Qualified_table_nameContext(_ctx, getState());
		enterRule(_localctx, 82, RULE_qualified_table_name);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(659);
			table_name();
			setState(662);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==AS_) {
				{
				setState(660);
				match(AS_);
				setState(661);
				table_alias();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Order_by_stmtContext extends ParserRuleContext {
		public TerminalNode ORDER_() { return getToken(SQLParser.ORDER_, 0); }
		public TerminalNode BY_() { return getToken(SQLParser.BY_, 0); }
		public List<Ordering_termContext> ordering_term() {
			return getRuleContexts(Ordering_termContext.class);
		}
		public Ordering_termContext ordering_term(int i) {
			return getRuleContext(Ordering_termContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(SQLParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(SQLParser.COMMA, i);
		}
		public Order_by_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_order_by_stmt; }
	}

	public final Order_by_stmtContext order_by_stmt() throws RecognitionException {
		Order_by_stmtContext _localctx = new Order_by_stmtContext(_ctx, getState());
		enterRule(_localctx, 84, RULE_order_by_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(664);
			match(ORDER_);
			setState(665);
			match(BY_);
			setState(666);
			ordering_term();
			setState(671);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(667);
				match(COMMA);
				setState(668);
				ordering_term();
				}
				}
				setState(673);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Limit_stmtContext extends ParserRuleContext {
		public TerminalNode LIMIT_() { return getToken(SQLParser.LIMIT_, 0); }
		public List<ExprContext> expr() {
			return getRuleContexts(ExprContext.class);
		}
		public ExprContext expr(int i) {
			return getRuleContext(ExprContext.class,i);
		}
		public TerminalNode OFFSET_() { return getToken(SQLParser.OFFSET_, 0); }
		public Limit_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_limit_stmt; }
	}

	public final Limit_stmtContext limit_stmt() throws RecognitionException {
		Limit_stmtContext _localctx = new Limit_stmtContext(_ctx, getState());
		enterRule(_localctx, 86, RULE_limit_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(674);
			match(LIMIT_);
			setState(675);
			expr(0);
			setState(678);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OFFSET_) {
				{
				setState(676);
				match(OFFSET_);
				setState(677);
				expr(0);
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Ordering_termContext extends ParserRuleContext {
		public ExprContext expr() {
			return getRuleContext(ExprContext.class,0);
		}
		public Asc_descContext asc_desc() {
			return getRuleContext(Asc_descContext.class,0);
		}
		public TerminalNode NULLS_() { return getToken(SQLParser.NULLS_, 0); }
		public TerminalNode FIRST_() { return getToken(SQLParser.FIRST_, 0); }
		public TerminalNode LAST_() { return getToken(SQLParser.LAST_, 0); }
		public Ordering_termContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ordering_term; }
	}

	public final Ordering_termContext ordering_term() throws RecognitionException {
		Ordering_termContext _localctx = new Ordering_termContext(_ctx, getState());
		enterRule(_localctx, 88, RULE_ordering_term);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(680);
			expr(0);
			setState(682);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==ASC_ || _la==DESC_) {
				{
				setState(681);
				asc_desc();
				}
			}

			setState(686);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==NULLS_) {
				{
				setState(684);
				match(NULLS_);
				setState(685);
				_la = _input.LA(1);
				if ( !(_la==FIRST_ || _la==LAST_) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Asc_descContext extends ParserRuleContext {
		public TerminalNode ASC_() { return getToken(SQLParser.ASC_, 0); }
		public TerminalNode DESC_() { return getToken(SQLParser.DESC_, 0); }
		public Asc_descContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_asc_desc; }
	}

	public final Asc_descContext asc_desc() throws RecognitionException {
		Asc_descContext _localctx = new Asc_descContext(_ctx, getState());
		enterRule(_localctx, 90, RULE_asc_desc);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(688);
			_la = _input.LA(1);
			if ( !(_la==ASC_ || _la==DESC_) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Function_keywordContext extends ParserRuleContext {
		public TerminalNode LIKE_() { return getToken(SQLParser.LIKE_, 0); }
		public TerminalNode REPLACE_() { return getToken(SQLParser.REPLACE_, 0); }
		public Function_keywordContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_function_keyword; }
	}

	public final Function_keywordContext function_keyword() throws RecognitionException {
		Function_keywordContext _localctx = new Function_keywordContext(_ctx, getState());
		enterRule(_localctx, 92, RULE_function_keyword);
		try {
			setState(693);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OPEN_PAR:
				enterOuterAlt(_localctx, 1);
				{
				}
				break;
			case LIKE_:
				enterOuterAlt(_localctx, 2);
				{
				setState(691);
				match(LIKE_);
				}
				break;
			case REPLACE_:
				enterOuterAlt(_localctx, 3);
				{
				setState(692);
				match(REPLACE_);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Function_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Function_keywordContext function_keyword() {
			return getRuleContext(Function_keywordContext.class,0);
		}
		public Function_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_function_name; }
	}

	public final Function_nameContext function_name() throws RecognitionException {
		Function_nameContext _localctx = new Function_nameContext(_ctx, getState());
		enterRule(_localctx, 94, RULE_function_name);
		try {
			setState(697);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(695);
				match(IDENTIFIER);
				}
				break;
			case OPEN_PAR:
			case LIKE_:
			case REPLACE_:
				enterOuterAlt(_localctx, 2);
				{
				setState(696);
				function_keyword();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Table_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Table_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table_name; }
	}

	public final Table_nameContext table_name() throws RecognitionException {
		Table_nameContext _localctx = new Table_nameContext(_ctx, getState());
		enterRule(_localctx, 96, RULE_table_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(699);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Table_aliasContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Table_aliasContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table_alias; }
	}

	public final Table_aliasContext table_alias() throws RecognitionException {
		Table_aliasContext _localctx = new Table_aliasContext(_ctx, getState());
		enterRule(_localctx, 98, RULE_table_alias);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(701);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Column_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Column_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column_name; }
	}

	public final Column_nameContext column_name() throws RecognitionException {
		Column_nameContext _localctx = new Column_nameContext(_ctx, getState());
		enterRule(_localctx, 100, RULE_column_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(703);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Column_aliasContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Column_aliasContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column_alias; }
	}

	public final Column_aliasContext column_alias() throws RecognitionException {
		Column_aliasContext _localctx = new Column_aliasContext(_ctx, getState());
		enterRule(_localctx, 102, RULE_column_alias);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(705);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Collation_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Collation_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_collation_name; }
	}

	public final Collation_nameContext collation_name() throws RecognitionException {
		Collation_nameContext _localctx = new Collation_nameContext(_ctx, getState());
		enterRule(_localctx, 104, RULE_collation_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(707);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Index_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(SQLParser.IDENTIFIER, 0); }
		public Index_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_index_name; }
	}

	public final Index_nameContext index_name() throws RecognitionException {
		Index_nameContext _localctx = new Index_nameContext(_ctx, getState());
		enterRule(_localctx, 106, RULE_index_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(709);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public boolean sempred(RuleContext _localctx, int ruleIndex, int predIndex) {
		switch (ruleIndex) {
		case 13:
			return expr_sempred((ExprContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean expr_sempred(ExprContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 12);
		case 1:
			return precpred(_ctx, 11);
		case 2:
			return precpred(_ctx, 8);
		case 3:
			return precpred(_ctx, 6);
		case 4:
			return precpred(_ctx, 2);
		case 5:
			return precpred(_ctx, 1);
		case 6:
			return precpred(_ctx, 18);
		case 7:
			return precpred(_ctx, 10);
		case 8:
			return precpred(_ctx, 9);
		case 9:
			return precpred(_ctx, 7);
		case 10:
			return precpred(_ctx, 5);
		case 11:
			return precpred(_ctx, 4);
		}
		return true;
	}

	public static final String _serializedATN =
		"\u0004\u0001a\u02c8\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007\u000f"+
		"\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007\u0012"+
		"\u0002\u0013\u0007\u0013\u0002\u0014\u0007\u0014\u0002\u0015\u0007\u0015"+
		"\u0002\u0016\u0007\u0016\u0002\u0017\u0007\u0017\u0002\u0018\u0007\u0018"+
		"\u0002\u0019\u0007\u0019\u0002\u001a\u0007\u001a\u0002\u001b\u0007\u001b"+
		"\u0002\u001c\u0007\u001c\u0002\u001d\u0007\u001d\u0002\u001e\u0007\u001e"+
		"\u0002\u001f\u0007\u001f\u0002 \u0007 \u0002!\u0007!\u0002\"\u0007\"\u0002"+
		"#\u0007#\u0002$\u0007$\u0002%\u0007%\u0002&\u0007&\u0002\'\u0007\'\u0002"+
		"(\u0007(\u0002)\u0007)\u0002*\u0007*\u0002+\u0007+\u0002,\u0007,\u0002"+
		"-\u0007-\u0002.\u0007.\u0002/\u0007/\u00020\u00070\u00021\u00071\u0002"+
		"2\u00072\u00023\u00073\u00024\u00074\u00025\u00075\u0001\u0000\u0005\u0000"+
		"n\b\u0000\n\u0000\f\u0000q\t\u0000\u0001\u0000\u0001\u0000\u0001\u0001"+
		"\u0005\u0001v\b\u0001\n\u0001\f\u0001y\t\u0001\u0001\u0001\u0001\u0001"+
		"\u0004\u0001}\b\u0001\u000b\u0001\f\u0001~\u0001\u0001\u0005\u0001\u0082"+
		"\b\u0001\n\u0001\f\u0001\u0085\t\u0001\u0001\u0001\u0005\u0001\u0088\b"+
		"\u0001\n\u0001\f\u0001\u008b\t\u0001\u0001\u0002\u0001\u0002\u0001\u0002"+
		"\u0001\u0002\u0003\u0002\u0091\b\u0002\u0001\u0003\u0001\u0003\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0005\u0004\u009a\b\u0004"+
		"\n\u0004\f\u0004\u009d\t\u0004\u0001\u0004\u0001\u0004\u0003\u0004\u00a1"+
		"\b\u0004\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001"+
		"\u0005\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0006\u0005\u0006\u00ad"+
		"\b\u0006\n\u0006\f\u0006\u00b0\t\u0006\u0001\u0007\u0001\u0007\u0001\u0007"+
		"\u0001\u0007\u0001\u0007\u0003\u0007\u00b7\b\u0007\u0001\u0007\u0003\u0007"+
		"\u00ba\b\u0007\u0001\b\u0003\b\u00bd\b\b\u0001\b\u0001\b\u0001\t\u0001"+
		"\t\u0001\n\u0001\n\u0001\n\u0003\n\u00c6\b\n\u0001\n\u0001\n\u0001\n\u0005"+
		"\n\u00cb\b\n\n\n\f\n\u00ce\t\n\u0001\n\u0003\n\u00d1\b\n\u0001\n\u0001"+
		"\n\u0001\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u00d8\b\u000b\u0001"+
		"\u000b\u0001\u000b\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\r\u0001"+
		"\r\u0001\r\u0003\r\u00e4\b\r\u0001\r\u0001\r\u0003\r\u00e8\b\r\u0001\r"+
		"\u0001\r\u0003\r\u00ec\b\r\u0001\r\u0001\r\u0003\r\u00f0\b\r\u0001\r\u0001"+
		"\r\u0003\r\u00f4\b\r\u0001\r\u0001\r\u0003\r\u00f8\b\r\u0001\r\u0001\r"+
		"\u0003\r\u00fc\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003"+
		"\r\u0104\b\r\u0001\r\u0003\r\u0107\b\r\u0001\r\u0003\r\u010a\b\r\u0001"+
		"\r\u0001\r\u0001\r\u0003\r\u010f\b\r\u0001\r\u0004\r\u0112\b\r\u000b\r"+
		"\f\r\u0113\u0001\r\u0001\r\u0003\r\u0118\b\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u0122\b\r\u0001\r\u0001\r\u0003"+
		"\r\u0126\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0003\r\u0130\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u0146\b\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0003\r\u014c\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0003\r\u0155\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u015b"+
		"\b\r\u0001\r\u0001\r\u0001\r\u0003\r\u0160\b\r\u0001\r\u0001\r\u0001\r"+
		"\u0001\r\u0001\r\u0003\r\u0167\b\r\u0001\r\u0001\r\u0005\r\u016b\b\r\n"+
		"\r\f\r\u016e\t\r\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001"+
		"\u000f\u0001\u000f\u0001\u000f\u0005\u000f\u0177\b\u000f\n\u000f\f\u000f"+
		"\u017a\t\u000f\u0001\u0010\u0001\u0010\u0001\u0011\u0001\u0011\u0001\u0011"+
		"\u0003\u0011\u0181\b\u0011\u0001\u0012\u0001\u0012\u0001\u0012\u0001\u0013"+
		"\u0001\u0013\u0001\u0013\u0001\u0013\u0005\u0013\u018a\b\u0013\n\u0013"+
		"\f\u0013\u018d\t\u0013\u0001\u0013\u0001\u0013\u0001\u0014\u0001\u0014"+
		"\u0001\u0014\u0001\u0014\u0005\u0014\u0195\b\u0014\n\u0014\f\u0014\u0198"+
		"\t\u0014\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0003"+
		"\u0015\u019f\b\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0005"+
		"\u0015\u01a5\b\u0015\n\u0015\f\u0015\u01a8\t\u0015\u0001\u0015\u0001\u0015"+
		"\u0003\u0015\u01ac\b\u0015\u0001\u0015\u0001\u0015\u0003\u0015\u01b0\b"+
		"\u0015\u0001\u0015\u0003\u0015\u01b3\b\u0015\u0001\u0016\u0003\u0016\u01b6"+
		"\b\u0016\u0001\u0016\u0001\u0016\u0001\u0017\u0001\u0017\u0001\u0017\u0001"+
		"\u0017\u0005\u0017\u01be\b\u0017\n\u0017\f\u0017\u01c1\t\u0017\u0001\u0018"+
		"\u0001\u0018\u0003\u0018\u01c5\b\u0018\u0001\u0018\u0001\u0018\u0001\u0018"+
		"\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019"+
		"\u0005\u0019\u01d0\b\u0019\n\u0019\f\u0019\u01d3\t\u0019\u0001\u0019\u0001"+
		"\u0019\u0001\u0019\u0003\u0019\u01d8\b\u0019\u0003\u0019\u01da\b\u0019"+
		"\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019"+
		"\u0001\u0019\u0005\u0019\u01e3\b\u0019\n\u0019\f\u0019\u01e6\t\u0019\u0001"+
		"\u0019\u0001\u0019\u0003\u0019\u01ea\b\u0019\u0003\u0019\u01ec\b\u0019"+
		"\u0001\u001a\u0001\u001a\u0001\u001a\u0001\u001a\u0005\u001a\u01f2\b\u001a"+
		"\n\u001a\f\u001a\u01f5\t\u001a\u0001\u001a\u0003\u001a\u01f8\b\u001a\u0001"+
		"\u001a\u0003\u001a\u01fb\b\u001a\u0001\u001b\u0003\u001b\u01fe\b\u001b"+
		"\u0001\u001b\u0001\u001b\u0001\u001c\u0001\u001c\u0001\u001c\u0001\u001c"+
		"\u0001\u001d\u0001\u001d\u0005\u001d\u0208\b\u001d\n\u001d\f\u001d\u020b"+
		"\t\u001d\u0001\u001e\u0001\u001e\u0003\u001e\u020f\b\u001e\u0001\u001e"+
		"\u0001\u001e\u0001\u001e\u0005\u001e\u0214\b\u001e\n\u001e\f\u001e\u0217"+
		"\t\u001e\u0001\u001e\u0001\u001e\u0003\u001e\u021b\b\u001e\u0001\u001e"+
		"\u0001\u001e\u0003\u001e\u021f\b\u001e\u0001\u001e\u0001\u001e\u0001\u001e"+
		"\u0001\u001e\u0001\u001e\u0005\u001e\u0226\b\u001e\n\u001e\f\u001e\u0229"+
		"\t\u001e\u0001\u001e\u0001\u001e\u0003\u001e\u022d\b\u001e\u0003\u001e"+
		"\u022f\b\u001e\u0001\u001f\u0001\u001f\u0001\u001f\u0003\u001f\u0234\b"+
		"\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0003"+
		"\u001f\u023b\b\u001f\u0003\u001f\u023d\b\u001f\u0001 \u0001 \u0001 \u0001"+
		" \u0001 \u0001 \u0001 \u0001 \u0003 \u0247\b \u0003 \u0249\b \u0001!\u0001"+
		"!\u0001!\u0001!\u0003!\u024f\b!\u0003!\u0251\b!\u0001\"\u0001\"\u0003"+
		"\"\u0255\b\"\u0001\"\u0003\"\u0258\b\"\u0001\"\u0001\"\u0001#\u0001#\u0001"+
		"#\u0001$\u0001$\u0003$\u0261\b$\u0001$\u0001$\u0003$\u0265\b$\u0001%\u0001"+
		"%\u0003%\u0269\b%\u0001%\u0001%\u0001%\u0001&\u0001&\u0001&\u0001&\u0001"+
		"&\u0001&\u0005&\u0274\b&\n&\f&\u0277\t&\u0001&\u0001&\u0003&\u027b\b&"+
		"\u0001&\u0001&\u0003&\u027f\b&\u0001&\u0003&\u0282\b&\u0001\'\u0003\'"+
		"\u0285\b\'\u0001\'\u0001\'\u0001(\u0001(\u0001(\u0001(\u0005(\u028d\b"+
		"(\n(\f(\u0290\t(\u0001(\u0001(\u0001)\u0001)\u0001)\u0003)\u0297\b)\u0001"+
		"*\u0001*\u0001*\u0001*\u0001*\u0005*\u029e\b*\n*\f*\u02a1\t*\u0001+\u0001"+
		"+\u0001+\u0001+\u0003+\u02a7\b+\u0001,\u0001,\u0003,\u02ab\b,\u0001,\u0001"+
		",\u0003,\u02af\b,\u0001-\u0001-\u0001.\u0001.\u0001.\u0003.\u02b6\b.\u0001"+
		"/\u0001/\u0003/\u02ba\b/\u00010\u00010\u00011\u00011\u00012\u00012\u0001"+
		"3\u00013\u00014\u00014\u00015\u00015\u00015\u0000\u0001\u001a6\u0000\u0002"+
		"\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a\u001c\u001e"+
		" \"$&(*,.02468:<>@BDFHJLNPRTVXZ\\^`bdfhj\u0000\u0007\u0001\u0000\n\u000b"+
		"\u0002\u0000\t\t\f\r\u0002\u000088@@\u0002\u0000\b\b\u000e\u0013\u0003"+
		"\u0000//<<LL\u0002\u0000--;;\u0002\u0000\u0018\u0018$$\u030e\u0000o\u0001"+
		"\u0000\u0000\u0000\u0002w\u0001\u0000\u0000\u0000\u0004\u0090\u0001\u0000"+
		"\u0000\u0000\u0006\u0092\u0001\u0000\u0000\u0000\b\u0094\u0001\u0000\u0000"+
		"\u0000\n\u00a2\u0001\u0000\u0000\u0000\f\u00a8\u0001\u0000\u0000\u0000"+
		"\u000e\u00b1\u0001\u0000\u0000\u0000\u0010\u00bc\u0001\u0000\u0000\u0000"+
		"\u0012\u00c0\u0001\u0000\u0000\u0000\u0014\u00c2\u0001\u0000\u0000\u0000"+
		"\u0016\u00d7\u0001\u0000\u0000\u0000\u0018\u00db\u0001\u0000\u0000\u0000"+
		"\u001a\u0125\u0001\u0000\u0000\u0000\u001c\u016f\u0001\u0000\u0000\u0000"+
		"\u001e\u0173\u0001\u0000\u0000\u0000 \u017b\u0001\u0000\u0000\u0000\""+
		"\u017d\u0001\u0000\u0000\u0000$\u0182\u0001\u0000\u0000\u0000&\u0185\u0001"+
		"\u0000\u0000\u0000(\u0190\u0001\u0000\u0000\u0000*\u0199\u0001\u0000\u0000"+
		"\u0000,\u01b5\u0001\u0000\u0000\u0000.\u01b9\u0001\u0000\u0000\u00000"+
		"\u01c4\u0001\u0000\u0000\u00002\u01c9\u0001\u0000\u0000\u00004\u01ed\u0001"+
		"\u0000\u0000\u00006\u01fd\u0001\u0000\u0000\u00008\u0201\u0001\u0000\u0000"+
		"\u0000:\u0205\u0001\u0000\u0000\u0000<\u020c\u0001\u0000\u0000\u0000>"+
		"\u023c\u0001\u0000\u0000\u0000@\u0248\u0001\u0000\u0000\u0000B\u0250\u0001"+
		"\u0000\u0000\u0000D\u0257\u0001\u0000\u0000\u0000F\u025b\u0001\u0000\u0000"+
		"\u0000H\u0264\u0001\u0000\u0000\u0000J\u0268\u0001\u0000\u0000\u0000L"+
		"\u026d\u0001\u0000\u0000\u0000N\u0284\u0001\u0000\u0000\u0000P\u0288\u0001"+
		"\u0000\u0000\u0000R\u0293\u0001\u0000\u0000\u0000T\u0298\u0001\u0000\u0000"+
		"\u0000V\u02a2\u0001\u0000\u0000\u0000X\u02a8\u0001\u0000\u0000\u0000Z"+
		"\u02b0\u0001\u0000\u0000\u0000\\\u02b5\u0001\u0000\u0000\u0000^\u02b9"+
		"\u0001\u0000\u0000\u0000`\u02bb\u0001\u0000\u0000\u0000b\u02bd\u0001\u0000"+
		"\u0000\u0000d\u02bf\u0001\u0000\u0000\u0000f\u02c1\u0001\u0000\u0000\u0000"+
		"h\u02c3\u0001\u0000\u0000\u0000j\u02c5\u0001\u0000\u0000\u0000ln\u0003"+
		"\u0002\u0001\u0000ml\u0001\u0000\u0000\u0000nq\u0001\u0000\u0000\u0000"+
		"om\u0001\u0000\u0000\u0000op\u0001\u0000\u0000\u0000pr\u0001\u0000\u0000"+
		"\u0000qo\u0001\u0000\u0000\u0000rs\u0005\u0000\u0000\u0001s\u0001\u0001"+
		"\u0000\u0000\u0000tv\u0005\u0001\u0000\u0000ut\u0001\u0000\u0000\u0000"+
		"vy\u0001\u0000\u0000\u0000wu\u0001\u0000\u0000\u0000wx\u0001\u0000\u0000"+
		"\u0000xz\u0001\u0000\u0000\u0000yw\u0001\u0000\u0000\u0000z\u0083\u0003"+
		"\u0004\u0002\u0000{}\u0005\u0001\u0000\u0000|{\u0001\u0000\u0000\u0000"+
		"}~\u0001\u0000\u0000\u0000~|\u0001\u0000\u0000\u0000~\u007f\u0001\u0000"+
		"\u0000\u0000\u007f\u0080\u0001\u0000\u0000\u0000\u0080\u0082\u0003\u0004"+
		"\u0002\u0000\u0081|\u0001\u0000\u0000\u0000\u0082\u0085\u0001\u0000\u0000"+
		"\u0000\u0083\u0081\u0001\u0000\u0000\u0000\u0083\u0084\u0001\u0000\u0000"+
		"\u0000\u0084\u0089\u0001\u0000\u0000\u0000\u0085\u0083\u0001\u0000\u0000"+
		"\u0000\u0086\u0088\u0005\u0001\u0000\u0000\u0087\u0086\u0001\u0000\u0000"+
		"\u0000\u0088\u008b\u0001\u0000\u0000\u0000\u0089\u0087\u0001\u0000\u0000"+
		"\u0000\u0089\u008a\u0001\u0000\u0000\u0000\u008a\u0003\u0001\u0000\u0000"+
		"\u0000\u008b\u0089\u0001\u0000\u0000\u0000\u008c\u0091\u0003\u0010\b\u0000"+
		"\u008d\u0091\u0003,\u0016\u0000\u008e\u0091\u00036\u001b\u0000\u008f\u0091"+
		"\u0003N\'\u0000\u0090\u008c\u0001\u0000\u0000\u0000\u0090\u008d\u0001"+
		"\u0000\u0000\u0000\u0090\u008e\u0001\u0000\u0000\u0000\u0090\u008f\u0001"+
		"\u0000\u0000\u0000\u0091\u0005\u0001\u0000\u0000\u0000\u0092\u0093\u0003"+
		"d2\u0000\u0093\u0007\u0001\u0000\u0000\u0000\u0094\u00a0\u0003`0\u0000"+
		"\u0095\u0096\u0005\u0003\u0000\u0000\u0096\u009b\u0003d2\u0000\u0097\u0098"+
		"\u0005\u0007\u0000\u0000\u0098\u009a\u0003d2\u0000\u0099\u0097\u0001\u0000"+
		"\u0000\u0000\u009a\u009d\u0001\u0000\u0000\u0000\u009b\u0099\u0001\u0000"+
		"\u0000\u0000\u009b\u009c\u0001\u0000\u0000\u0000\u009c\u009e\u0001\u0000"+
		"\u0000\u0000\u009d\u009b\u0001\u0000\u0000\u0000\u009e\u009f\u0005\u0004"+
		"\u0000\u0000\u009f\u00a1\u0001\u0000\u0000\u0000\u00a0\u0095\u0001\u0000"+
		"\u0000\u0000\u00a0\u00a1\u0001\u0000\u0000\u0000\u00a1\t\u0001\u0000\u0000"+
		"\u0000\u00a2\u00a3\u0003\b\u0004\u0000\u00a3\u00a4\u0005\u0019\u0000\u0000"+
		"\u00a4\u00a5\u0005\u0003\u0000\u0000\u00a5\u00a6\u00034\u001a\u0000\u00a6"+
		"\u00a7\u0005\u0004\u0000\u0000\u00a7\u000b\u0001\u0000\u0000\u0000\u00a8"+
		"\u00a9\u0005V\u0000\u0000\u00a9\u00ae\u0003\n\u0005\u0000\u00aa\u00ab"+
		"\u0005\u0007\u0000\u0000\u00ab\u00ad\u0003\n\u0005\u0000\u00ac\u00aa\u0001"+
		"\u0000\u0000\u0000\u00ad\u00b0\u0001\u0000\u0000\u0000\u00ae\u00ac\u0001"+
		"\u0000\u0000\u0000\u00ae\u00af\u0001\u0000\u0000\u0000\u00af\r\u0001\u0000"+
		"\u0000\u0000\u00b0\u00ae\u0001\u0000\u0000\u0000\u00b1\u00b2\u0005#\u0000"+
		"\u0000\u00b2\u00b3\u0005.\u0000\u0000\u00b3\u00b6\u0003R)\u0000\u00b4"+
		"\u00b5\u0005U\u0000\u0000\u00b5\u00b7\u0003\u001a\r\u0000\u00b6\u00b4"+
		"\u0001\u0000\u0000\u0000\u00b6\u00b7\u0001\u0000\u0000\u0000\u00b7\u00b9"+
		"\u0001\u0000\u0000\u0000\u00b8\u00ba\u0003.\u0017\u0000\u00b9\u00b8\u0001"+
		"\u0000\u0000\u0000\u00b9\u00ba\u0001\u0000\u0000\u0000\u00ba\u000f\u0001"+
		"\u0000\u0000\u0000\u00bb\u00bd\u0003\f\u0006\u0000\u00bc\u00bb\u0001\u0000"+
		"\u0000\u0000\u00bc\u00bd\u0001\u0000\u0000\u0000\u00bd\u00be\u0001\u0000"+
		"\u0000\u0000\u00be\u00bf\u0003\u000e\u0007\u0000\u00bf\u0011\u0001\u0000"+
		"\u0000\u0000\u00c0\u00c1\u0005]\u0000\u0000\u00c1\u0013\u0001\u0000\u0000"+
		"\u0000\u00c2\u00c3\u0003^/\u0000\u00c3\u00d0\u0005\u0003\u0000\u0000\u00c4"+
		"\u00c6\u0005%\u0000\u0000\u00c5\u00c4\u0001\u0000\u0000\u0000\u00c5\u00c6"+
		"\u0001\u0000\u0000\u0000\u00c6\u00c7\u0001\u0000\u0000\u0000\u00c7\u00cc"+
		"\u0003\u001a\r\u0000\u00c8\u00c9\u0005\u0007\u0000\u0000\u00c9\u00cb\u0003"+
		"\u001a\r\u0000\u00ca\u00c8\u0001\u0000\u0000\u0000\u00cb\u00ce\u0001\u0000"+
		"\u0000\u0000\u00cc\u00ca\u0001\u0000\u0000\u0000\u00cc\u00cd\u0001\u0000"+
		"\u0000\u0000\u00cd\u00d1\u0001\u0000\u0000\u0000\u00ce\u00cc\u0001\u0000"+
		"\u0000\u0000\u00cf\u00d1\u0005\t\u0000\u0000\u00d0\u00c5\u0001\u0000\u0000"+
		"\u0000\u00d0\u00cf\u0001\u0000\u0000\u0000\u00d0\u00d1\u0001\u0000\u0000"+
		"\u0000\u00d1\u00d2\u0001\u0000\u0000\u0000\u00d2\u00d3\u0005\u0004\u0000"+
		"\u0000\u00d3\u0015\u0001\u0000\u0000\u0000\u00d4\u00d5\u0003`0\u0000\u00d5"+
		"\u00d6\u0005\u0002\u0000\u0000\u00d6\u00d8\u0001\u0000\u0000\u0000\u00d7"+
		"\u00d4\u0001\u0000\u0000\u0000\u00d7\u00d8\u0001\u0000\u0000\u0000\u00d8"+
		"\u00d9\u0001\u0000\u0000\u0000\u00d9\u00da\u0003d2\u0000\u00da\u0017\u0001"+
		"\u0000\u0000\u0000\u00db\u00dc\u0005T\u0000\u0000\u00dc\u00dd\u0003\u001a"+
		"\r\u0000\u00dd\u00de\u0005O\u0000\u0000\u00de\u00df\u0003\u001a\r\u0000"+
		"\u00df\u0019\u0001\u0000\u0000\u0000\u00e0\u00e1\u0006\r\uffff\uffff\u0000"+
		"\u00e1\u00e3\u0005Z\u0000\u0000\u00e2\u00e4\u0003$\u0012\u0000\u00e3\u00e2"+
		"\u0001\u0000\u0000\u0000\u00e3\u00e4\u0001\u0000\u0000\u0000\u00e4\u0126"+
		"\u0001\u0000\u0000\u0000\u00e5\u00e7\u0005W\u0000\u0000\u00e6\u00e8\u0003"+
		"$\u0012\u0000\u00e7\u00e6\u0001\u0000\u0000\u0000\u00e7\u00e8\u0001\u0000"+
		"\u0000\u0000\u00e8\u0126\u0001\u0000\u0000\u0000\u00e9\u00eb\u0005X\u0000"+
		"\u0000\u00ea\u00ec\u0003$\u0012\u0000\u00eb\u00ea\u0001\u0000\u0000\u0000"+
		"\u00eb\u00ec\u0001\u0000\u0000\u0000\u00ec\u0126\u0001\u0000\u0000\u0000"+
		"\u00ed\u00ef\u0005[\u0000\u0000\u00ee\u00f0\u0003$\u0012\u0000\u00ef\u00ee"+
		"\u0001\u0000\u0000\u0000\u00ef\u00f0\u0001\u0000\u0000\u0000\u00f0\u0126"+
		"\u0001\u0000\u0000\u0000\u00f1\u00f3\u0005Y\u0000\u0000\u00f2\u00f4\u0003"+
		"$\u0012\u0000\u00f3\u00f2\u0001\u0000\u0000\u0000\u00f3\u00f4\u0001\u0000"+
		"\u0000\u0000\u00f4\u0126\u0001\u0000\u0000\u0000\u00f5\u00f7\u0003\u0012"+
		"\t\u0000\u00f6\u00f8\u0003$\u0012\u0000\u00f7\u00f6\u0001\u0000\u0000"+
		"\u0000\u00f7\u00f8\u0001\u0000\u0000\u0000\u00f8\u0126\u0001\u0000\u0000"+
		"\u0000\u00f9\u00fb\u0003\u0016\u000b\u0000\u00fa\u00fc\u0003$\u0012\u0000"+
		"\u00fb\u00fa\u0001\u0000\u0000\u0000\u00fb\u00fc\u0001\u0000\u0000\u0000"+
		"\u00fc\u0126\u0001\u0000\u0000\u0000\u00fd\u00fe\u0007\u0000\u0000\u0000"+
		"\u00fe\u0126\u0003\u001a\r\u0013\u00ff\u0100\u0005\u0003\u0000\u0000\u0100"+
		"\u0101\u0003\u001a\r\u0000\u0101\u0103\u0005\u0004\u0000\u0000\u0102\u0104"+
		"\u0003$\u0012\u0000\u0103\u0102\u0001\u0000\u0000\u0000\u0103\u0104\u0001"+
		"\u0000\u0000\u0000\u0104\u0126\u0001\u0000\u0000\u0000\u0105\u0107\u0005"+
		"A\u0000\u0000\u0106\u0105\u0001\u0000\u0000\u0000\u0106\u0107\u0001\u0000"+
		"\u0000\u0000\u0107\u0108\u0001\u0000\u0000\u0000\u0108\u010a\u0005+\u0000"+
		"\u0000\u0109\u0106\u0001\u0000\u0000\u0000\u0109\u010a\u0001\u0000\u0000"+
		"\u0000\u010a\u010b\u0001\u0000\u0000\u0000\u010b\u0126\u0003\u001c\u000e"+
		"\u0000\u010c\u010e\u0005\u001c\u0000\u0000\u010d\u010f\u0003\u001a\r\u0000"+
		"\u010e\u010d\u0001\u0000\u0000\u0000\u010e\u010f\u0001\u0000\u0000\u0000"+
		"\u010f\u0111\u0001\u0000\u0000\u0000\u0110\u0112\u0003\u0018\f\u0000\u0111"+
		"\u0110\u0001\u0000\u0000\u0000\u0112\u0113\u0001\u0000\u0000\u0000\u0113"+
		"\u0111\u0001\u0000\u0000\u0000\u0113\u0114\u0001\u0000\u0000\u0000\u0114"+
		"\u0117\u0001\u0000\u0000\u0000\u0115\u0116\u0005\'\u0000\u0000\u0116\u0118"+
		"\u0003\u001a\r\u0000\u0117\u0115\u0001\u0000\u0000\u0000\u0117\u0118\u0001"+
		"\u0000\u0000\u0000\u0118\u0119\u0001\u0000\u0000\u0000\u0119\u011a\u0005"+
		"(\u0000\u0000\u011a\u0126\u0001\u0000\u0000\u0000\u011b\u011c\u0005\u0003"+
		"\u0000\u0000\u011c\u011d\u0003\u001e\u000f\u0000\u011d\u011e\u0005\u0004"+
		"\u0000\u0000\u011e\u0126\u0001\u0000\u0000\u0000\u011f\u0121\u0003\u0014"+
		"\n\u0000\u0120\u0122\u0003$\u0012\u0000\u0121\u0120\u0001\u0000\u0000"+
		"\u0000\u0121\u0122\u0001\u0000\u0000\u0000\u0122\u0126\u0001\u0000\u0000"+
		"\u0000\u0123\u0124\u0005A\u0000\u0000\u0124\u0126\u0003\u001a\r\u0003"+
		"\u0125\u00e0\u0001\u0000\u0000\u0000\u0125\u00e5\u0001\u0000\u0000\u0000"+
		"\u0125\u00e9\u0001\u0000\u0000\u0000\u0125\u00ed\u0001\u0000\u0000\u0000"+
		"\u0125\u00f1\u0001\u0000\u0000\u0000\u0125\u00f5\u0001\u0000\u0000\u0000"+
		"\u0125\u00f9\u0001\u0000\u0000\u0000\u0125\u00fd\u0001\u0000\u0000\u0000"+
		"\u0125\u00ff\u0001\u0000\u0000\u0000\u0125\u0109\u0001\u0000\u0000\u0000"+
		"\u0125\u010c\u0001\u0000\u0000\u0000\u0125\u011b\u0001\u0000\u0000\u0000"+
		"\u0125\u011f\u0001\u0000\u0000\u0000\u0125\u0123\u0001\u0000\u0000\u0000"+
		"\u0126\u016c\u0001\u0000\u0000\u0000\u0127\u0128\n\f\u0000\u0000\u0128"+
		"\u0129\u0007\u0001\u0000\u0000\u0129\u016b\u0003\u001a\r\r\u012a\u012b"+
		"\n\u000b\u0000\u0000\u012b\u012c\u0007\u0000\u0000\u0000\u012c\u016b\u0003"+
		"\u001a\r\f\u012d\u012f\n\b\u0000\u0000\u012e\u0130\u0005A\u0000\u0000"+
		"\u012f\u012e\u0001\u0000\u0000\u0000\u012f\u0130\u0001\u0000\u0000\u0000"+
		"\u0130\u0131\u0001\u0000\u0000\u0000\u0131\u0132\u0005\u001a\u0000\u0000"+
		"\u0132\u0133\u0003\u001a\r\u0000\u0133\u0134\u0005\u0017\u0000\u0000\u0134"+
		"\u0135\u0003\u001a\r\t\u0135\u016b\u0001\u0000\u0000\u0000\u0136\u0137"+
		"\n\u0006\u0000\u0000\u0137\u0138\u0003 \u0010\u0000\u0138\u0139\u0003"+
		"\u001a\r\u0007\u0139\u016b\u0001\u0000\u0000\u0000\u013a\u013b\n\u0002"+
		"\u0000\u0000\u013b\u013c\u0005\u0017\u0000\u0000\u013c\u016b\u0003\u001a"+
		"\r\u0003\u013d\u013e\n\u0001\u0000\u0000\u013e\u013f\u0005G\u0000\u0000"+
		"\u013f\u016b\u0003\u001a\r\u0002\u0140\u0141\n\u0012\u0000\u0000\u0141"+
		"\u0142\u0005\u001d\u0000\u0000\u0142\u016b\u0003h4\u0000\u0143\u0145\n"+
		"\n\u0000\u0000\u0144\u0146\u0005A\u0000\u0000\u0145\u0144\u0001\u0000"+
		"\u0000\u0000\u0145\u0146\u0001\u0000\u0000\u0000\u0146\u0147\u0001\u0000"+
		"\u0000\u0000\u0147\u0148\u00057\u0000\u0000\u0148\u016b\u0003\u001c\u000e"+
		"\u0000\u0149\u014b\n\t\u0000\u0000\u014a\u014c\u0005A\u0000\u0000\u014b"+
		"\u014a\u0001\u0000\u0000\u0000\u014b\u014c\u0001\u0000\u0000\u0000\u014c"+
		"\u014d\u0001\u0000\u0000\u0000\u014d\u014e\u00057\u0000\u0000\u014e\u014f"+
		"\u0005\u0003\u0000\u0000\u014f\u0150\u0003\u001e\u000f\u0000\u0150\u0151"+
		"\u0005\u0004\u0000\u0000\u0151\u016b\u0001\u0000\u0000\u0000\u0152\u0154"+
		"\n\u0007\u0000\u0000\u0153\u0155\u0005A\u0000\u0000\u0154\u0153\u0001"+
		"\u0000\u0000\u0000\u0154\u0155\u0001\u0000\u0000\u0000\u0155\u0156\u0001"+
		"\u0000\u0000\u0000\u0156\u0157\u0005=\u0000\u0000\u0157\u015a\u0003\u001a"+
		"\r\u0000\u0158\u0159\u0005)\u0000\u0000\u0159\u015b\u0003\u001a\r\u0000"+
		"\u015a\u0158\u0001\u0000\u0000\u0000\u015a\u015b\u0001\u0000\u0000\u0000"+
		"\u015b\u016b\u0001\u0000\u0000\u0000\u015c\u015d\n\u0005\u0000\u0000\u015d"+
		"\u015f\u00059\u0000\u0000\u015e\u0160\u0005A\u0000\u0000\u015f\u015e\u0001"+
		"\u0000\u0000\u0000\u015f\u0160\u0001\u0000\u0000\u0000\u0160\u0166\u0001"+
		"\u0000\u0000\u0000\u0161\u0162\u0005%\u0000\u0000\u0162\u0163\u0005.\u0000"+
		"\u0000\u0163\u0167\u0003\u001a\r\u0000\u0164\u0167\u0005W\u0000\u0000"+
		"\u0165\u0167\u0005[\u0000\u0000\u0166\u0161\u0001\u0000\u0000\u0000\u0166"+
		"\u0164\u0001\u0000\u0000\u0000\u0166\u0165\u0001\u0000\u0000\u0000\u0167"+
		"\u016b\u0001\u0000\u0000\u0000\u0168\u0169\n\u0004\u0000\u0000\u0169\u016b"+
		"\u0007\u0002\u0000\u0000\u016a\u0127\u0001\u0000\u0000\u0000\u016a\u012a"+
		"\u0001\u0000\u0000\u0000\u016a\u012d\u0001\u0000\u0000\u0000\u016a\u0136"+
		"\u0001\u0000\u0000\u0000\u016a\u013a\u0001\u0000\u0000\u0000\u016a\u013d"+
		"\u0001\u0000\u0000\u0000\u016a\u0140\u0001\u0000\u0000\u0000\u016a\u0143"+
		"\u0001\u0000\u0000\u0000\u016a\u0149\u0001\u0000\u0000\u0000\u016a\u0152"+
		"\u0001\u0000\u0000\u0000\u016a\u015c\u0001\u0000\u0000\u0000\u016a\u0168"+
		"\u0001\u0000\u0000\u0000\u016b\u016e\u0001\u0000\u0000\u0000\u016c\u016a"+
		"\u0001\u0000\u0000\u0000\u016c\u016d\u0001\u0000\u0000\u0000\u016d\u001b"+
		"\u0001\u0000\u0000\u0000\u016e\u016c\u0001\u0000\u0000\u0000\u016f\u0170"+
		"\u0005\u0003\u0000\u0000\u0170\u0171\u00034\u001a\u0000\u0171\u0172\u0005"+
		"\u0004\u0000\u0000\u0172\u001d\u0001\u0000\u0000\u0000\u0173\u0178\u0003"+
		"\u001a\r\u0000\u0174\u0175\u0005\u0007\u0000\u0000\u0175\u0177\u0003\u001a"+
		"\r\u0000\u0176\u0174\u0001\u0000\u0000\u0000\u0177\u017a\u0001\u0000\u0000"+
		"\u0000\u0178\u0176\u0001\u0000\u0000\u0000\u0178\u0179\u0001\u0000\u0000"+
		"\u0000\u0179\u001f\u0001\u0000\u0000\u0000\u017a\u0178\u0001\u0000\u0000"+
		"\u0000\u017b\u017c\u0007\u0003\u0000\u0000\u017c!\u0001\u0000\u0000\u0000"+
		"\u017d\u0180\u0005\\\u0000\u0000\u017e\u017f\u0005\u0005\u0000\u0000\u017f"+
		"\u0181\u0005\u0006\u0000\u0000\u0180\u017e\u0001\u0000\u0000\u0000\u0180"+
		"\u0181\u0001\u0000\u0000\u0000\u0181#\u0001\u0000\u0000\u0000\u0182\u0183"+
		"\u0005\u0014\u0000\u0000\u0183\u0184\u0003\"\u0011\u0000\u0184%\u0001"+
		"\u0000\u0000\u0000\u0185\u0186\u0005\u0003\u0000\u0000\u0186\u018b\u0003"+
		"\u001a\r\u0000\u0187\u0188\u0005\u0007\u0000\u0000\u0188\u018a\u0003\u001a"+
		"\r\u0000\u0189\u0187\u0001\u0000\u0000\u0000\u018a\u018d\u0001\u0000\u0000"+
		"\u0000\u018b\u0189\u0001\u0000\u0000\u0000\u018b\u018c\u0001\u0000\u0000"+
		"\u0000\u018c\u018e\u0001\u0000\u0000\u0000\u018d\u018b\u0001\u0000\u0000"+
		"\u0000\u018e\u018f\u0005\u0004\u0000\u0000\u018f\'\u0001\u0000\u0000\u0000"+
		"\u0190\u0191\u0005S\u0000\u0000\u0191\u0196\u0003&\u0013\u0000\u0192\u0193"+
		"\u0005\u0007\u0000\u0000\u0193\u0195\u0003&\u0013\u0000\u0194\u0192\u0001"+
		"\u0000\u0000\u0000\u0195\u0198\u0001\u0000\u0000\u0000\u0196\u0194\u0001"+
		"\u0000\u0000\u0000\u0196\u0197\u0001\u0000\u0000\u0000\u0197)\u0001\u0000"+
		"\u0000\u0000\u0198\u0196\u0001\u0000\u0000\u0000\u0199\u019a\u00054\u0000"+
		"\u0000\u019a\u019b\u00056\u0000\u0000\u019b\u019e\u0003`0\u0000\u019c"+
		"\u019d\u0005\u0019\u0000\u0000\u019d\u019f\u0003b1\u0000\u019e\u019c\u0001"+
		"\u0000\u0000\u0000\u019e\u019f\u0001\u0000\u0000\u0000\u019f\u01ab\u0001"+
		"\u0000\u0000\u0000\u01a0\u01a1\u0005\u0003\u0000\u0000\u01a1\u01a6\u0003"+
		"d2\u0000\u01a2\u01a3\u0005\u0007\u0000\u0000\u01a3\u01a5\u0003d2\u0000"+
		"\u01a4\u01a2\u0001\u0000\u0000\u0000\u01a5\u01a8\u0001\u0000\u0000\u0000"+
		"\u01a6\u01a4\u0001\u0000\u0000\u0000\u01a6\u01a7\u0001\u0000\u0000\u0000"+
		"\u01a7\u01a9\u0001\u0000\u0000\u0000\u01a8\u01a6\u0001\u0000\u0000\u0000"+
		"\u01a9\u01aa\u0005\u0004\u0000\u0000\u01aa\u01ac\u0001\u0000\u0000\u0000"+
		"\u01ab\u01a0\u0001\u0000\u0000\u0000\u01ab\u01ac\u0001\u0000\u0000\u0000"+
		"\u01ac\u01ad\u0001\u0000\u0000\u0000\u01ad\u01af\u0003(\u0014\u0000\u01ae"+
		"\u01b0\u00032\u0019\u0000\u01af\u01ae\u0001\u0000\u0000\u0000\u01af\u01b0"+
		"\u0001\u0000\u0000\u0000\u01b0\u01b2\u0001\u0000\u0000\u0000\u01b1\u01b3"+
		"\u0003.\u0017\u0000\u01b2\u01b1\u0001\u0000\u0000\u0000\u01b2\u01b3\u0001"+
		"\u0000\u0000\u0000\u01b3+\u0001\u0000\u0000\u0000\u01b4\u01b6\u0003\f"+
		"\u0006\u0000\u01b5\u01b4\u0001\u0000\u0000\u0000\u01b5\u01b6\u0001\u0000"+
		"\u0000\u0000\u01b6\u01b7\u0001\u0000\u0000\u0000\u01b7\u01b8\u0003*\u0015"+
		"\u0000\u01b8-\u0001\u0000\u0000\u0000\u01b9\u01ba\u0005K\u0000\u0000\u01ba"+
		"\u01bf\u0003B!\u0000\u01bb\u01bc\u0005\u0007\u0000\u0000\u01bc\u01be\u0003"+
		"B!\u0000\u01bd\u01bb\u0001\u0000\u0000\u0000\u01be\u01c1\u0001\u0000\u0000"+
		"\u0000\u01bf\u01bd\u0001\u0000\u0000\u0000\u01bf\u01c0\u0001\u0000\u0000"+
		"\u0000\u01c0/\u0001\u0000\u0000\u0000\u01c1\u01bf\u0001\u0000\u0000\u0000"+
		"\u01c2\u01c5\u0003d2\u0000\u01c3\u01c5\u0003P(\u0000\u01c4\u01c2\u0001"+
		"\u0000\u0000\u0000\u01c4\u01c3\u0001\u0000\u0000\u0000\u01c5\u01c6\u0001"+
		"\u0000\u0000\u0000\u01c6\u01c7\u0005\b\u0000\u0000\u01c7\u01c8\u0003\u001a"+
		"\r\u0000\u01c81\u0001\u0000\u0000\u0000\u01c9\u01ca\u0005E\u0000\u0000"+
		"\u01ca\u01d9\u0005\u001f\u0000\u0000\u01cb\u01cc\u0005\u0003\u0000\u0000"+
		"\u01cc\u01d1\u0003\u0006\u0003\u0000\u01cd\u01ce\u0005\u0007\u0000\u0000"+
		"\u01ce\u01d0\u0003\u0006\u0003\u0000\u01cf\u01cd\u0001\u0000\u0000\u0000"+
		"\u01d0\u01d3\u0001\u0000\u0000\u0000\u01d1\u01cf\u0001\u0000\u0000\u0000"+
		"\u01d1\u01d2\u0001\u0000\u0000\u0000\u01d2\u01d4\u0001\u0000\u0000\u0000"+
		"\u01d3\u01d1\u0001\u0000\u0000\u0000\u01d4\u01d7\u0005\u0004\u0000\u0000"+
		"\u01d5\u01d6\u0005U\u0000\u0000\u01d6\u01d8\u0003\u001a\r\u0000\u01d7"+
		"\u01d5\u0001\u0000\u0000\u0000\u01d7\u01d8\u0001\u0000\u0000\u0000\u01d8"+
		"\u01da\u0001\u0000\u0000\u0000\u01d9\u01cb\u0001\u0000\u0000\u0000\u01d9"+
		"\u01da\u0001\u0000\u0000\u0000\u01da\u01db\u0001\u0000\u0000\u0000\u01db"+
		"\u01eb\u0005&\u0000\u0000\u01dc\u01ec\u0005?\u0000\u0000\u01dd\u01de\u0005"+
		"Q\u0000\u0000\u01de\u01df\u0005N\u0000\u0000\u01df\u01e4\u00030\u0018"+
		"\u0000\u01e0\u01e1\u0005\u0007\u0000\u0000\u01e1\u01e3\u00030\u0018\u0000"+
		"\u01e2\u01e0\u0001\u0000\u0000\u0000\u01e3\u01e6\u0001\u0000\u0000\u0000"+
		"\u01e4\u01e2\u0001\u0000\u0000\u0000\u01e4\u01e5\u0001\u0000\u0000\u0000"+
		"\u01e5\u01e9\u0001\u0000\u0000\u0000\u01e6\u01e4\u0001\u0000\u0000\u0000"+
		"\u01e7\u01e8\u0005U\u0000\u0000\u01e8\u01ea\u0003\u001a\r\u0000\u01e9"+
		"\u01e7\u0001\u0000\u0000\u0000\u01e9\u01ea\u0001\u0000\u0000\u0000\u01ea"+
		"\u01ec\u0001\u0000\u0000\u0000\u01eb\u01dc\u0001\u0000\u0000\u0000\u01eb"+
		"\u01dd\u0001\u0000\u0000\u0000\u01ec3\u0001\u0000\u0000\u0000\u01ed\u01f3"+
		"\u0003<\u001e\u0000\u01ee\u01ef\u0003H$\u0000\u01ef\u01f0\u0003<\u001e"+
		"\u0000\u01f0\u01f2\u0001\u0000\u0000\u0000\u01f1\u01ee\u0001\u0000\u0000"+
		"\u0000\u01f2\u01f5\u0001\u0000\u0000\u0000\u01f3\u01f1\u0001\u0000\u0000"+
		"\u0000\u01f3\u01f4\u0001\u0000\u0000\u0000\u01f4\u01f7\u0001\u0000\u0000"+
		"\u0000\u01f5\u01f3\u0001\u0000\u0000\u0000\u01f6\u01f8\u0003T*\u0000\u01f7"+
		"\u01f6\u0001\u0000\u0000\u0000\u01f7\u01f8\u0001\u0000\u0000\u0000\u01f8"+
		"\u01fa\u0001\u0000\u0000\u0000\u01f9\u01fb\u0003V+\u0000\u01fa\u01f9\u0001"+
		"\u0000\u0000\u0000\u01fa\u01fb\u0001\u0000\u0000\u0000\u01fb5\u0001\u0000"+
		"\u0000\u0000\u01fc\u01fe\u0003\f\u0006\u0000\u01fd\u01fc\u0001\u0000\u0000"+
		"\u0000\u01fd\u01fe\u0001\u0000\u0000\u0000\u01fe\u01ff\u0001\u0000\u0000"+
		"\u0000\u01ff\u0200\u00034\u001a\u0000\u02007\u0001\u0000\u0000\u0000\u0201"+
		"\u0202\u0003D\"\u0000\u0202\u0203\u0003>\u001f\u0000\u0203\u0204\u0003"+
		"F#\u0000\u02049\u0001\u0000\u0000\u0000\u0205\u0209\u0003>\u001f\u0000"+
		"\u0206\u0208\u00038\u001c\u0000\u0207\u0206\u0001\u0000\u0000\u0000\u0208"+
		"\u020b\u0001\u0000\u0000\u0000\u0209\u0207\u0001\u0000\u0000\u0000\u0209"+
		"\u020a\u0001\u0000\u0000\u0000\u020a;\u0001\u0000\u0000\u0000\u020b\u0209"+
		"\u0001\u0000\u0000\u0000\u020c\u020e\u0005M\u0000\u0000\u020d\u020f\u0005"+
		"%\u0000\u0000\u020e\u020d\u0001\u0000\u0000\u0000\u020e\u020f\u0001\u0000"+
		"\u0000\u0000\u020f\u0210\u0001\u0000\u0000\u0000\u0210\u0215\u0003@ \u0000"+
		"\u0211\u0212\u0005\u0007\u0000\u0000\u0212\u0214\u0003@ \u0000\u0213\u0211"+
		"\u0001\u0000\u0000\u0000\u0214\u0217\u0001\u0000\u0000\u0000\u0215\u0213"+
		"\u0001\u0000\u0000\u0000\u0215\u0216\u0001\u0000\u0000\u0000\u0216\u021a"+
		"\u0001\u0000\u0000\u0000\u0217\u0215\u0001\u0000\u0000\u0000\u0218\u0219"+
		"\u0005.\u0000\u0000\u0219\u021b\u0003:\u001d\u0000\u021a\u0218\u0001\u0000"+
		"\u0000\u0000\u021a\u021b\u0001\u0000\u0000\u0000\u021b\u021e\u0001\u0000"+
		"\u0000\u0000\u021c\u021d\u0005U\u0000\u0000\u021d\u021f\u0003\u001a\r"+
		"\u0000\u021e\u021c\u0001\u0000\u0000\u0000\u021e\u021f\u0001\u0000\u0000"+
		"\u0000\u021f\u022e\u0001\u0000\u0000\u0000\u0220\u0221\u00051\u0000\u0000"+
		"\u0221\u0222\u0005\u001b\u0000\u0000\u0222\u0227\u0003\u001a\r\u0000\u0223"+
		"\u0224\u0005\u0007\u0000\u0000\u0224\u0226\u0003\u001a\r\u0000\u0225\u0223"+
		"\u0001\u0000\u0000\u0000\u0226\u0229\u0001\u0000\u0000\u0000\u0227\u0225"+
		"\u0001\u0000\u0000\u0000\u0227\u0228\u0001\u0000\u0000\u0000\u0228\u022c"+
		"\u0001\u0000\u0000\u0000\u0229\u0227\u0001\u0000\u0000\u0000\u022a\u022b"+
		"\u00052\u0000\u0000\u022b\u022d\u0003\u001a\r\u0000\u022c\u022a\u0001"+
		"\u0000\u0000\u0000\u022c\u022d\u0001\u0000\u0000\u0000\u022d\u022f\u0001"+
		"\u0000\u0000\u0000\u022e\u0220\u0001\u0000\u0000\u0000\u022e\u022f\u0001"+
		"\u0000\u0000\u0000\u022f=\u0001\u0000\u0000\u0000\u0230\u0233\u0003`0"+
		"\u0000\u0231\u0232\u0005\u0019\u0000\u0000\u0232\u0234\u0003b1\u0000\u0233"+
		"\u0231\u0001\u0000\u0000\u0000\u0233\u0234\u0001\u0000\u0000\u0000\u0234"+
		"\u023d\u0001\u0000\u0000\u0000\u0235\u0236\u0005\u0003\u0000\u0000\u0236"+
		"\u0237\u00034\u001a\u0000\u0237\u023a\u0005\u0004\u0000\u0000\u0238\u0239"+
		"\u0005\u0019\u0000\u0000\u0239\u023b\u0003b1\u0000\u023a\u0238\u0001\u0000"+
		"\u0000\u0000\u023a\u023b\u0001\u0000\u0000\u0000\u023b\u023d\u0001\u0000"+
		"\u0000\u0000\u023c\u0230\u0001\u0000\u0000\u0000\u023c\u0235\u0001\u0000"+
		"\u0000\u0000\u023d?\u0001\u0000\u0000\u0000\u023e\u0249\u0005\t\u0000"+
		"\u0000\u023f\u0240\u0003`0\u0000\u0240\u0241\u0005\u0002\u0000\u0000\u0241"+
		"\u0242\u0005\t\u0000\u0000\u0242\u0249\u0001\u0000\u0000\u0000\u0243\u0246"+
		"\u0003\u001a\r\u0000\u0244\u0245\u0005\u0019\u0000\u0000\u0245\u0247\u0003"+
		"f3\u0000\u0246\u0244\u0001\u0000\u0000\u0000\u0246\u0247\u0001\u0000\u0000"+
		"\u0000\u0247\u0249\u0001\u0000\u0000\u0000\u0248\u023e\u0001\u0000\u0000"+
		"\u0000\u0248\u023f\u0001\u0000\u0000\u0000\u0248\u0243\u0001\u0000\u0000"+
		"\u0000\u0249A\u0001\u0000\u0000\u0000\u024a\u0251\u0005\t\u0000\u0000"+
		"\u024b\u024e\u0003\u001a\r\u0000\u024c\u024d\u0005\u0019\u0000\u0000\u024d"+
		"\u024f\u0003f3\u0000\u024e\u024c\u0001\u0000\u0000\u0000\u024e\u024f\u0001"+
		"\u0000\u0000\u0000\u024f\u0251\u0001\u0000\u0000\u0000\u0250\u024a\u0001"+
		"\u0000\u0000\u0000\u0250\u024b\u0001\u0000\u0000\u0000\u0251C\u0001\u0000"+
		"\u0000\u0000\u0252\u0254\u0007\u0004\u0000\u0000\u0253\u0255\u0005H\u0000"+
		"\u0000\u0254\u0253\u0001\u0000\u0000\u0000\u0254\u0255\u0001\u0000\u0000"+
		"\u0000\u0255\u0258\u0001\u0000\u0000\u0000\u0256\u0258\u00053\u0000\u0000"+
		"\u0257\u0252\u0001\u0000\u0000\u0000\u0257\u0256\u0001\u0000\u0000\u0000"+
		"\u0257\u0258\u0001\u0000\u0000\u0000\u0258\u0259\u0001\u0000\u0000\u0000"+
		"\u0259\u025a\u0005:\u0000\u0000\u025aE\u0001\u0000\u0000\u0000\u025b\u025c"+
		"\u0005E\u0000\u0000\u025c\u025d\u0003\u001a\r\u0000\u025dG\u0001\u0000"+
		"\u0000\u0000\u025e\u0260\u0005P\u0000\u0000\u025f\u0261\u0005\u0016\u0000"+
		"\u0000\u0260\u025f\u0001\u0000\u0000\u0000\u0260\u0261\u0001\u0000\u0000"+
		"\u0000\u0261\u0265\u0001\u0000\u0000\u0000\u0262\u0265\u00055\u0000\u0000"+
		"\u0263\u0265\u0005*\u0000\u0000\u0264\u025e\u0001\u0000\u0000\u0000\u0264"+
		"\u0262\u0001\u0000\u0000\u0000\u0264\u0263\u0001\u0000\u0000\u0000\u0265"+
		"I\u0001\u0000\u0000\u0000\u0266\u0269\u0003d2\u0000\u0267\u0269\u0003"+
		"P(\u0000\u0268\u0266\u0001\u0000\u0000\u0000\u0268\u0267\u0001\u0000\u0000"+
		"\u0000\u0269\u026a\u0001\u0000\u0000\u0000\u026a\u026b\u0005\b\u0000\u0000"+
		"\u026b\u026c\u0003\u001a\r\u0000\u026cK\u0001\u0000\u0000\u0000\u026d"+
		"\u026e\u0005Q\u0000\u0000\u026e\u026f\u0003R)\u0000\u026f\u0270\u0005"+
		"N\u0000\u0000\u0270\u0275\u0003J%\u0000\u0271\u0272\u0005\u0007\u0000"+
		"\u0000\u0272\u0274\u0003J%\u0000\u0273\u0271\u0001\u0000\u0000\u0000\u0274"+
		"\u0277\u0001\u0000\u0000\u0000\u0275\u0273\u0001\u0000\u0000\u0000\u0275"+
		"\u0276\u0001\u0000\u0000\u0000\u0276\u027a\u0001\u0000\u0000\u0000\u0277"+
		"\u0275\u0001\u0000\u0000\u0000\u0278\u0279\u0005.\u0000\u0000\u0279\u027b"+
		"\u0003:\u001d\u0000\u027a\u0278\u0001\u0000\u0000\u0000\u027a\u027b\u0001"+
		"\u0000\u0000\u0000\u027b\u027e\u0001\u0000\u0000\u0000\u027c\u027d\u0005"+
		"U\u0000\u0000\u027d\u027f\u0003\u001a\r\u0000\u027e\u027c\u0001\u0000"+
		"\u0000\u0000\u027e\u027f\u0001\u0000\u0000\u0000\u027f\u0281\u0001\u0000"+
		"\u0000\u0000\u0280\u0282\u0003.\u0017\u0000\u0281\u0280\u0001\u0000\u0000"+
		"\u0000\u0281\u0282\u0001\u0000\u0000\u0000\u0282M\u0001\u0000\u0000\u0000"+
		"\u0283\u0285\u0003\f\u0006\u0000\u0284\u0283\u0001\u0000\u0000\u0000\u0284"+
		"\u0285\u0001\u0000\u0000\u0000\u0285\u0286\u0001\u0000\u0000\u0000\u0286"+
		"\u0287\u0003L&\u0000\u0287O\u0001\u0000\u0000\u0000\u0288\u0289\u0005"+
		"\u0003\u0000\u0000\u0289\u028e\u0003d2\u0000\u028a\u028b\u0005\u0007\u0000"+
		"\u0000\u028b\u028d\u0003d2\u0000\u028c\u028a\u0001\u0000\u0000\u0000\u028d"+
		"\u0290\u0001\u0000\u0000\u0000\u028e\u028c\u0001\u0000\u0000\u0000\u028e"+
		"\u028f\u0001\u0000\u0000\u0000\u028f\u0291\u0001\u0000\u0000\u0000\u0290"+
		"\u028e\u0001\u0000\u0000\u0000\u0291\u0292\u0005\u0004\u0000\u0000\u0292"+
		"Q\u0001\u0000\u0000\u0000\u0293\u0296\u0003`0\u0000\u0294\u0295\u0005"+
		"\u0019\u0000\u0000\u0295\u0297\u0003b1\u0000\u0296\u0294\u0001\u0000\u0000"+
		"\u0000\u0296\u0297\u0001\u0000\u0000\u0000\u0297S\u0001\u0000\u0000\u0000"+
		"\u0298\u0299\u0005F\u0000\u0000\u0299\u029a\u0005\u001b\u0000\u0000\u029a"+
		"\u029f\u0003X,\u0000\u029b\u029c\u0005\u0007\u0000\u0000\u029c\u029e\u0003"+
		"X,\u0000\u029d\u029b\u0001\u0000\u0000\u0000\u029e\u02a1\u0001\u0000\u0000"+
		"\u0000\u029f\u029d\u0001\u0000\u0000\u0000\u029f\u02a0\u0001\u0000\u0000"+
		"\u0000\u02a0U\u0001\u0000\u0000\u0000\u02a1\u029f\u0001\u0000\u0000\u0000"+
		"\u02a2\u02a3\u0005>\u0000\u0000\u02a3\u02a6\u0003\u001a\r\u0000\u02a4"+
		"\u02a5\u0005C\u0000\u0000\u02a5\u02a7\u0003\u001a\r\u0000\u02a6\u02a4"+
		"\u0001\u0000\u0000\u0000\u02a6\u02a7\u0001\u0000\u0000\u0000\u02a7W\u0001"+
		"\u0000\u0000\u0000\u02a8\u02aa\u0003\u001a\r\u0000\u02a9\u02ab\u0003Z"+
		"-\u0000\u02aa\u02a9\u0001\u0000\u0000\u0000\u02aa\u02ab\u0001\u0000\u0000"+
		"\u0000\u02ab\u02ae\u0001\u0000\u0000\u0000\u02ac\u02ad\u0005B\u0000\u0000"+
		"\u02ad\u02af\u0007\u0005\u0000\u0000\u02ae\u02ac\u0001\u0000\u0000\u0000"+
		"\u02ae\u02af\u0001\u0000\u0000\u0000\u02afY\u0001\u0000\u0000\u0000\u02b0"+
		"\u02b1\u0007\u0006\u0000\u0000\u02b1[\u0001\u0000\u0000\u0000\u02b2\u02b6"+
		"\u0001\u0000\u0000\u0000\u02b3\u02b6\u0005=\u0000\u0000\u02b4\u02b6\u0005"+
		"J\u0000\u0000\u02b5\u02b2\u0001\u0000\u0000\u0000\u02b5\u02b3\u0001\u0000"+
		"\u0000\u0000\u02b5\u02b4\u0001\u0000\u0000\u0000\u02b6]\u0001\u0000\u0000"+
		"\u0000\u02b7\u02ba\u0005\\\u0000\u0000\u02b8\u02ba\u0003\\.\u0000\u02b9"+
		"\u02b7\u0001\u0000\u0000\u0000\u02b9\u02b8\u0001\u0000\u0000\u0000\u02ba"+
		"_\u0001\u0000\u0000\u0000\u02bb\u02bc\u0005\\\u0000\u0000\u02bca\u0001"+
		"\u0000\u0000\u0000\u02bd\u02be\u0005\\\u0000\u0000\u02bec\u0001\u0000"+
		"\u0000\u0000\u02bf\u02c0\u0005\\\u0000\u0000\u02c0e\u0001\u0000\u0000"+
		"\u0000\u02c1\u02c2\u0005\\\u0000\u0000\u02c2g\u0001\u0000\u0000\u0000"+
		"\u02c3\u02c4\u0005\\\u0000\u0000\u02c4i\u0001\u0000\u0000\u0000\u02c5"+
		"\u02c6\u0005\\\u0000\u0000\u02c6k\u0001\u0000\u0000\u0000_ow~\u0083\u0089"+
		"\u0090\u009b\u00a0\u00ae\u00b6\u00b9\u00bc\u00c5\u00cc\u00d0\u00d7\u00e3"+
		"\u00e7\u00eb\u00ef\u00f3\u00f7\u00fb\u0103\u0106\u0109\u010e\u0113\u0117"+
		"\u0121\u0125\u012f\u0145\u014b\u0154\u015a\u015f\u0166\u016a\u016c\u0178"+
		"\u0180\u018b\u0196\u019e\u01a6\u01ab\u01af\u01b2\u01b5\u01bf\u01c4\u01d1"+
		"\u01d7\u01d9\u01e4\u01e9\u01eb\u01f3\u01f7\u01fa\u01fd\u0209\u020e\u0215"+
		"\u021a\u021e\u0227\u022c\u022e\u0233\u023a\u023c\u0246\u0248\u024e\u0250"+
		"\u0254\u0257\u0260\u0264\u0268\u0275\u027a\u027e\u0281\u0284\u028e\u0296"+
		"\u029f\u02a6\u02aa\u02ae\u02b5\u02b9";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}