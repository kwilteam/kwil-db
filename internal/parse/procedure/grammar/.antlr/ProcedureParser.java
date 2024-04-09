// Generated from /Users/brennanlamey/kwil-db/internal/engine/procedural/parser/grammar/ProcedureParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class ProcedureParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		SEMICOLON=1, LPAREN=2, RPAREN=3, LBRACE=4, RBRACE=5, COMMA=6, COLON=7, 
		DOLLAR=8, AT=9, ASSIGN=10, PERIOD=11, LBRACKET=12, RBRACKET=13, SINGLE_QUOTE=14, 
		PLUS=15, MINUS=16, MUL=17, DIV=18, MOD=19, LT=20, LT_EQ=21, GT=22, GT_EQ=23, 
		NEQ=24, EQ=25, ANY_SQL=26, FOR=27, IN=28, IF=29, ELSEIF=30, ELSE=31, TO=32, 
		RETURN=33, BREAK=34, NEXT=35, BOOLEAN_LITERAL=36, INT_LITERAL=37, BLOB_LITERAL=38, 
		TEXT_LITERAL=39, NULL_LITERAL=40, IDENTIFIER=41, VARIABLE=42, WS=43, TERMINATOR=44, 
		BLOCK_COMMENT=45, LINE_COMMENT=46;
	public static final int
		RULE_program = 0, RULE_statement = 1, RULE_type = 2, RULE_expression = 3, 
		RULE_expression_list = 4, RULE_expression_make_array = 5, RULE_call_expression = 6, 
		RULE_range = 7, RULE_if_then_block = 8;
	private static String[] makeRuleNames() {
		return new String[] {
			"program", "statement", "type", "expression", "expression_list", "expression_make_array", 
			"call_expression", "range", "if_then_block"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "';'", "'('", "')'", "'{'", "'}'", "','", "':'", "'$'", "'@'", 
			"':='", "'.'", "'['", "']'", "'''", "'+'", "'-'", "'*'", "'/'", "'%'", 
			"'<'", "'<='", "'>'", "'>='", "'!='", "'=='", null, "'for'", "'in'", 
			"'if'", "'elseif'", "'else'", "'to'", "'return'", "'break'", "'next'", 
			null, null, null, null, "'null'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "SEMICOLON", "LPAREN", "RPAREN", "LBRACE", "RBRACE", "COMMA", "COLON", 
			"DOLLAR", "AT", "ASSIGN", "PERIOD", "LBRACKET", "RBRACKET", "SINGLE_QUOTE", 
			"PLUS", "MINUS", "MUL", "DIV", "MOD", "LT", "LT_EQ", "GT", "GT_EQ", "NEQ", 
			"EQ", "ANY_SQL", "FOR", "IN", "IF", "ELSEIF", "ELSE", "TO", "RETURN", 
			"BREAK", "NEXT", "BOOLEAN_LITERAL", "INT_LITERAL", "BLOB_LITERAL", "TEXT_LITERAL", 
			"NULL_LITERAL", "IDENTIFIER", "VARIABLE", "WS", "TERMINATOR", "BLOCK_COMMENT", 
			"LINE_COMMENT"
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
	public String getGrammarFileName() { return "ProcedureParser.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public ProcedureParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ProgramContext extends ParserRuleContext {
		public List<StatementContext> statement() {
			return getRuleContexts(StatementContext.class);
		}
		public StatementContext statement(int i) {
			return getRuleContext(StatementContext.class,i);
		}
		public ProgramContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_program; }
	}

	public final ProgramContext program() throws RecognitionException {
		ProgramContext _localctx = new ProgramContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_program);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(21);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 6623577767936L) != 0)) {
				{
				{
				setState(18);
				statement();
				}
				}
				setState(23);
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
	public static class StatementContext extends ParserRuleContext {
		public StatementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statement; }
	 
		public StatementContext() { }
		public void copyFrom(StatementContext ctx) {
			super.copyFrom(ctx);
		}
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_ifContext extends StatementContext {
		public TerminalNode IF() { return getToken(ProcedureParser.IF, 0); }
		public List<If_then_blockContext> if_then_block() {
			return getRuleContexts(If_then_blockContext.class);
		}
		public If_then_blockContext if_then_block(int i) {
			return getRuleContext(If_then_blockContext.class,i);
		}
		public List<TerminalNode> ELSEIF() { return getTokens(ProcedureParser.ELSEIF); }
		public TerminalNode ELSEIF(int i) {
			return getToken(ProcedureParser.ELSEIF, i);
		}
		public TerminalNode ELSE() { return getToken(ProcedureParser.ELSE, 0); }
		public TerminalNode LBRACE() { return getToken(ProcedureParser.LBRACE, 0); }
		public TerminalNode RBRACE() { return getToken(ProcedureParser.RBRACE, 0); }
		public List<StatementContext> statement() {
			return getRuleContexts(StatementContext.class);
		}
		public StatementContext statement(int i) {
			return getRuleContext(StatementContext.class,i);
		}
		public Stmt_ifContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_breakContext extends StatementContext {
		public TerminalNode BREAK() { return getToken(ProcedureParser.BREAK, 0); }
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_breakContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_variable_assignment_with_declarationContext extends StatementContext {
		public TerminalNode VARIABLE() { return getToken(ProcedureParser.VARIABLE, 0); }
		public TypeContext type() {
			return getRuleContext(TypeContext.class,0);
		}
		public TerminalNode ASSIGN() { return getToken(ProcedureParser.ASSIGN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_variable_assignment_with_declarationContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_variable_declarationContext extends StatementContext {
		public TerminalNode VARIABLE() { return getToken(ProcedureParser.VARIABLE, 0); }
		public TypeContext type() {
			return getRuleContext(TypeContext.class,0);
		}
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_variable_declarationContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_return_nextContext extends StatementContext {
		public TerminalNode RETURN() { return getToken(ProcedureParser.RETURN, 0); }
		public TerminalNode NEXT() { return getToken(ProcedureParser.NEXT, 0); }
		public TerminalNode VARIABLE() { return getToken(ProcedureParser.VARIABLE, 0); }
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_return_nextContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_for_loopContext extends StatementContext {
		public TerminalNode FOR() { return getToken(ProcedureParser.FOR, 0); }
		public List<TerminalNode> VARIABLE() { return getTokens(ProcedureParser.VARIABLE); }
		public TerminalNode VARIABLE(int i) {
			return getToken(ProcedureParser.VARIABLE, i);
		}
		public TerminalNode IN() { return getToken(ProcedureParser.IN, 0); }
		public TerminalNode LBRACE() { return getToken(ProcedureParser.LBRACE, 0); }
		public TerminalNode RBRACE() { return getToken(ProcedureParser.RBRACE, 0); }
		public RangeContext range() {
			return getRuleContext(RangeContext.class,0);
		}
		public Call_expressionContext call_expression() {
			return getRuleContext(Call_expressionContext.class,0);
		}
		public TerminalNode ANY_SQL() { return getToken(ProcedureParser.ANY_SQL, 0); }
		public List<StatementContext> statement() {
			return getRuleContexts(StatementContext.class);
		}
		public StatementContext statement(int i) {
			return getRuleContext(StatementContext.class,i);
		}
		public Stmt_for_loopContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_returnContext extends StatementContext {
		public TerminalNode RETURN() { return getToken(ProcedureParser.RETURN, 0); }
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Expression_listContext expression_list() {
			return getRuleContext(Expression_listContext.class,0);
		}
		public TerminalNode ANY_SQL() { return getToken(ProcedureParser.ANY_SQL, 0); }
		public Stmt_returnContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_procedure_callContext extends StatementContext {
		public Call_expressionContext call_expression() {
			return getRuleContext(Call_expressionContext.class,0);
		}
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public List<TerminalNode> VARIABLE() { return getTokens(ProcedureParser.VARIABLE); }
		public TerminalNode VARIABLE(int i) {
			return getToken(ProcedureParser.VARIABLE, i);
		}
		public TerminalNode ASSIGN() { return getToken(ProcedureParser.ASSIGN, 0); }
		public TerminalNode COMMA() { return getToken(ProcedureParser.COMMA, 0); }
		public Stmt_procedure_callContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_variable_assignmentContext extends StatementContext {
		public TerminalNode VARIABLE() { return getToken(ProcedureParser.VARIABLE, 0); }
		public TerminalNode ASSIGN() { return getToken(ProcedureParser.ASSIGN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_variable_assignmentContext(StatementContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Stmt_sqlContext extends StatementContext {
		public TerminalNode ANY_SQL() { return getToken(ProcedureParser.ANY_SQL, 0); }
		public TerminalNode SEMICOLON() { return getToken(ProcedureParser.SEMICOLON, 0); }
		public Stmt_sqlContext(StatementContext ctx) { copyFrom(ctx); }
	}

	public final StatementContext statement() throws RecognitionException {
		StatementContext _localctx = new StatementContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_statement);
		int _la;
		try {
			setState(100);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,8,_ctx) ) {
			case 1:
				_localctx = new Stmt_variable_declarationContext(_localctx);
				enterOuterAlt(_localctx, 1);
				{
				setState(24);
				match(VARIABLE);
				setState(25);
				type();
				setState(26);
				match(SEMICOLON);
				}
				break;
			case 2:
				_localctx = new Stmt_variable_assignmentContext(_localctx);
				enterOuterAlt(_localctx, 2);
				{
				setState(28);
				match(VARIABLE);
				setState(29);
				match(ASSIGN);
				setState(30);
				expression(0);
				setState(31);
				match(SEMICOLON);
				}
				break;
			case 3:
				_localctx = new Stmt_variable_assignment_with_declarationContext(_localctx);
				enterOuterAlt(_localctx, 3);
				{
				setState(33);
				match(VARIABLE);
				setState(34);
				type();
				setState(35);
				match(ASSIGN);
				setState(36);
				expression(0);
				setState(37);
				match(SEMICOLON);
				}
				break;
			case 4:
				_localctx = new Stmt_procedure_callContext(_localctx);
				enterOuterAlt(_localctx, 4);
				{
				setState(44);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==VARIABLE) {
					{
					setState(39);
					match(VARIABLE);
					{
					setState(40);
					match(COMMA);
					setState(41);
					match(VARIABLE);
					}
					setState(43);
					match(ASSIGN);
					}
				}

				setState(46);
				call_expression();
				setState(47);
				match(SEMICOLON);
				}
				break;
			case 5:
				_localctx = new Stmt_for_loopContext(_localctx);
				enterOuterAlt(_localctx, 5);
				{
				setState(49);
				match(FOR);
				setState(50);
				match(VARIABLE);
				setState(51);
				match(IN);
				setState(56);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,2,_ctx) ) {
				case 1:
					{
					setState(52);
					range();
					}
					break;
				case 2:
					{
					setState(53);
					call_expression();
					}
					break;
				case 3:
					{
					setState(54);
					match(VARIABLE);
					}
					break;
				case 4:
					{
					setState(55);
					match(ANY_SQL);
					}
					break;
				}
				setState(58);
				match(LBRACE);
				setState(62);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 6623577767936L) != 0)) {
					{
					{
					setState(59);
					statement();
					}
					}
					setState(64);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(65);
				match(RBRACE);
				}
				break;
			case 6:
				_localctx = new Stmt_ifContext(_localctx);
				enterOuterAlt(_localctx, 6);
				{
				setState(66);
				match(IF);
				setState(67);
				if_then_block();
				setState(72);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==ELSEIF) {
					{
					{
					setState(68);
					match(ELSEIF);
					setState(69);
					if_then_block();
					}
					}
					setState(74);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(84);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==ELSE) {
					{
					setState(75);
					match(ELSE);
					setState(76);
					match(LBRACE);
					setState(80);
					_errHandler.sync(this);
					_la = _input.LA(1);
					while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 6623577767936L) != 0)) {
						{
						{
						setState(77);
						statement();
						}
						}
						setState(82);
						_errHandler.sync(this);
						_la = _input.LA(1);
					}
					setState(83);
					match(RBRACE);
					}
				}

				}
				break;
			case 7:
				_localctx = new Stmt_sqlContext(_localctx);
				enterOuterAlt(_localctx, 7);
				{
				setState(86);
				match(ANY_SQL);
				setState(87);
				match(SEMICOLON);
				}
				break;
			case 8:
				_localctx = new Stmt_breakContext(_localctx);
				enterOuterAlt(_localctx, 8);
				{
				setState(88);
				match(BREAK);
				setState(89);
				match(SEMICOLON);
				}
				break;
			case 9:
				_localctx = new Stmt_returnContext(_localctx);
				enterOuterAlt(_localctx, 9);
				{
				setState(90);
				match(RETURN);
				setState(93);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case LPAREN:
				case LBRACKET:
				case BOOLEAN_LITERAL:
				case INT_LITERAL:
				case BLOB_LITERAL:
				case TEXT_LITERAL:
				case NULL_LITERAL:
				case IDENTIFIER:
				case VARIABLE:
					{
					setState(91);
					expression_list();
					}
					break;
				case ANY_SQL:
					{
					setState(92);
					match(ANY_SQL);
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				setState(95);
				match(SEMICOLON);
				}
				break;
			case 10:
				_localctx = new Stmt_return_nextContext(_localctx);
				enterOuterAlt(_localctx, 10);
				{
				setState(96);
				match(RETURN);
				setState(97);
				match(NEXT);
				setState(98);
				match(VARIABLE);
				setState(99);
				match(SEMICOLON);
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
	public static class TypeContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(ProcedureParser.IDENTIFIER, 0); }
		public TerminalNode LBRACKET() { return getToken(ProcedureParser.LBRACKET, 0); }
		public TerminalNode RBRACKET() { return getToken(ProcedureParser.RBRACKET, 0); }
		public TypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_type; }
	}

	public final TypeContext type() throws RecognitionException {
		TypeContext _localctx = new TypeContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_type);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(102);
			match(IDENTIFIER);
			setState(105);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LBRACKET) {
				{
				setState(103);
				match(LBRACKET);
				setState(104);
				match(RBRACKET);
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
	public static class ExpressionContext extends ParserRuleContext {
		public ExpressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression; }
	 
		public ExpressionContext() { }
		public void copyFrom(ExpressionContext ctx) {
			super.copyFrom(ctx);
		}
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_array_accessContext extends ExpressionContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode LBRACKET() { return getToken(ProcedureParser.LBRACKET, 0); }
		public TerminalNode RBRACKET() { return getToken(ProcedureParser.RBRACKET, 0); }
		public Expr_array_accessContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_arithmeticContext extends ExpressionContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode MUL() { return getToken(ProcedureParser.MUL, 0); }
		public TerminalNode DIV() { return getToken(ProcedureParser.DIV, 0); }
		public TerminalNode MOD() { return getToken(ProcedureParser.MOD, 0); }
		public TerminalNode PLUS() { return getToken(ProcedureParser.PLUS, 0); }
		public TerminalNode MINUS() { return getToken(ProcedureParser.MINUS, 0); }
		public Expr_arithmeticContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_variableContext extends ExpressionContext {
		public TerminalNode VARIABLE() { return getToken(ProcedureParser.VARIABLE, 0); }
		public Expr_variableContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_null_literalContext extends ExpressionContext {
		public TerminalNode NULL_LITERAL() { return getToken(ProcedureParser.NULL_LITERAL, 0); }
		public Expr_null_literalContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_blob_literalContext extends ExpressionContext {
		public TerminalNode BLOB_LITERAL() { return getToken(ProcedureParser.BLOB_LITERAL, 0); }
		public Expr_blob_literalContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_comparisonContext extends ExpressionContext {
		public ExpressionContext left;
		public Token operator;
		public ExpressionContext right;
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode LT() { return getToken(ProcedureParser.LT, 0); }
		public TerminalNode LT_EQ() { return getToken(ProcedureParser.LT_EQ, 0); }
		public TerminalNode GT() { return getToken(ProcedureParser.GT, 0); }
		public TerminalNode GT_EQ() { return getToken(ProcedureParser.GT_EQ, 0); }
		public TerminalNode NEQ() { return getToken(ProcedureParser.NEQ, 0); }
		public TerminalNode EQ() { return getToken(ProcedureParser.EQ, 0); }
		public Expr_comparisonContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_boolean_literalContext extends ExpressionContext {
		public TerminalNode BOOLEAN_LITERAL() { return getToken(ProcedureParser.BOOLEAN_LITERAL, 0); }
		public Expr_boolean_literalContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_callContext extends ExpressionContext {
		public Call_expressionContext call_expression() {
			return getRuleContext(Call_expressionContext.class,0);
		}
		public Expr_callContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_make_arrayContext extends ExpressionContext {
		public Expression_make_arrayContext expression_make_array() {
			return getRuleContext(Expression_make_arrayContext.class,0);
		}
		public Expr_make_arrayContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_field_accessContext extends ExpressionContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode PERIOD() { return getToken(ProcedureParser.PERIOD, 0); }
		public TerminalNode IDENTIFIER() { return getToken(ProcedureParser.IDENTIFIER, 0); }
		public Expr_field_accessContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_int_literalContext extends ExpressionContext {
		public TerminalNode INT_LITERAL() { return getToken(ProcedureParser.INT_LITERAL, 0); }
		public Expr_int_literalContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_text_literalContext extends ExpressionContext {
		public TerminalNode TEXT_LITERAL() { return getToken(ProcedureParser.TEXT_LITERAL, 0); }
		public Expr_text_literalContext(ExpressionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class Expr_parenthesizedContext extends ExpressionContext {
		public TerminalNode LPAREN() { return getToken(ProcedureParser.LPAREN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(ProcedureParser.RPAREN, 0); }
		public Expr_parenthesizedContext(ExpressionContext ctx) { copyFrom(ctx); }
	}

	public final ExpressionContext expression() throws RecognitionException {
		return expression(0);
	}

	private ExpressionContext expression(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		ExpressionContext _localctx = new ExpressionContext(_ctx, _parentState);
		ExpressionContext _prevctx = _localctx;
		int _startState = 6;
		enterRecursionRule(_localctx, 6, RULE_expression, _p);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(120);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case TEXT_LITERAL:
				{
				_localctx = new Expr_text_literalContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(108);
				match(TEXT_LITERAL);
				}
				break;
			case BOOLEAN_LITERAL:
				{
				_localctx = new Expr_boolean_literalContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(109);
				match(BOOLEAN_LITERAL);
				}
				break;
			case INT_LITERAL:
				{
				_localctx = new Expr_int_literalContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(110);
				match(INT_LITERAL);
				}
				break;
			case NULL_LITERAL:
				{
				_localctx = new Expr_null_literalContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(111);
				match(NULL_LITERAL);
				}
				break;
			case BLOB_LITERAL:
				{
				_localctx = new Expr_blob_literalContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(112);
				match(BLOB_LITERAL);
				}
				break;
			case LBRACKET:
				{
				_localctx = new Expr_make_arrayContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(113);
				expression_make_array();
				}
				break;
			case IDENTIFIER:
				{
				_localctx = new Expr_callContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(114);
				call_expression();
				}
				break;
			case VARIABLE:
				{
				_localctx = new Expr_variableContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(115);
				match(VARIABLE);
				}
				break;
			case LPAREN:
				{
				_localctx = new Expr_parenthesizedContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(116);
				match(LPAREN);
				setState(117);
				expression(0);
				setState(118);
				match(RPAREN);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			_ctx.stop = _input.LT(-1);
			setState(141);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,12,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(139);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,11,_ctx) ) {
					case 1:
						{
						_localctx = new Expr_comparisonContext(new ExpressionContext(_parentctx, _parentState));
						((Expr_comparisonContext)_localctx).left = _prevctx;
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(122);
						if (!(precpred(_ctx, 3))) throw new FailedPredicateException(this, "precpred(_ctx, 3)");
						setState(123);
						((Expr_comparisonContext)_localctx).operator = _input.LT(1);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 66060288L) != 0)) ) {
							((Expr_comparisonContext)_localctx).operator = (Token)_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(124);
						((Expr_comparisonContext)_localctx).right = expression(4);
						}
						break;
					case 2:
						{
						_localctx = new Expr_arithmeticContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(125);
						if (!(precpred(_ctx, 2))) throw new FailedPredicateException(this, "precpred(_ctx, 2)");
						setState(126);
						_la = _input.LA(1);
						if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 917504L) != 0)) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(127);
						expression(3);
						}
						break;
					case 3:
						{
						_localctx = new Expr_arithmeticContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(128);
						if (!(precpred(_ctx, 1))) throw new FailedPredicateException(this, "precpred(_ctx, 1)");
						setState(129);
						_la = _input.LA(1);
						if ( !(_la==PLUS || _la==MINUS) ) {
						_errHandler.recoverInline(this);
						}
						else {
							if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
							_errHandler.reportMatch(this);
							consume();
						}
						setState(130);
						expression(2);
						}
						break;
					case 4:
						{
						_localctx = new Expr_array_accessContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(131);
						if (!(precpred(_ctx, 6))) throw new FailedPredicateException(this, "precpred(_ctx, 6)");
						setState(132);
						match(LBRACKET);
						setState(133);
						expression(0);
						setState(134);
						match(RBRACKET);
						}
						break;
					case 5:
						{
						_localctx = new Expr_field_accessContext(new ExpressionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_expression);
						setState(136);
						if (!(precpred(_ctx, 5))) throw new FailedPredicateException(this, "precpred(_ctx, 5)");
						setState(137);
						match(PERIOD);
						setState(138);
						match(IDENTIFIER);
						}
						break;
					}
					} 
				}
				setState(143);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,12,_ctx);
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
	public static class Expression_listContext extends ParserRuleContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(ProcedureParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(ProcedureParser.COMMA, i);
		}
		public Expression_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression_list; }
	}

	public final Expression_listContext expression_list() throws RecognitionException {
		Expression_listContext _localctx = new Expression_listContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_expression_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(144);
			expression(0);
			setState(149);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(145);
				match(COMMA);
				setState(146);
				expression(0);
				}
				}
				setState(151);
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
	public static class Expression_make_arrayContext extends ParserRuleContext {
		public TerminalNode LBRACKET() { return getToken(ProcedureParser.LBRACKET, 0); }
		public TerminalNode RBRACKET() { return getToken(ProcedureParser.RBRACKET, 0); }
		public Expression_listContext expression_list() {
			return getRuleContext(Expression_listContext.class,0);
		}
		public Expression_make_arrayContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression_make_array; }
	}

	public final Expression_make_arrayContext expression_make_array() throws RecognitionException {
		Expression_make_arrayContext _localctx = new Expression_make_arrayContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_expression_make_array);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(152);
			match(LBRACKET);
			setState(154);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 8727373549572L) != 0)) {
				{
				setState(153);
				expression_list();
				}
			}

			setState(156);
			match(RBRACKET);
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
	public static class Call_expressionContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(ProcedureParser.IDENTIFIER, 0); }
		public TerminalNode LPAREN() { return getToken(ProcedureParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(ProcedureParser.RPAREN, 0); }
		public Expression_listContext expression_list() {
			return getRuleContext(Expression_listContext.class,0);
		}
		public Call_expressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_expression; }
	}

	public final Call_expressionContext call_expression() throws RecognitionException {
		Call_expressionContext _localctx = new Call_expressionContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_call_expression);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(158);
			match(IDENTIFIER);
			setState(159);
			match(LPAREN);
			setState(161);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 8727373549572L) != 0)) {
				{
				setState(160);
				expression_list();
				}
			}

			setState(163);
			match(RPAREN);
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
	public static class RangeContext extends ParserRuleContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode COLON() { return getToken(ProcedureParser.COLON, 0); }
		public RangeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_range; }
	}

	public final RangeContext range() throws RecognitionException {
		RangeContext _localctx = new RangeContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_range);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(165);
			expression(0);
			setState(166);
			match(COLON);
			setState(167);
			expression(0);
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
	public static class If_then_blockContext extends ParserRuleContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode LBRACE() { return getToken(ProcedureParser.LBRACE, 0); }
		public TerminalNode RBRACE() { return getToken(ProcedureParser.RBRACE, 0); }
		public List<StatementContext> statement() {
			return getRuleContexts(StatementContext.class);
		}
		public StatementContext statement(int i) {
			return getRuleContext(StatementContext.class,i);
		}
		public If_then_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_if_then_block; }
	}

	public final If_then_blockContext if_then_block() throws RecognitionException {
		If_then_blockContext _localctx = new If_then_blockContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_if_then_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(169);
			expression(0);
			setState(170);
			match(LBRACE);
			setState(174);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 6623577767936L) != 0)) {
				{
				{
				setState(171);
				statement();
				}
				}
				setState(176);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(177);
			match(RBRACE);
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
		case 3:
			return expression_sempred((ExpressionContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean expression_sempred(ExpressionContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 3);
		case 1:
			return precpred(_ctx, 2);
		case 2:
			return precpred(_ctx, 1);
		case 3:
			return precpred(_ctx, 6);
		case 4:
			return precpred(_ctx, 5);
		}
		return true;
	}

	public static final String _serializedATN =
		"\u0004\u0001.\u00b4\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0001\u0000\u0005\u0000\u0014\b\u0000\n\u0000\f\u0000\u0017"+
		"\t\u0000\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0003\u0001-\b\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0003\u00019\b\u0001\u0001\u0001\u0001"+
		"\u0001\u0005\u0001=\b\u0001\n\u0001\f\u0001@\t\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0005\u0001G\b\u0001\n\u0001"+
		"\f\u0001J\t\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0005\u0001O\b\u0001"+
		"\n\u0001\f\u0001R\t\u0001\u0001\u0001\u0003\u0001U\b\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0003\u0001^\b\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0003\u0001e\b\u0001\u0001\u0002\u0001\u0002\u0001\u0002"+
		"\u0003\u0002j\b\u0002\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0003\u0003y\b\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0005\u0003\u008c\b\u0003"+
		"\n\u0003\f\u0003\u008f\t\u0003\u0001\u0004\u0001\u0004\u0001\u0004\u0005"+
		"\u0004\u0094\b\u0004\n\u0004\f\u0004\u0097\t\u0004\u0001\u0005\u0001\u0005"+
		"\u0003\u0005\u009b\b\u0005\u0001\u0005\u0001\u0005\u0001\u0006\u0001\u0006"+
		"\u0001\u0006\u0003\u0006\u00a2\b\u0006\u0001\u0006\u0001\u0006\u0001\u0007"+
		"\u0001\u0007\u0001\u0007\u0001\u0007\u0001\b\u0001\b\u0001\b\u0005\b\u00ad"+
		"\b\b\n\b\f\b\u00b0\t\b\u0001\b\u0001\b\u0001\b\u0000\u0001\u0006\t\u0000"+
		"\u0002\u0004\u0006\b\n\f\u000e\u0010\u0000\u0003\u0001\u0000\u0014\u0019"+
		"\u0001\u0000\u0011\u0013\u0001\u0000\u000f\u0010\u00cf\u0000\u0015\u0001"+
		"\u0000\u0000\u0000\u0002d\u0001\u0000\u0000\u0000\u0004f\u0001\u0000\u0000"+
		"\u0000\u0006x\u0001\u0000\u0000\u0000\b\u0090\u0001\u0000\u0000\u0000"+
		"\n\u0098\u0001\u0000\u0000\u0000\f\u009e\u0001\u0000\u0000\u0000\u000e"+
		"\u00a5\u0001\u0000\u0000\u0000\u0010\u00a9\u0001\u0000\u0000\u0000\u0012"+
		"\u0014\u0003\u0002\u0001\u0000\u0013\u0012\u0001\u0000\u0000\u0000\u0014"+
		"\u0017\u0001\u0000\u0000\u0000\u0015\u0013\u0001\u0000\u0000\u0000\u0015"+
		"\u0016\u0001\u0000\u0000\u0000\u0016\u0001\u0001\u0000\u0000\u0000\u0017"+
		"\u0015\u0001\u0000\u0000\u0000\u0018\u0019\u0005*\u0000\u0000\u0019\u001a"+
		"\u0003\u0004\u0002\u0000\u001a\u001b\u0005\u0001\u0000\u0000\u001be\u0001"+
		"\u0000\u0000\u0000\u001c\u001d\u0005*\u0000\u0000\u001d\u001e\u0005\n"+
		"\u0000\u0000\u001e\u001f\u0003\u0006\u0003\u0000\u001f \u0005\u0001\u0000"+
		"\u0000 e\u0001\u0000\u0000\u0000!\"\u0005*\u0000\u0000\"#\u0003\u0004"+
		"\u0002\u0000#$\u0005\n\u0000\u0000$%\u0003\u0006\u0003\u0000%&\u0005\u0001"+
		"\u0000\u0000&e\u0001\u0000\u0000\u0000\'(\u0005*\u0000\u0000()\u0005\u0006"+
		"\u0000\u0000)*\u0005*\u0000\u0000*+\u0001\u0000\u0000\u0000+-\u0005\n"+
		"\u0000\u0000,\'\u0001\u0000\u0000\u0000,-\u0001\u0000\u0000\u0000-.\u0001"+
		"\u0000\u0000\u0000./\u0003\f\u0006\u0000/0\u0005\u0001\u0000\u00000e\u0001"+
		"\u0000\u0000\u000012\u0005\u001b\u0000\u000023\u0005*\u0000\u000038\u0005"+
		"\u001c\u0000\u000049\u0003\u000e\u0007\u000059\u0003\f\u0006\u000069\u0005"+
		"*\u0000\u000079\u0005\u001a\u0000\u000084\u0001\u0000\u0000\u000085\u0001"+
		"\u0000\u0000\u000086\u0001\u0000\u0000\u000087\u0001\u0000\u0000\u0000"+
		"9:\u0001\u0000\u0000\u0000:>\u0005\u0004\u0000\u0000;=\u0003\u0002\u0001"+
		"\u0000<;\u0001\u0000\u0000\u0000=@\u0001\u0000\u0000\u0000><\u0001\u0000"+
		"\u0000\u0000>?\u0001\u0000\u0000\u0000?A\u0001\u0000\u0000\u0000@>\u0001"+
		"\u0000\u0000\u0000Ae\u0005\u0005\u0000\u0000BC\u0005\u001d\u0000\u0000"+
		"CH\u0003\u0010\b\u0000DE\u0005\u001e\u0000\u0000EG\u0003\u0010\b\u0000"+
		"FD\u0001\u0000\u0000\u0000GJ\u0001\u0000\u0000\u0000HF\u0001\u0000\u0000"+
		"\u0000HI\u0001\u0000\u0000\u0000IT\u0001\u0000\u0000\u0000JH\u0001\u0000"+
		"\u0000\u0000KL\u0005\u001f\u0000\u0000LP\u0005\u0004\u0000\u0000MO\u0003"+
		"\u0002\u0001\u0000NM\u0001\u0000\u0000\u0000OR\u0001\u0000\u0000\u0000"+
		"PN\u0001\u0000\u0000\u0000PQ\u0001\u0000\u0000\u0000QS\u0001\u0000\u0000"+
		"\u0000RP\u0001\u0000\u0000\u0000SU\u0005\u0005\u0000\u0000TK\u0001\u0000"+
		"\u0000\u0000TU\u0001\u0000\u0000\u0000Ue\u0001\u0000\u0000\u0000VW\u0005"+
		"\u001a\u0000\u0000We\u0005\u0001\u0000\u0000XY\u0005\"\u0000\u0000Ye\u0005"+
		"\u0001\u0000\u0000Z]\u0005!\u0000\u0000[^\u0003\b\u0004\u0000\\^\u0005"+
		"\u001a\u0000\u0000][\u0001\u0000\u0000\u0000]\\\u0001\u0000\u0000\u0000"+
		"^_\u0001\u0000\u0000\u0000_e\u0005\u0001\u0000\u0000`a\u0005!\u0000\u0000"+
		"ab\u0005#\u0000\u0000bc\u0005*\u0000\u0000ce\u0005\u0001\u0000\u0000d"+
		"\u0018\u0001\u0000\u0000\u0000d\u001c\u0001\u0000\u0000\u0000d!\u0001"+
		"\u0000\u0000\u0000d,\u0001\u0000\u0000\u0000d1\u0001\u0000\u0000\u0000"+
		"dB\u0001\u0000\u0000\u0000dV\u0001\u0000\u0000\u0000dX\u0001\u0000\u0000"+
		"\u0000dZ\u0001\u0000\u0000\u0000d`\u0001\u0000\u0000\u0000e\u0003\u0001"+
		"\u0000\u0000\u0000fi\u0005)\u0000\u0000gh\u0005\f\u0000\u0000hj\u0005"+
		"\r\u0000\u0000ig\u0001\u0000\u0000\u0000ij\u0001\u0000\u0000\u0000j\u0005"+
		"\u0001\u0000\u0000\u0000kl\u0006\u0003\uffff\uffff\u0000ly\u0005\'\u0000"+
		"\u0000my\u0005$\u0000\u0000ny\u0005%\u0000\u0000oy\u0005(\u0000\u0000"+
		"py\u0005&\u0000\u0000qy\u0003\n\u0005\u0000ry\u0003\f\u0006\u0000sy\u0005"+
		"*\u0000\u0000tu\u0005\u0002\u0000\u0000uv\u0003\u0006\u0003\u0000vw\u0005"+
		"\u0003\u0000\u0000wy\u0001\u0000\u0000\u0000xk\u0001\u0000\u0000\u0000"+
		"xm\u0001\u0000\u0000\u0000xn\u0001\u0000\u0000\u0000xo\u0001\u0000\u0000"+
		"\u0000xp\u0001\u0000\u0000\u0000xq\u0001\u0000\u0000\u0000xr\u0001\u0000"+
		"\u0000\u0000xs\u0001\u0000\u0000\u0000xt\u0001\u0000\u0000\u0000y\u008d"+
		"\u0001\u0000\u0000\u0000z{\n\u0003\u0000\u0000{|\u0007\u0000\u0000\u0000"+
		"|\u008c\u0003\u0006\u0003\u0004}~\n\u0002\u0000\u0000~\u007f\u0007\u0001"+
		"\u0000\u0000\u007f\u008c\u0003\u0006\u0003\u0003\u0080\u0081\n\u0001\u0000"+
		"\u0000\u0081\u0082\u0007\u0002\u0000\u0000\u0082\u008c\u0003\u0006\u0003"+
		"\u0002\u0083\u0084\n\u0006\u0000\u0000\u0084\u0085\u0005\f\u0000\u0000"+
		"\u0085\u0086\u0003\u0006\u0003\u0000\u0086\u0087\u0005\r\u0000\u0000\u0087"+
		"\u008c\u0001\u0000\u0000\u0000\u0088\u0089\n\u0005\u0000\u0000\u0089\u008a"+
		"\u0005\u000b\u0000\u0000\u008a\u008c\u0005)\u0000\u0000\u008bz\u0001\u0000"+
		"\u0000\u0000\u008b}\u0001\u0000\u0000\u0000\u008b\u0080\u0001\u0000\u0000"+
		"\u0000\u008b\u0083\u0001\u0000\u0000\u0000\u008b\u0088\u0001\u0000\u0000"+
		"\u0000\u008c\u008f\u0001\u0000\u0000\u0000\u008d\u008b\u0001\u0000\u0000"+
		"\u0000\u008d\u008e\u0001\u0000\u0000\u0000\u008e\u0007\u0001\u0000\u0000"+
		"\u0000\u008f\u008d\u0001\u0000\u0000\u0000\u0090\u0095\u0003\u0006\u0003"+
		"\u0000\u0091\u0092\u0005\u0006\u0000\u0000\u0092\u0094\u0003\u0006\u0003"+
		"\u0000\u0093\u0091\u0001\u0000\u0000\u0000\u0094\u0097\u0001\u0000\u0000"+
		"\u0000\u0095\u0093\u0001\u0000\u0000\u0000\u0095\u0096\u0001\u0000\u0000"+
		"\u0000\u0096\t\u0001\u0000\u0000\u0000\u0097\u0095\u0001\u0000\u0000\u0000"+
		"\u0098\u009a\u0005\f\u0000\u0000\u0099\u009b\u0003\b\u0004\u0000\u009a"+
		"\u0099\u0001\u0000\u0000\u0000\u009a\u009b\u0001\u0000\u0000\u0000\u009b"+
		"\u009c\u0001\u0000\u0000\u0000\u009c\u009d\u0005\r\u0000\u0000\u009d\u000b"+
		"\u0001\u0000\u0000\u0000\u009e\u009f\u0005)\u0000\u0000\u009f\u00a1\u0005"+
		"\u0002\u0000\u0000\u00a0\u00a2\u0003\b\u0004\u0000\u00a1\u00a0\u0001\u0000"+
		"\u0000\u0000\u00a1\u00a2\u0001\u0000\u0000\u0000\u00a2\u00a3\u0001\u0000"+
		"\u0000\u0000\u00a3\u00a4\u0005\u0003\u0000\u0000\u00a4\r\u0001\u0000\u0000"+
		"\u0000\u00a5\u00a6\u0003\u0006\u0003\u0000\u00a6\u00a7\u0005\u0007\u0000"+
		"\u0000\u00a7\u00a8\u0003\u0006\u0003\u0000\u00a8\u000f\u0001\u0000\u0000"+
		"\u0000\u00a9\u00aa\u0003\u0006\u0003\u0000\u00aa\u00ae\u0005\u0004\u0000"+
		"\u0000\u00ab\u00ad\u0003\u0002\u0001\u0000\u00ac\u00ab\u0001\u0000\u0000"+
		"\u0000\u00ad\u00b0\u0001\u0000\u0000\u0000\u00ae\u00ac\u0001\u0000\u0000"+
		"\u0000\u00ae\u00af\u0001\u0000\u0000\u0000\u00af\u00b1\u0001\u0000\u0000"+
		"\u0000\u00b0\u00ae\u0001\u0000\u0000\u0000\u00b1\u00b2\u0005\u0005\u0000"+
		"\u0000\u00b2\u0011\u0001\u0000\u0000\u0000\u0011\u0015,8>HPT]dix\u008b"+
		"\u008d\u0095\u009a\u00a1\u00ae";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}