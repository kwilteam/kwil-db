// Generated from /Users/brennanlamey/kwil-db/internal/parse/action/grammar/ActionParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class ActionParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		SCOL=1, L_PAREN=2, R_PAREN=3, COMMA=4, DOLLAR=5, AT=6, ASSIGN=7, PERIOD=8, 
		PLUS=9, MINUS=10, STAR=11, DIV=12, MOD=13, LT=14, LT_EQ=15, GT=16, GT_EQ=17, 
		SQL_NOT_EQ1=18, SQL_NOT_EQ2=19, SELECT_=20, INSERT_=21, UPDATE_=22, DELETE_=23, 
		WITH_=24, NOT_=25, AND_=26, OR_=27, SQL_KEYWORDS=28, SQL_STMT=29, IDENTIFIER=30, 
		VARIABLE=31, BLOCK_VARIABLE=32, UNSIGNED_NUMBER_LITERAL=33, STRING_LITERAL=34, 
		WS=35, TERMINATOR=36, BLOCK_COMMENT=37, LINE_COMMENT=38;
	public static final int
		RULE_statement = 0, RULE_literal_value = 1, RULE_action_name = 2, RULE_stmt = 3, 
		RULE_sql_stmt = 4, RULE_call_stmt = 5, RULE_call_receivers = 6, RULE_call_body = 7, 
		RULE_variable = 8, RULE_block_var = 9, RULE_extension_call_name = 10, 
		RULE_fn_name = 11, RULE_sfn_name = 12, RULE_fn_arg_list = 13, RULE_fn_arg_expr = 14;
	private static String[] makeRuleNames() {
		return new String[] {
			"statement", "literal_value", "action_name", "stmt", "sql_stmt", "call_stmt", 
			"call_receivers", "call_body", "variable", "block_var", "extension_call_name", 
			"fn_name", "sfn_name", "fn_arg_list", "fn_arg_expr"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "';'", "'('", "')'", "','", "'$'", "'@'", "'='", "'.'", "'+'", 
			"'-'", "'*'", "'/'", "'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'<>'", 
			null, null, null, null, null, "'not'", "'and'", "'or'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "SCOL", "L_PAREN", "R_PAREN", "COMMA", "DOLLAR", "AT", "ASSIGN", 
			"PERIOD", "PLUS", "MINUS", "STAR", "DIV", "MOD", "LT", "LT_EQ", "GT", 
			"GT_EQ", "SQL_NOT_EQ1", "SQL_NOT_EQ2", "SELECT_", "INSERT_", "UPDATE_", 
			"DELETE_", "WITH_", "NOT_", "AND_", "OR_", "SQL_KEYWORDS", "SQL_STMT", 
			"IDENTIFIER", "VARIABLE", "BLOCK_VARIABLE", "UNSIGNED_NUMBER_LITERAL", 
			"STRING_LITERAL", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT"
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
	public String getGrammarFileName() { return "ActionParser.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public ActionParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StatementContext extends ParserRuleContext {
		public List<StmtContext> stmt() {
			return getRuleContexts(StmtContext.class);
		}
		public StmtContext stmt(int i) {
			return getRuleContext(StmtContext.class,i);
		}
		public StatementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statement; }
	}

	public final StatementContext statement() throws RecognitionException {
		StatementContext _localctx = new StatementContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(31); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(30);
				stmt();
				}
				}
				setState(33); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( (((_la) & ~0x3f) == 0 && ((1L << _la) & 3758096384L) != 0) );
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
	public static class Literal_valueContext extends ParserRuleContext {
		public TerminalNode STRING_LITERAL() { return getToken(ActionParser.STRING_LITERAL, 0); }
		public TerminalNode UNSIGNED_NUMBER_LITERAL() { return getToken(ActionParser.UNSIGNED_NUMBER_LITERAL, 0); }
		public Literal_valueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_literal_value; }
	}

	public final Literal_valueContext literal_value() throws RecognitionException {
		Literal_valueContext _localctx = new Literal_valueContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_literal_value);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(35);
			_la = _input.LA(1);
			if ( !(_la==UNSIGNED_NUMBER_LITERAL || _la==STRING_LITERAL) ) {
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
	public static class Action_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(ActionParser.IDENTIFIER, 0); }
		public Action_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_action_name; }
	}

	public final Action_nameContext action_name() throws RecognitionException {
		Action_nameContext _localctx = new Action_nameContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_action_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(37);
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
	public static class StmtContext extends ParserRuleContext {
		public Sql_stmtContext sql_stmt() {
			return getRuleContext(Sql_stmtContext.class,0);
		}
		public Call_stmtContext call_stmt() {
			return getRuleContext(Call_stmtContext.class,0);
		}
		public StmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stmt; }
	}

	public final StmtContext stmt() throws RecognitionException {
		StmtContext _localctx = new StmtContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_stmt);
		try {
			setState(41);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case SQL_STMT:
				enterOuterAlt(_localctx, 1);
				{
				setState(39);
				sql_stmt();
				}
				break;
			case IDENTIFIER:
			case VARIABLE:
				enterOuterAlt(_localctx, 2);
				{
				setState(40);
				call_stmt();
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
	public static class Sql_stmtContext extends ParserRuleContext {
		public TerminalNode SQL_STMT() { return getToken(ActionParser.SQL_STMT, 0); }
		public TerminalNode SCOL() { return getToken(ActionParser.SCOL, 0); }
		public Sql_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sql_stmt; }
	}

	public final Sql_stmtContext sql_stmt() throws RecognitionException {
		Sql_stmtContext _localctx = new Sql_stmtContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_sql_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(43);
			match(SQL_STMT);
			setState(44);
			match(SCOL);
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
	public static class Call_stmtContext extends ParserRuleContext {
		public Call_bodyContext call_body() {
			return getRuleContext(Call_bodyContext.class,0);
		}
		public TerminalNode SCOL() { return getToken(ActionParser.SCOL, 0); }
		public Call_receiversContext call_receivers() {
			return getRuleContext(Call_receiversContext.class,0);
		}
		public TerminalNode ASSIGN() { return getToken(ActionParser.ASSIGN, 0); }
		public Call_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_stmt; }
	}

	public final Call_stmtContext call_stmt() throws RecognitionException {
		Call_stmtContext _localctx = new Call_stmtContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_call_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(49);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==VARIABLE) {
				{
				setState(46);
				call_receivers();
				setState(47);
				match(ASSIGN);
				}
			}

			setState(51);
			call_body();
			setState(52);
			match(SCOL);
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
	public static class Call_receiversContext extends ParserRuleContext {
		public List<VariableContext> variable() {
			return getRuleContexts(VariableContext.class);
		}
		public VariableContext variable(int i) {
			return getRuleContext(VariableContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(ActionParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(ActionParser.COMMA, i);
		}
		public Call_receiversContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_receivers; }
	}

	public final Call_receiversContext call_receivers() throws RecognitionException {
		Call_receiversContext _localctx = new Call_receiversContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_call_receivers);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(54);
			variable();
			setState(59);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(55);
				match(COMMA);
				setState(56);
				variable();
				}
				}
				setState(61);
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
	public static class Call_bodyContext extends ParserRuleContext {
		public Fn_nameContext fn_name() {
			return getRuleContext(Fn_nameContext.class,0);
		}
		public TerminalNode L_PAREN() { return getToken(ActionParser.L_PAREN, 0); }
		public Fn_arg_listContext fn_arg_list() {
			return getRuleContext(Fn_arg_listContext.class,0);
		}
		public TerminalNode R_PAREN() { return getToken(ActionParser.R_PAREN, 0); }
		public Call_bodyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_body; }
	}

	public final Call_bodyContext call_body() throws RecognitionException {
		Call_bodyContext _localctx = new Call_bodyContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_call_body);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(62);
			fn_name();
			setState(63);
			match(L_PAREN);
			setState(64);
			fn_arg_list();
			setState(65);
			match(R_PAREN);
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
		public TerminalNode VARIABLE() { return getToken(ActionParser.VARIABLE, 0); }
		public VariableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_variable; }
	}

	public final VariableContext variable() throws RecognitionException {
		VariableContext _localctx = new VariableContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_variable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(67);
			match(VARIABLE);
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
	public static class Block_varContext extends ParserRuleContext {
		public TerminalNode BLOCK_VARIABLE() { return getToken(ActionParser.BLOCK_VARIABLE, 0); }
		public Block_varContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_block_var; }
	}

	public final Block_varContext block_var() throws RecognitionException {
		Block_varContext _localctx = new Block_varContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_block_var);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(69);
			match(BLOCK_VARIABLE);
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
	public static class Extension_call_nameContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(ActionParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(ActionParser.IDENTIFIER, i);
		}
		public TerminalNode PERIOD() { return getToken(ActionParser.PERIOD, 0); }
		public Extension_call_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_extension_call_name; }
	}

	public final Extension_call_nameContext extension_call_name() throws RecognitionException {
		Extension_call_nameContext _localctx = new Extension_call_nameContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_extension_call_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(71);
			match(IDENTIFIER);
			setState(72);
			match(PERIOD);
			setState(73);
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
	public static class Fn_nameContext extends ParserRuleContext {
		public Extension_call_nameContext extension_call_name() {
			return getRuleContext(Extension_call_nameContext.class,0);
		}
		public Action_nameContext action_name() {
			return getRuleContext(Action_nameContext.class,0);
		}
		public Fn_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fn_name; }
	}

	public final Fn_nameContext fn_name() throws RecognitionException {
		Fn_nameContext _localctx = new Fn_nameContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_fn_name);
		try {
			setState(77);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,4,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(75);
				extension_call_name();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(76);
				action_name();
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
	public static class Sfn_nameContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(ActionParser.IDENTIFIER, 0); }
		public Sfn_nameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sfn_name; }
	}

	public final Sfn_nameContext sfn_name() throws RecognitionException {
		Sfn_nameContext _localctx = new Sfn_nameContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_sfn_name);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(79);
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
	public static class Fn_arg_listContext extends ParserRuleContext {
		public List<Fn_arg_exprContext> fn_arg_expr() {
			return getRuleContexts(Fn_arg_exprContext.class);
		}
		public Fn_arg_exprContext fn_arg_expr(int i) {
			return getRuleContext(Fn_arg_exprContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(ActionParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(ActionParser.COMMA, i);
		}
		public Fn_arg_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fn_arg_list; }
	}

	public final Fn_arg_listContext fn_arg_list() throws RecognitionException {
		Fn_arg_listContext _localctx = new Fn_arg_listContext(_ctx, getState());
		enterRule(_localctx, 26, RULE_fn_arg_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(82);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 33319552516L) != 0)) {
				{
				setState(81);
				fn_arg_expr(0);
				}
			}

			setState(88);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(84);
				match(COMMA);
				setState(85);
				fn_arg_expr(0);
				}
				}
				setState(90);
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
	public static class Fn_arg_exprContext extends ParserRuleContext {
		public Fn_arg_exprContext elevate_expr;
		public Fn_arg_exprContext unary_expr;
		public Literal_valueContext literal_value() {
			return getRuleContext(Literal_valueContext.class,0);
		}
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public Block_varContext block_var() {
			return getRuleContext(Block_varContext.class,0);
		}
		public Sfn_nameContext sfn_name() {
			return getRuleContext(Sfn_nameContext.class,0);
		}
		public TerminalNode L_PAREN() { return getToken(ActionParser.L_PAREN, 0); }
		public TerminalNode R_PAREN() { return getToken(ActionParser.R_PAREN, 0); }
		public TerminalNode STAR() { return getToken(ActionParser.STAR, 0); }
		public List<Fn_arg_exprContext> fn_arg_expr() {
			return getRuleContexts(Fn_arg_exprContext.class);
		}
		public Fn_arg_exprContext fn_arg_expr(int i) {
			return getRuleContext(Fn_arg_exprContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(ActionParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(ActionParser.COMMA, i);
		}
		public TerminalNode MINUS() { return getToken(ActionParser.MINUS, 0); }
		public TerminalNode PLUS() { return getToken(ActionParser.PLUS, 0); }
		public TerminalNode NOT_() { return getToken(ActionParser.NOT_, 0); }
		public TerminalNode DIV() { return getToken(ActionParser.DIV, 0); }
		public TerminalNode MOD() { return getToken(ActionParser.MOD, 0); }
		public TerminalNode LT() { return getToken(ActionParser.LT, 0); }
		public TerminalNode LT_EQ() { return getToken(ActionParser.LT_EQ, 0); }
		public TerminalNode GT() { return getToken(ActionParser.GT, 0); }
		public TerminalNode GT_EQ() { return getToken(ActionParser.GT_EQ, 0); }
		public TerminalNode ASSIGN() { return getToken(ActionParser.ASSIGN, 0); }
		public TerminalNode SQL_NOT_EQ1() { return getToken(ActionParser.SQL_NOT_EQ1, 0); }
		public TerminalNode SQL_NOT_EQ2() { return getToken(ActionParser.SQL_NOT_EQ2, 0); }
		public TerminalNode AND_() { return getToken(ActionParser.AND_, 0); }
		public TerminalNode OR_() { return getToken(ActionParser.OR_, 0); }
		public Fn_arg_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fn_arg_expr; }
	}

	public final Fn_arg_exprContext fn_arg_expr() throws RecognitionException {
		return fn_arg_expr(0);
	}

	private Fn_arg_exprContext fn_arg_expr(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		Fn_arg_exprContext _localctx = new Fn_arg_exprContext(_ctx, _parentState);
		Fn_arg_exprContext _prevctx = _localctx;
		int _startState = 28;
		enterRecursionRule(_localctx, 28, RULE_fn_arg_expr, _p);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(118);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case UNSIGNED_NUMBER_LITERAL:
			case STRING_LITERAL:
				{
				setState(92);
				literal_value();
				}
				break;
			case VARIABLE:
				{
				setState(93);
				variable();
				}
				break;
			case BLOCK_VARIABLE:
				{
				setState(94);
				block_var();
				}
				break;
			case IDENTIFIER:
				{
				setState(95);
				sfn_name();
				setState(96);
				match(L_PAREN);
				setState(106);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case L_PAREN:
				case PLUS:
				case MINUS:
				case NOT_:
				case IDENTIFIER:
				case VARIABLE:
				case BLOCK_VARIABLE:
				case UNSIGNED_NUMBER_LITERAL:
				case STRING_LITERAL:
					{
					{
					setState(97);
					fn_arg_expr(0);
					setState(102);
					_errHandler.sync(this);
					_la = _input.LA(1);
					while (_la==COMMA) {
						{
						{
						setState(98);
						match(COMMA);
						setState(99);
						fn_arg_expr(0);
						}
						}
						setState(104);
						_errHandler.sync(this);
						_la = _input.LA(1);
					}
					}
					}
					break;
				case STAR:
					{
					setState(105);
					match(STAR);
					}
					break;
				case R_PAREN:
					break;
				default:
					break;
				}
				setState(108);
				match(R_PAREN);
				}
				break;
			case L_PAREN:
				{
				setState(110);
				match(L_PAREN);
				setState(111);
				((Fn_arg_exprContext)_localctx).elevate_expr = fn_arg_expr(0);
				setState(112);
				match(R_PAREN);
				}
				break;
			case PLUS:
			case MINUS:
				{
				setState(114);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(115);
				((Fn_arg_exprContext)_localctx).unary_expr = fn_arg_expr(8);
				}
				break;
			case NOT_:
				{
				setState(116);
				match(NOT_);
				setState(117);
				((Fn_arg_exprContext)_localctx).unary_expr = fn_arg_expr(3);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			_ctx.stop = _input.LT(-1);
			setState(140);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,11,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(138);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,10,_ctx) ) {
					case 1:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(120);
						if (!(precpred(_ctx, 7))) throw new FailedPredicateException(this, "precpred(_ctx, 7)");
						setState(121);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 14336L) != 0)) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(122);
						fn_arg_expr(8);
						}
						break;
					case 2:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(123);
						if (!(precpred(_ctx, 6))) throw new FailedPredicateException(this, "precpred(_ctx, 6)");
						setState(124);
						_la = _input.LA(1);
						if ( !(_la==PLUS || _la==MINUS) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(125);
						fn_arg_expr(7);
						}
						break;
					case 3:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(126);
						if (!(precpred(_ctx, 5))) throw new FailedPredicateException(this, "precpred(_ctx, 5)");
						setState(127);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 245760L) != 0)) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(128);
						fn_arg_expr(6);
						}
						break;
					case 4:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(129);
						if (!(precpred(_ctx, 4))) throw new FailedPredicateException(this, "precpred(_ctx, 4)");
						setState(130);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 786560L) != 0)) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(131);
						fn_arg_expr(5);
						}
						break;
					case 5:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(132);
						if (!(precpred(_ctx, 2))) throw new FailedPredicateException(this, "precpred(_ctx, 2)");
						setState(133);
						match(AND_);
						setState(134);
						fn_arg_expr(3);
						}
						break;
					case 6:
						{
						_localctx = new Fn_arg_exprContext(_parentctx, _parentState);
						pushNewRecursionContext(_localctx, _startState, RULE_fn_arg_expr);
						setState(135);
						if (!(precpred(_ctx, 1))) throw new FailedPredicateException(this, "precpred(_ctx, 1)");
						setState(136);
						match(OR_);
						setState(137);
						fn_arg_expr(2);
						}
						break;
					}
					} 
				}
				setState(142);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,11,_ctx);
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

	public boolean sempred(RuleContext _localctx, int ruleIndex, int predIndex) {
		switch (ruleIndex) {
		case 14:
			return fn_arg_expr_sempred((Fn_arg_exprContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean fn_arg_expr_sempred(Fn_arg_exprContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 7);
		case 1:
			return precpred(_ctx, 6);
		case 2:
			return precpred(_ctx, 5);
		case 3:
			return precpred(_ctx, 4);
		case 4:
			return precpred(_ctx, 2);
		case 5:
			return precpred(_ctx, 1);
		}
		return true;
	}

	public static final String _serializedATN =
		"\u0004\u0001&\u0090\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0001\u0000\u0004\u0000"+
		" \b\u0000\u000b\u0000\f\u0000!\u0001\u0001\u0001\u0001\u0001\u0002\u0001"+
		"\u0002\u0001\u0003\u0001\u0003\u0003\u0003*\b\u0003\u0001\u0004\u0001"+
		"\u0004\u0001\u0004\u0001\u0005\u0001\u0005\u0001\u0005\u0003\u00052\b"+
		"\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0006\u0001\u0006\u0001"+
		"\u0006\u0005\u0006:\b\u0006\n\u0006\f\u0006=\t\u0006\u0001\u0007\u0001"+
		"\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0001\b\u0001\b\u0001\t\u0001"+
		"\t\u0001\n\u0001\n\u0001\n\u0001\n\u0001\u000b\u0001\u000b\u0003\u000b"+
		"N\b\u000b\u0001\f\u0001\f\u0001\r\u0003\rS\b\r\u0001\r\u0001\r\u0005\r"+
		"W\b\r\n\r\f\rZ\t\r\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0005\u000ee\b"+
		"\u000e\n\u000e\f\u000eh\t\u000e\u0001\u000e\u0003\u000ek\b\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0003\u000ew\b\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0005"+
		"\u000e\u008b\b\u000e\n\u000e\f\u000e\u008e\t\u000e\u0001\u000e\u0000\u0001"+
		"\u001c\u000f\u0000\u0002\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016"+
		"\u0018\u001a\u001c\u0000\u0005\u0001\u0000!\"\u0001\u0000\t\n\u0001\u0000"+
		"\u000b\r\u0001\u0000\u000e\u0011\u0002\u0000\u0007\u0007\u0012\u0013\u0096"+
		"\u0000\u001f\u0001\u0000\u0000\u0000\u0002#\u0001\u0000\u0000\u0000\u0004"+
		"%\u0001\u0000\u0000\u0000\u0006)\u0001\u0000\u0000\u0000\b+\u0001\u0000"+
		"\u0000\u0000\n1\u0001\u0000\u0000\u0000\f6\u0001\u0000\u0000\u0000\u000e"+
		">\u0001\u0000\u0000\u0000\u0010C\u0001\u0000\u0000\u0000\u0012E\u0001"+
		"\u0000\u0000\u0000\u0014G\u0001\u0000\u0000\u0000\u0016M\u0001\u0000\u0000"+
		"\u0000\u0018O\u0001\u0000\u0000\u0000\u001aR\u0001\u0000\u0000\u0000\u001c"+
		"v\u0001\u0000\u0000\u0000\u001e \u0003\u0006\u0003\u0000\u001f\u001e\u0001"+
		"\u0000\u0000\u0000 !\u0001\u0000\u0000\u0000!\u001f\u0001\u0000\u0000"+
		"\u0000!\"\u0001\u0000\u0000\u0000\"\u0001\u0001\u0000\u0000\u0000#$\u0007"+
		"\u0000\u0000\u0000$\u0003\u0001\u0000\u0000\u0000%&\u0005\u001e\u0000"+
		"\u0000&\u0005\u0001\u0000\u0000\u0000\'*\u0003\b\u0004\u0000(*\u0003\n"+
		"\u0005\u0000)\'\u0001\u0000\u0000\u0000)(\u0001\u0000\u0000\u0000*\u0007"+
		"\u0001\u0000\u0000\u0000+,\u0005\u001d\u0000\u0000,-\u0005\u0001\u0000"+
		"\u0000-\t\u0001\u0000\u0000\u0000./\u0003\f\u0006\u0000/0\u0005\u0007"+
		"\u0000\u000002\u0001\u0000\u0000\u00001.\u0001\u0000\u0000\u000012\u0001"+
		"\u0000\u0000\u000023\u0001\u0000\u0000\u000034\u0003\u000e\u0007\u0000"+
		"45\u0005\u0001\u0000\u00005\u000b\u0001\u0000\u0000\u00006;\u0003\u0010"+
		"\b\u000078\u0005\u0004\u0000\u00008:\u0003\u0010\b\u000097\u0001\u0000"+
		"\u0000\u0000:=\u0001\u0000\u0000\u0000;9\u0001\u0000\u0000\u0000;<\u0001"+
		"\u0000\u0000\u0000<\r\u0001\u0000\u0000\u0000=;\u0001\u0000\u0000\u0000"+
		">?\u0003\u0016\u000b\u0000?@\u0005\u0002\u0000\u0000@A\u0003\u001a\r\u0000"+
		"AB\u0005\u0003\u0000\u0000B\u000f\u0001\u0000\u0000\u0000CD\u0005\u001f"+
		"\u0000\u0000D\u0011\u0001\u0000\u0000\u0000EF\u0005 \u0000\u0000F\u0013"+
		"\u0001\u0000\u0000\u0000GH\u0005\u001e\u0000\u0000HI\u0005\b\u0000\u0000"+
		"IJ\u0005\u001e\u0000\u0000J\u0015\u0001\u0000\u0000\u0000KN\u0003\u0014"+
		"\n\u0000LN\u0003\u0004\u0002\u0000MK\u0001\u0000\u0000\u0000ML\u0001\u0000"+
		"\u0000\u0000N\u0017\u0001\u0000\u0000\u0000OP\u0005\u001e\u0000\u0000"+
		"P\u0019\u0001\u0000\u0000\u0000QS\u0003\u001c\u000e\u0000RQ\u0001\u0000"+
		"\u0000\u0000RS\u0001\u0000\u0000\u0000SX\u0001\u0000\u0000\u0000TU\u0005"+
		"\u0004\u0000\u0000UW\u0003\u001c\u000e\u0000VT\u0001\u0000\u0000\u0000"+
		"WZ\u0001\u0000\u0000\u0000XV\u0001\u0000\u0000\u0000XY\u0001\u0000\u0000"+
		"\u0000Y\u001b\u0001\u0000\u0000\u0000ZX\u0001\u0000\u0000\u0000[\\\u0006"+
		"\u000e\uffff\uffff\u0000\\w\u0003\u0002\u0001\u0000]w\u0003\u0010\b\u0000"+
		"^w\u0003\u0012\t\u0000_`\u0003\u0018\f\u0000`j\u0005\u0002\u0000\u0000"+
		"af\u0003\u001c\u000e\u0000bc\u0005\u0004\u0000\u0000ce\u0003\u001c\u000e"+
		"\u0000db\u0001\u0000\u0000\u0000eh\u0001\u0000\u0000\u0000fd\u0001\u0000"+
		"\u0000\u0000fg\u0001\u0000\u0000\u0000gk\u0001\u0000\u0000\u0000hf\u0001"+
		"\u0000\u0000\u0000ik\u0005\u000b\u0000\u0000ja\u0001\u0000\u0000\u0000"+
		"ji\u0001\u0000\u0000\u0000jk\u0001\u0000\u0000\u0000kl\u0001\u0000\u0000"+
		"\u0000lm\u0005\u0003\u0000\u0000mw\u0001\u0000\u0000\u0000no\u0005\u0002"+
		"\u0000\u0000op\u0003\u001c\u000e\u0000pq\u0005\u0003\u0000\u0000qw\u0001"+
		"\u0000\u0000\u0000rs\u0007\u0001\u0000\u0000sw\u0003\u001c\u000e\btu\u0005"+
		"\u0019\u0000\u0000uw\u0003\u001c\u000e\u0003v[\u0001\u0000\u0000\u0000"+
		"v]\u0001\u0000\u0000\u0000v^\u0001\u0000\u0000\u0000v_\u0001\u0000\u0000"+
		"\u0000vn\u0001\u0000\u0000\u0000vr\u0001\u0000\u0000\u0000vt\u0001\u0000"+
		"\u0000\u0000w\u008c\u0001\u0000\u0000\u0000xy\n\u0007\u0000\u0000yz\u0007"+
		"\u0002\u0000\u0000z\u008b\u0003\u001c\u000e\b{|\n\u0006\u0000\u0000|}"+
		"\u0007\u0001\u0000\u0000}\u008b\u0003\u001c\u000e\u0007~\u007f\n\u0005"+
		"\u0000\u0000\u007f\u0080\u0007\u0003\u0000\u0000\u0080\u008b\u0003\u001c"+
		"\u000e\u0006\u0081\u0082\n\u0004\u0000\u0000\u0082\u0083\u0007\u0004\u0000"+
		"\u0000\u0083\u008b\u0003\u001c\u000e\u0005\u0084\u0085\n\u0002\u0000\u0000"+
		"\u0085\u0086\u0005\u001a\u0000\u0000\u0086\u008b\u0003\u001c\u000e\u0003"+
		"\u0087\u0088\n\u0001\u0000\u0000\u0088\u0089\u0005\u001b\u0000\u0000\u0089"+
		"\u008b\u0003\u001c\u000e\u0002\u008ax\u0001\u0000\u0000\u0000\u008a{\u0001"+
		"\u0000\u0000\u0000\u008a~\u0001\u0000\u0000\u0000\u008a\u0081\u0001\u0000"+
		"\u0000\u0000\u008a\u0084\u0001\u0000\u0000\u0000\u008a\u0087\u0001\u0000"+
		"\u0000\u0000\u008b\u008e\u0001\u0000\u0000\u0000\u008c\u008a\u0001\u0000"+
		"\u0000\u0000\u008c\u008d\u0001\u0000\u0000\u0000\u008d\u001d\u0001\u0000"+
		"\u0000\u0000\u008e\u008c\u0001\u0000\u0000\u0000\f!)1;MRXfjv\u008a\u008c";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}