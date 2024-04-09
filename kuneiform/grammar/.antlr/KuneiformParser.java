// Generated from /Users/brennanlamey/kwil-db/kuneiform/grammar/KuneiformParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class KuneiformParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		LBRACE=1, RBRACE=2, LBRACKET=3, RBRACKET=4, COL=5, SCOL=6, LPAREN=7, RPAREN=8, 
		COMMA=9, AT=10, PERIOD=11, EQUALS=12, DATABASE=13, USE=14, IMPORT=15, 
		AS=16, MIN=17, MAX=18, MIN_LEN=19, MAX_LEN=20, NOT_NULL=21, PRIMARY=22, 
		DEFAULT=23, UNIQUE=24, INDEX=25, TABLE=26, TYPE=27, FOREIGN_KEY=28, REFERENCES=29, 
		ON_UPDATE=30, ON_DELETE=31, DO_NO_ACTION=32, DO_CASCADE=33, DO_SET_NULL=34, 
		DO_SET_DEFAULT=35, DO_RESTRICT=36, DO=37, START_ACTION=38, START_PROCEDURE=39, 
		NUMERIC_LITERAL=40, TEXT_LITERAL=41, BOOLEAN_LITERAL=42, BLOB_LITERAL=43, 
		VAR=44, INDEX_NAME=45, IDENTIFIER=46, ANNOTATION=47, WS=48, TERMINATOR=49, 
		BLOCK_COMMENT=50, LINE_COMMENT=51, STMT_BODY=52, TEXT=53, STMT_LPAREN=54, 
		STMT_RPAREN=55, STMT_COMMA=56, STMT_PERIOD=57, STMT_RETURNS=58, STMT_TABLE=59, 
		STMT_ARRAY=60, STMT_VAR=61, STMT_ACCESS=62, STMT_IDENTIFIER=63, STMT_WS=64, 
		STMT_TERMINATOR=65, STMT_BLOCK_COMMENT=66, STMT_LINE_COMMENT=67;
	public static final int
		RULE_program = 0, RULE_stmt_mode = 1, RULE_database_declaration = 2, RULE_use_declaration = 3, 
		RULE_table_declaration = 4, RULE_column_def = 5, RULE_index_def = 6, RULE_foreign_key_def = 7, 
		RULE_foreign_key_action = 8, RULE_identifier_list = 9, RULE_literal = 10, 
		RULE_type_selector = 11, RULE_constraint = 12, RULE_action_declaration = 13, 
		RULE_procedure_declaration = 14, RULE_table_return = 15, RULE_stmt_typed_param_list = 16, 
		RULE_stmt_type_list = 17, RULE_stmt_type_selector = 18;
	private static String[] makeRuleNames() {
		return new String[] {
			"program", "stmt_mode", "database_declaration", "use_declaration", "table_declaration", 
			"column_def", "index_def", "foreign_key_def", "foreign_key_action", "identifier_list", 
			"literal", "type_selector", "constraint", "action_declaration", "procedure_declaration", 
			"table_return", "stmt_typed_param_list", "stmt_type_list", "stmt_type_selector"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'{'", "'}'", "'['", "']'", "':'", "';'", "'('", "')'", "','", 
			"'@'", "'.'", "'='", "'database'", "'use'", "'import'", "'as'", "'min'", 
			"'max'", "'minlen'", "'maxlen'", null, null, "'default'", "'unique'", 
			"'index'", "'table'", "'type'", null, null, null, null, null, "'cascade'", 
			null, null, "'restrict'", "'do'", "'action'", "'procedure'", null, null, 
			null, null, null, null, null, null, null, null, null, null, null, null, 
			null, null, null, null, "'returns'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "LBRACE", "RBRACE", "LBRACKET", "RBRACKET", "COL", "SCOL", "LPAREN", 
			"RPAREN", "COMMA", "AT", "PERIOD", "EQUALS", "DATABASE", "USE", "IMPORT", 
			"AS", "MIN", "MAX", "MIN_LEN", "MAX_LEN", "NOT_NULL", "PRIMARY", "DEFAULT", 
			"UNIQUE", "INDEX", "TABLE", "TYPE", "FOREIGN_KEY", "REFERENCES", "ON_UPDATE", 
			"ON_DELETE", "DO_NO_ACTION", "DO_CASCADE", "DO_SET_NULL", "DO_SET_DEFAULT", 
			"DO_RESTRICT", "DO", "START_ACTION", "START_PROCEDURE", "NUMERIC_LITERAL", 
			"TEXT_LITERAL", "BOOLEAN_LITERAL", "BLOB_LITERAL", "VAR", "INDEX_NAME", 
			"IDENTIFIER", "ANNOTATION", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT", 
			"STMT_BODY", "TEXT", "STMT_LPAREN", "STMT_RPAREN", "STMT_COMMA", "STMT_PERIOD", 
			"STMT_RETURNS", "STMT_TABLE", "STMT_ARRAY", "STMT_VAR", "STMT_ACCESS", 
			"STMT_IDENTIFIER", "STMT_WS", "STMT_TERMINATOR", "STMT_BLOCK_COMMENT", 
			"STMT_LINE_COMMENT"
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
	public String getGrammarFileName() { return "KuneiformParser.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public KuneiformParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ProgramContext extends ParserRuleContext {
		public Database_declarationContext database_declaration() {
			return getRuleContext(Database_declarationContext.class,0);
		}
		public TerminalNode EOF() { return getToken(KuneiformParser.EOF, 0); }
		public List<Use_declarationContext> use_declaration() {
			return getRuleContexts(Use_declarationContext.class);
		}
		public Use_declarationContext use_declaration(int i) {
			return getRuleContext(Use_declarationContext.class,i);
		}
		public List<Table_declarationContext> table_declaration() {
			return getRuleContexts(Table_declarationContext.class);
		}
		public Table_declarationContext table_declaration(int i) {
			return getRuleContext(Table_declarationContext.class,i);
		}
		public List<Stmt_modeContext> stmt_mode() {
			return getRuleContexts(Stmt_modeContext.class);
		}
		public Stmt_modeContext stmt_mode(int i) {
			return getRuleContext(Stmt_modeContext.class,i);
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
			setState(38);
			database_declaration();
			setState(44);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 141562189201408L) != 0)) {
				{
				setState(42);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case USE:
					{
					setState(39);
					use_declaration();
					}
					break;
				case TABLE:
					{
					setState(40);
					table_declaration();
					}
					break;
				case START_ACTION:
				case START_PROCEDURE:
				case ANNOTATION:
					{
					setState(41);
					stmt_mode();
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(46);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(47);
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
	public static class Stmt_modeContext extends ParserRuleContext {
		public Action_declarationContext action_declaration() {
			return getRuleContext(Action_declarationContext.class,0);
		}
		public Procedure_declarationContext procedure_declaration() {
			return getRuleContext(Procedure_declarationContext.class,0);
		}
		public List<TerminalNode> ANNOTATION() { return getTokens(KuneiformParser.ANNOTATION); }
		public TerminalNode ANNOTATION(int i) {
			return getToken(KuneiformParser.ANNOTATION, i);
		}
		public Stmt_modeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stmt_mode; }
	}

	public final Stmt_modeContext stmt_mode() throws RecognitionException {
		Stmt_modeContext _localctx = new Stmt_modeContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_stmt_mode);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(52);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==ANNOTATION) {
				{
				{
				setState(49);
				match(ANNOTATION);
				}
				}
				setState(54);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(57);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case START_ACTION:
				{
				setState(55);
				action_declaration();
				}
				break;
			case START_PROCEDURE:
				{
				setState(56);
				procedure_declaration();
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
	public static class Database_declarationContext extends ParserRuleContext {
		public TerminalNode DATABASE() { return getToken(KuneiformParser.DATABASE, 0); }
		public TerminalNode IDENTIFIER() { return getToken(KuneiformParser.IDENTIFIER, 0); }
		public TerminalNode SCOL() { return getToken(KuneiformParser.SCOL, 0); }
		public Database_declarationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_database_declaration; }
	}

	public final Database_declarationContext database_declaration() throws RecognitionException {
		Database_declarationContext _localctx = new Database_declarationContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_database_declaration);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(59);
			match(DATABASE);
			setState(60);
			match(IDENTIFIER);
			setState(61);
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
	public static class Use_declarationContext extends ParserRuleContext {
		public Token extension_name;
		public Token alias;
		public TerminalNode USE() { return getToken(KuneiformParser.USE, 0); }
		public TerminalNode AS() { return getToken(KuneiformParser.AS, 0); }
		public TerminalNode SCOL() { return getToken(KuneiformParser.SCOL, 0); }
		public List<TerminalNode> IDENTIFIER() { return getTokens(KuneiformParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(KuneiformParser.IDENTIFIER, i);
		}
		public TerminalNode LBRACE() { return getToken(KuneiformParser.LBRACE, 0); }
		public List<TerminalNode> COL() { return getTokens(KuneiformParser.COL); }
		public TerminalNode COL(int i) {
			return getToken(KuneiformParser.COL, i);
		}
		public List<LiteralContext> literal() {
			return getRuleContexts(LiteralContext.class);
		}
		public LiteralContext literal(int i) {
			return getRuleContext(LiteralContext.class,i);
		}
		public TerminalNode RBRACE() { return getToken(KuneiformParser.RBRACE, 0); }
		public TerminalNode COMMA() { return getToken(KuneiformParser.COMMA, 0); }
		public Use_declarationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_use_declaration; }
	}

	public final Use_declarationContext use_declaration() throws RecognitionException {
		Use_declarationContext _localctx = new Use_declarationContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_use_declaration);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(63);
			match(USE);
			setState(64);
			((Use_declarationContext)_localctx).extension_name = match(IDENTIFIER);
			setState(77);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LBRACE) {
				{
				setState(65);
				match(LBRACE);
				setState(66);
				match(IDENTIFIER);
				setState(67);
				match(COL);
				setState(68);
				literal();
				setState(73);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COMMA) {
					{
					setState(69);
					match(COMMA);
					setState(70);
					match(IDENTIFIER);
					setState(71);
					match(COL);
					setState(72);
					literal();
					}
				}

				setState(75);
				match(RBRACE);
				}
			}

			setState(79);
			match(AS);
			setState(80);
			((Use_declarationContext)_localctx).alias = match(IDENTIFIER);
			setState(81);
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
	public static class Table_declarationContext extends ParserRuleContext {
		public TerminalNode TABLE() { return getToken(KuneiformParser.TABLE, 0); }
		public TerminalNode IDENTIFIER() { return getToken(KuneiformParser.IDENTIFIER, 0); }
		public TerminalNode LBRACE() { return getToken(KuneiformParser.LBRACE, 0); }
		public List<Column_defContext> column_def() {
			return getRuleContexts(Column_defContext.class);
		}
		public Column_defContext column_def(int i) {
			return getRuleContext(Column_defContext.class,i);
		}
		public TerminalNode RBRACE() { return getToken(KuneiformParser.RBRACE, 0); }
		public List<TerminalNode> COMMA() { return getTokens(KuneiformParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(KuneiformParser.COMMA, i);
		}
		public List<Index_defContext> index_def() {
			return getRuleContexts(Index_defContext.class);
		}
		public Index_defContext index_def(int i) {
			return getRuleContext(Index_defContext.class,i);
		}
		public List<Foreign_key_defContext> foreign_key_def() {
			return getRuleContexts(Foreign_key_defContext.class);
		}
		public Foreign_key_defContext foreign_key_def(int i) {
			return getRuleContext(Foreign_key_defContext.class,i);
		}
		public Table_declarationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table_declaration; }
	}

	public final Table_declarationContext table_declaration() throws RecognitionException {
		Table_declarationContext _localctx = new Table_declarationContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_table_declaration);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(83);
			match(TABLE);
			setState(84);
			match(IDENTIFIER);
			setState(85);
			match(LBRACE);
			setState(86);
			column_def();
			setState(95);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(87);
				match(COMMA);
				setState(91);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case IDENTIFIER:
					{
					setState(88);
					column_def();
					}
					break;
				case INDEX_NAME:
					{
					setState(89);
					index_def();
					}
					break;
				case FOREIGN_KEY:
					{
					setState(90);
					foreign_key_def();
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				}
				setState(97);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(98);
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

	@SuppressWarnings("CheckReturnValue")
	public static class Column_defContext extends ParserRuleContext {
		public Token name;
		public Type_selectorContext type;
		public TerminalNode IDENTIFIER() { return getToken(KuneiformParser.IDENTIFIER, 0); }
		public Type_selectorContext type_selector() {
			return getRuleContext(Type_selectorContext.class,0);
		}
		public List<ConstraintContext> constraint() {
			return getRuleContexts(ConstraintContext.class);
		}
		public ConstraintContext constraint(int i) {
			return getRuleContext(ConstraintContext.class,i);
		}
		public Column_defContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column_def; }
	}

	public final Column_defContext column_def() throws RecognitionException {
		Column_defContext _localctx = new Column_defContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_column_def);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(100);
			((Column_defContext)_localctx).name = match(IDENTIFIER);
			setState(101);
			((Column_defContext)_localctx).type = type_selector();
			setState(105);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 33423360L) != 0)) {
				{
				{
				setState(102);
				constraint();
				}
				}
				setState(107);
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
	public static class Index_defContext extends ParserRuleContext {
		public Identifier_listContext columns;
		public TerminalNode INDEX_NAME() { return getToken(KuneiformParser.INDEX_NAME, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public TerminalNode UNIQUE() { return getToken(KuneiformParser.UNIQUE, 0); }
		public TerminalNode INDEX() { return getToken(KuneiformParser.INDEX, 0); }
		public TerminalNode PRIMARY() { return getToken(KuneiformParser.PRIMARY, 0); }
		public Identifier_listContext identifier_list() {
			return getRuleContext(Identifier_listContext.class,0);
		}
		public Index_defContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_index_def; }
	}

	public final Index_defContext index_def() throws RecognitionException {
		Index_defContext _localctx = new Index_defContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_index_def);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(108);
			match(INDEX_NAME);
			setState(109);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 54525952L) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(110);
			match(LPAREN);
			setState(111);
			((Index_defContext)_localctx).columns = identifier_list();
			setState(112);
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
	public static class Foreign_key_defContext extends ParserRuleContext {
		public Identifier_listContext child_keys;
		public Token parent_table;
		public Identifier_listContext parent_keys;
		public TerminalNode FOREIGN_KEY() { return getToken(KuneiformParser.FOREIGN_KEY, 0); }
		public List<TerminalNode> LPAREN() { return getTokens(KuneiformParser.LPAREN); }
		public TerminalNode LPAREN(int i) {
			return getToken(KuneiformParser.LPAREN, i);
		}
		public List<TerminalNode> RPAREN() { return getTokens(KuneiformParser.RPAREN); }
		public TerminalNode RPAREN(int i) {
			return getToken(KuneiformParser.RPAREN, i);
		}
		public TerminalNode REFERENCES() { return getToken(KuneiformParser.REFERENCES, 0); }
		public List<Identifier_listContext> identifier_list() {
			return getRuleContexts(Identifier_listContext.class);
		}
		public Identifier_listContext identifier_list(int i) {
			return getRuleContext(Identifier_listContext.class,i);
		}
		public TerminalNode IDENTIFIER() { return getToken(KuneiformParser.IDENTIFIER, 0); }
		public List<Foreign_key_actionContext> foreign_key_action() {
			return getRuleContexts(Foreign_key_actionContext.class);
		}
		public Foreign_key_actionContext foreign_key_action(int i) {
			return getRuleContext(Foreign_key_actionContext.class,i);
		}
		public Foreign_key_defContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_foreign_key_def; }
	}

	public final Foreign_key_defContext foreign_key_def() throws RecognitionException {
		Foreign_key_defContext _localctx = new Foreign_key_defContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_foreign_key_def);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(114);
			match(FOREIGN_KEY);
			setState(115);
			match(LPAREN);
			setState(116);
			((Foreign_key_defContext)_localctx).child_keys = identifier_list();
			setState(117);
			match(RPAREN);
			setState(118);
			match(REFERENCES);
			setState(119);
			((Foreign_key_defContext)_localctx).parent_table = match(IDENTIFIER);
			setState(120);
			match(LPAREN);
			setState(121);
			((Foreign_key_defContext)_localctx).parent_keys = identifier_list();
			setState(122);
			match(RPAREN);
			setState(126);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==ON_UPDATE || _la==ON_DELETE) {
				{
				{
				setState(123);
				foreign_key_action();
				}
				}
				setState(128);
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
	public static class Foreign_key_actionContext extends ParserRuleContext {
		public TerminalNode ON_UPDATE() { return getToken(KuneiformParser.ON_UPDATE, 0); }
		public TerminalNode ON_DELETE() { return getToken(KuneiformParser.ON_DELETE, 0); }
		public TerminalNode DO_NO_ACTION() { return getToken(KuneiformParser.DO_NO_ACTION, 0); }
		public TerminalNode DO_CASCADE() { return getToken(KuneiformParser.DO_CASCADE, 0); }
		public TerminalNode DO_SET_NULL() { return getToken(KuneiformParser.DO_SET_NULL, 0); }
		public TerminalNode DO_SET_DEFAULT() { return getToken(KuneiformParser.DO_SET_DEFAULT, 0); }
		public TerminalNode DO_RESTRICT() { return getToken(KuneiformParser.DO_RESTRICT, 0); }
		public TerminalNode DO() { return getToken(KuneiformParser.DO, 0); }
		public Foreign_key_actionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_foreign_key_action; }
	}

	public final Foreign_key_actionContext foreign_key_action() throws RecognitionException {
		Foreign_key_actionContext _localctx = new Foreign_key_actionContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_foreign_key_action);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(129);
			_la = _input.LA(1);
			if ( !(_la==ON_UPDATE || _la==ON_DELETE) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(131);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==DO) {
				{
				setState(130);
				match(DO);
				}
			}

			setState(133);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 133143986176L) != 0)) ) {
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
	public static class Identifier_listContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(KuneiformParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(KuneiformParser.IDENTIFIER, i);
		}
		public List<TerminalNode> COMMA() { return getTokens(KuneiformParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(KuneiformParser.COMMA, i);
		}
		public Identifier_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_identifier_list; }
	}

	public final Identifier_listContext identifier_list() throws RecognitionException {
		Identifier_listContext _localctx = new Identifier_listContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_identifier_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(135);
			match(IDENTIFIER);
			setState(140);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(136);
				match(COMMA);
				setState(137);
				match(IDENTIFIER);
				}
				}
				setState(142);
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
	public static class LiteralContext extends ParserRuleContext {
		public TerminalNode NUMERIC_LITERAL() { return getToken(KuneiformParser.NUMERIC_LITERAL, 0); }
		public TerminalNode BLOB_LITERAL() { return getToken(KuneiformParser.BLOB_LITERAL, 0); }
		public TerminalNode TEXT_LITERAL() { return getToken(KuneiformParser.TEXT_LITERAL, 0); }
		public TerminalNode BOOLEAN_LITERAL() { return getToken(KuneiformParser.BOOLEAN_LITERAL, 0); }
		public LiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_literal; }
	}

	public final LiteralContext literal() throws RecognitionException {
		LiteralContext _localctx = new LiteralContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_literal);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(143);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16492674416640L) != 0)) ) {
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
	public static class Type_selectorContext extends ParserRuleContext {
		public Token type;
		public TerminalNode IDENTIFIER() { return getToken(KuneiformParser.IDENTIFIER, 0); }
		public TerminalNode LBRACKET() { return getToken(KuneiformParser.LBRACKET, 0); }
		public TerminalNode RBRACKET() { return getToken(KuneiformParser.RBRACKET, 0); }
		public Type_selectorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_type_selector; }
	}

	public final Type_selectorContext type_selector() throws RecognitionException {
		Type_selectorContext _localctx = new Type_selectorContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_type_selector);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(145);
			((Type_selectorContext)_localctx).type = match(IDENTIFIER);
			setState(148);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LBRACKET) {
				{
				setState(146);
				match(LBRACKET);
				setState(147);
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
	public static class ConstraintContext extends ParserRuleContext {
		public ConstraintContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_constraint; }
	 
		public ConstraintContext() { }
		public void copyFrom(ConstraintContext ctx) {
			super.copyFrom(ctx);
		}
	}
	@SuppressWarnings("CheckReturnValue")
	public static class MIN_LENContext extends ConstraintContext {
		public TerminalNode MIN_LEN() { return getToken(KuneiformParser.MIN_LEN, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public TerminalNode NUMERIC_LITERAL() { return getToken(KuneiformParser.NUMERIC_LITERAL, 0); }
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public MIN_LENContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class MINContext extends ConstraintContext {
		public TerminalNode MIN() { return getToken(KuneiformParser.MIN, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public TerminalNode NUMERIC_LITERAL() { return getToken(KuneiformParser.NUMERIC_LITERAL, 0); }
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public MINContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class PRIMARY_KEYContext extends ConstraintContext {
		public TerminalNode PRIMARY() { return getToken(KuneiformParser.PRIMARY, 0); }
		public PRIMARY_KEYContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class MAXContext extends ConstraintContext {
		public TerminalNode MAX() { return getToken(KuneiformParser.MAX, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public TerminalNode NUMERIC_LITERAL() { return getToken(KuneiformParser.NUMERIC_LITERAL, 0); }
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public MAXContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class MAX_LENContext extends ConstraintContext {
		public TerminalNode MAX_LEN() { return getToken(KuneiformParser.MAX_LEN, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public TerminalNode NUMERIC_LITERAL() { return getToken(KuneiformParser.NUMERIC_LITERAL, 0); }
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public MAX_LENContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class UNIQUEContext extends ConstraintContext {
		public TerminalNode UNIQUE() { return getToken(KuneiformParser.UNIQUE, 0); }
		public UNIQUEContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class NOT_NULLContext extends ConstraintContext {
		public TerminalNode NOT_NULL() { return getToken(KuneiformParser.NOT_NULL, 0); }
		public NOT_NULLContext(ConstraintContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class DEFAULTContext extends ConstraintContext {
		public TerminalNode DEFAULT() { return getToken(KuneiformParser.DEFAULT, 0); }
		public TerminalNode LPAREN() { return getToken(KuneiformParser.LPAREN, 0); }
		public LiteralContext literal() {
			return getRuleContext(LiteralContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(KuneiformParser.RPAREN, 0); }
		public DEFAULTContext(ConstraintContext ctx) { copyFrom(ctx); }
	}

	public final ConstraintContext constraint() throws RecognitionException {
		ConstraintContext _localctx = new ConstraintContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_constraint);
		try {
			setState(174);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case MIN:
				_localctx = new MINContext(_localctx);
				enterOuterAlt(_localctx, 1);
				{
				setState(150);
				match(MIN);
				setState(151);
				match(LPAREN);
				setState(152);
				match(NUMERIC_LITERAL);
				setState(153);
				match(RPAREN);
				}
				break;
			case MAX:
				_localctx = new MAXContext(_localctx);
				enterOuterAlt(_localctx, 2);
				{
				setState(154);
				match(MAX);
				setState(155);
				match(LPAREN);
				setState(156);
				match(NUMERIC_LITERAL);
				setState(157);
				match(RPAREN);
				}
				break;
			case MIN_LEN:
				_localctx = new MIN_LENContext(_localctx);
				enterOuterAlt(_localctx, 3);
				{
				setState(158);
				match(MIN_LEN);
				setState(159);
				match(LPAREN);
				setState(160);
				match(NUMERIC_LITERAL);
				setState(161);
				match(RPAREN);
				}
				break;
			case MAX_LEN:
				_localctx = new MAX_LENContext(_localctx);
				enterOuterAlt(_localctx, 4);
				{
				setState(162);
				match(MAX_LEN);
				setState(163);
				match(LPAREN);
				setState(164);
				match(NUMERIC_LITERAL);
				setState(165);
				match(RPAREN);
				}
				break;
			case NOT_NULL:
				_localctx = new NOT_NULLContext(_localctx);
				enterOuterAlt(_localctx, 5);
				{
				setState(166);
				match(NOT_NULL);
				}
				break;
			case PRIMARY:
				_localctx = new PRIMARY_KEYContext(_localctx);
				enterOuterAlt(_localctx, 6);
				{
				setState(167);
				match(PRIMARY);
				}
				break;
			case DEFAULT:
				_localctx = new DEFAULTContext(_localctx);
				enterOuterAlt(_localctx, 7);
				{
				setState(168);
				match(DEFAULT);
				setState(169);
				match(LPAREN);
				setState(170);
				literal();
				setState(171);
				match(RPAREN);
				}
				break;
			case UNIQUE:
				_localctx = new UNIQUEContext(_localctx);
				enterOuterAlt(_localctx, 8);
				{
				setState(173);
				match(UNIQUE);
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
	public static class Action_declarationContext extends ParserRuleContext {
		public TerminalNode START_ACTION() { return getToken(KuneiformParser.START_ACTION, 0); }
		public TerminalNode STMT_IDENTIFIER() { return getToken(KuneiformParser.STMT_IDENTIFIER, 0); }
		public TerminalNode STMT_LPAREN() { return getToken(KuneiformParser.STMT_LPAREN, 0); }
		public TerminalNode STMT_RPAREN() { return getToken(KuneiformParser.STMT_RPAREN, 0); }
		public TerminalNode STMT_BODY() { return getToken(KuneiformParser.STMT_BODY, 0); }
		public List<TerminalNode> STMT_VAR() { return getTokens(KuneiformParser.STMT_VAR); }
		public TerminalNode STMT_VAR(int i) {
			return getToken(KuneiformParser.STMT_VAR, i);
		}
		public List<TerminalNode> STMT_ACCESS() { return getTokens(KuneiformParser.STMT_ACCESS); }
		public TerminalNode STMT_ACCESS(int i) {
			return getToken(KuneiformParser.STMT_ACCESS, i);
		}
		public List<TerminalNode> STMT_COMMA() { return getTokens(KuneiformParser.STMT_COMMA); }
		public TerminalNode STMT_COMMA(int i) {
			return getToken(KuneiformParser.STMT_COMMA, i);
		}
		public Action_declarationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_action_declaration; }
	}

	public final Action_declarationContext action_declaration() throws RecognitionException {
		Action_declarationContext _localctx = new Action_declarationContext(_ctx, getState());
		enterRule(_localctx, 26, RULE_action_declaration);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(176);
			match(START_ACTION);
			setState(177);
			match(STMT_IDENTIFIER);
			setState(178);
			match(STMT_LPAREN);
			setState(187);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STMT_VAR) {
				{
				setState(179);
				match(STMT_VAR);
				setState(184);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==STMT_COMMA) {
					{
					{
					setState(180);
					match(STMT_COMMA);
					setState(181);
					match(STMT_VAR);
					}
					}
					setState(186);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
			}

			setState(189);
			match(STMT_RPAREN);
			setState(191); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(190);
				match(STMT_ACCESS);
				}
				}
				setState(193); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==STMT_ACCESS );
			setState(195);
			match(STMT_BODY);
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
	public static class Procedure_declarationContext extends ParserRuleContext {
		public Token procedure_name;
		public TerminalNode START_PROCEDURE() { return getToken(KuneiformParser.START_PROCEDURE, 0); }
		public TerminalNode STMT_LPAREN() { return getToken(KuneiformParser.STMT_LPAREN, 0); }
		public TerminalNode STMT_RPAREN() { return getToken(KuneiformParser.STMT_RPAREN, 0); }
		public TerminalNode STMT_BODY() { return getToken(KuneiformParser.STMT_BODY, 0); }
		public TerminalNode STMT_IDENTIFIER() { return getToken(KuneiformParser.STMT_IDENTIFIER, 0); }
		public Stmt_typed_param_listContext stmt_typed_param_list() {
			return getRuleContext(Stmt_typed_param_listContext.class,0);
		}
		public List<TerminalNode> STMT_ACCESS() { return getTokens(KuneiformParser.STMT_ACCESS); }
		public TerminalNode STMT_ACCESS(int i) {
			return getToken(KuneiformParser.STMT_ACCESS, i);
		}
		public TerminalNode STMT_RETURNS() { return getToken(KuneiformParser.STMT_RETURNS, 0); }
		public Stmt_type_listContext stmt_type_list() {
			return getRuleContext(Stmt_type_listContext.class,0);
		}
		public Table_returnContext table_return() {
			return getRuleContext(Table_returnContext.class,0);
		}
		public Procedure_declarationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_procedure_declaration; }
	}

	public final Procedure_declarationContext procedure_declaration() throws RecognitionException {
		Procedure_declarationContext _localctx = new Procedure_declarationContext(_ctx, getState());
		enterRule(_localctx, 28, RULE_procedure_declaration);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(197);
			match(START_PROCEDURE);
			setState(198);
			((Procedure_declarationContext)_localctx).procedure_name = match(STMT_IDENTIFIER);
			setState(199);
			match(STMT_LPAREN);
			setState(201);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STMT_VAR) {
				{
				setState(200);
				stmt_typed_param_list();
				}
			}

			setState(203);
			match(STMT_RPAREN);
			setState(205); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(204);
				match(STMT_ACCESS);
				}
				}
				setState(207); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==STMT_ACCESS );
			setState(214);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STMT_RETURNS) {
				{
				setState(209);
				match(STMT_RETURNS);
				setState(212);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case STMT_LPAREN:
					{
					setState(210);
					stmt_type_list();
					}
					break;
				case STMT_TABLE:
					{
					setState(211);
					table_return();
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
			}

			setState(216);
			match(STMT_BODY);
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
	public static class Table_returnContext extends ParserRuleContext {
		public TerminalNode STMT_TABLE() { return getToken(KuneiformParser.STMT_TABLE, 0); }
		public TerminalNode STMT_LPAREN() { return getToken(KuneiformParser.STMT_LPAREN, 0); }
		public List<TerminalNode> STMT_IDENTIFIER() { return getTokens(KuneiformParser.STMT_IDENTIFIER); }
		public TerminalNode STMT_IDENTIFIER(int i) {
			return getToken(KuneiformParser.STMT_IDENTIFIER, i);
		}
		public List<Stmt_type_selectorContext> stmt_type_selector() {
			return getRuleContexts(Stmt_type_selectorContext.class);
		}
		public Stmt_type_selectorContext stmt_type_selector(int i) {
			return getRuleContext(Stmt_type_selectorContext.class,i);
		}
		public TerminalNode STMT_RPAREN() { return getToken(KuneiformParser.STMT_RPAREN, 0); }
		public List<TerminalNode> STMT_COMMA() { return getTokens(KuneiformParser.STMT_COMMA); }
		public TerminalNode STMT_COMMA(int i) {
			return getToken(KuneiformParser.STMT_COMMA, i);
		}
		public Table_returnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table_return; }
	}

	public final Table_returnContext table_return() throws RecognitionException {
		Table_returnContext _localctx = new Table_returnContext(_ctx, getState());
		enterRule(_localctx, 30, RULE_table_return);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(218);
			match(STMT_TABLE);
			setState(219);
			match(STMT_LPAREN);
			setState(220);
			match(STMT_IDENTIFIER);
			setState(221);
			stmt_type_selector();
			setState(227);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==STMT_COMMA) {
				{
				{
				setState(222);
				match(STMT_COMMA);
				setState(223);
				match(STMT_IDENTIFIER);
				setState(224);
				stmt_type_selector();
				}
				}
				setState(229);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(230);
			match(STMT_RPAREN);
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
	public static class Stmt_typed_param_listContext extends ParserRuleContext {
		public List<TerminalNode> STMT_VAR() { return getTokens(KuneiformParser.STMT_VAR); }
		public TerminalNode STMT_VAR(int i) {
			return getToken(KuneiformParser.STMT_VAR, i);
		}
		public List<Stmt_type_selectorContext> stmt_type_selector() {
			return getRuleContexts(Stmt_type_selectorContext.class);
		}
		public Stmt_type_selectorContext stmt_type_selector(int i) {
			return getRuleContext(Stmt_type_selectorContext.class,i);
		}
		public List<TerminalNode> STMT_COMMA() { return getTokens(KuneiformParser.STMT_COMMA); }
		public TerminalNode STMT_COMMA(int i) {
			return getToken(KuneiformParser.STMT_COMMA, i);
		}
		public Stmt_typed_param_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stmt_typed_param_list; }
	}

	public final Stmt_typed_param_listContext stmt_typed_param_list() throws RecognitionException {
		Stmt_typed_param_listContext _localctx = new Stmt_typed_param_listContext(_ctx, getState());
		enterRule(_localctx, 32, RULE_stmt_typed_param_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(232);
			match(STMT_VAR);
			setState(233);
			stmt_type_selector();
			setState(239);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==STMT_COMMA) {
				{
				{
				setState(234);
				match(STMT_COMMA);
				setState(235);
				match(STMT_VAR);
				setState(236);
				stmt_type_selector();
				}
				}
				setState(241);
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
	public static class Stmt_type_listContext extends ParserRuleContext {
		public TerminalNode STMT_LPAREN() { return getToken(KuneiformParser.STMT_LPAREN, 0); }
		public List<Stmt_type_selectorContext> stmt_type_selector() {
			return getRuleContexts(Stmt_type_selectorContext.class);
		}
		public Stmt_type_selectorContext stmt_type_selector(int i) {
			return getRuleContext(Stmt_type_selectorContext.class,i);
		}
		public TerminalNode STMT_RPAREN() { return getToken(KuneiformParser.STMT_RPAREN, 0); }
		public List<TerminalNode> STMT_COMMA() { return getTokens(KuneiformParser.STMT_COMMA); }
		public TerminalNode STMT_COMMA(int i) {
			return getToken(KuneiformParser.STMT_COMMA, i);
		}
		public Stmt_type_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stmt_type_list; }
	}

	public final Stmt_type_listContext stmt_type_list() throws RecognitionException {
		Stmt_type_listContext _localctx = new Stmt_type_listContext(_ctx, getState());
		enterRule(_localctx, 34, RULE_stmt_type_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(242);
			match(STMT_LPAREN);
			setState(243);
			stmt_type_selector();
			setState(248);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==STMT_COMMA) {
				{
				{
				setState(244);
				match(STMT_COMMA);
				setState(245);
				stmt_type_selector();
				}
				}
				setState(250);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(251);
			match(STMT_RPAREN);
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
	public static class Stmt_type_selectorContext extends ParserRuleContext {
		public Token type;
		public TerminalNode STMT_IDENTIFIER() { return getToken(KuneiformParser.STMT_IDENTIFIER, 0); }
		public TerminalNode STMT_ARRAY() { return getToken(KuneiformParser.STMT_ARRAY, 0); }
		public Stmt_type_selectorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stmt_type_selector; }
	}

	public final Stmt_type_selectorContext stmt_type_selector() throws RecognitionException {
		Stmt_type_selectorContext _localctx = new Stmt_type_selectorContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_stmt_type_selector);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(253);
			((Stmt_type_selectorContext)_localctx).type = match(STMT_IDENTIFIER);
			setState(255);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STMT_ARRAY) {
				{
				setState(254);
				match(STMT_ARRAY);
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

	public static final String _serializedATN =
		"\u0004\u0001C\u0102\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007\u000f"+
		"\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007\u0012"+
		"\u0001\u0000\u0001\u0000\u0001\u0000\u0001\u0000\u0005\u0000+\b\u0000"+
		"\n\u0000\f\u0000.\t\u0000\u0001\u0000\u0001\u0000\u0001\u0001\u0005\u0001"+
		"3\b\u0001\n\u0001\f\u00016\t\u0001\u0001\u0001\u0001\u0001\u0003\u0001"+
		":\b\u0001\u0001\u0002\u0001\u0002\u0001\u0002\u0001\u0002\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0003\u0003J\b\u0003\u0001\u0003"+
		"\u0001\u0003\u0003\u0003N\b\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0003\u0004\\\b\u0004\u0005\u0004"+
		"^\b\u0004\n\u0004\f\u0004a\t\u0004\u0001\u0004\u0001\u0004\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0005\u0005h\b\u0005\n\u0005\f\u0005k\t\u0005"+
		"\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0006"+
		"\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007"+
		"\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0005\u0007}\b\u0007"+
		"\n\u0007\f\u0007\u0080\t\u0007\u0001\b\u0001\b\u0003\b\u0084\b\b\u0001"+
		"\b\u0001\b\u0001\t\u0001\t\u0001\t\u0005\t\u008b\b\t\n\t\f\t\u008e\t\t"+
		"\u0001\n\u0001\n\u0001\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u0095"+
		"\b\u000b\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001"+
		"\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001"+
		"\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0003\f\u00af"+
		"\b\f\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0005\r\u00b7\b\r"+
		"\n\r\f\r\u00ba\t\r\u0003\r\u00bc\b\r\u0001\r\u0001\r\u0004\r\u00c0\b\r"+
		"\u000b\r\f\r\u00c1\u0001\r\u0001\r\u0001\u000e\u0001\u000e\u0001\u000e"+
		"\u0001\u000e\u0003\u000e\u00ca\b\u000e\u0001\u000e\u0001\u000e\u0004\u000e"+
		"\u00ce\b\u000e\u000b\u000e\f\u000e\u00cf\u0001\u000e\u0001\u000e\u0001"+
		"\u000e\u0003\u000e\u00d5\b\u000e\u0003\u000e\u00d7\b\u000e\u0001\u000e"+
		"\u0001\u000e\u0001\u000f\u0001\u000f\u0001\u000f\u0001\u000f\u0001\u000f"+
		"\u0001\u000f\u0001\u000f\u0005\u000f\u00e2\b\u000f\n\u000f\f\u000f\u00e5"+
		"\t\u000f\u0001\u000f\u0001\u000f\u0001\u0010\u0001\u0010\u0001\u0010\u0001"+
		"\u0010\u0001\u0010\u0005\u0010\u00ee\b\u0010\n\u0010\f\u0010\u00f1\t\u0010"+
		"\u0001\u0011\u0001\u0011\u0001\u0011\u0001\u0011\u0005\u0011\u00f7\b\u0011"+
		"\n\u0011\f\u0011\u00fa\t\u0011\u0001\u0011\u0001\u0011\u0001\u0012\u0001"+
		"\u0012\u0003\u0012\u0100\b\u0012\u0001\u0012\u0000\u0000\u0013\u0000\u0002"+
		"\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a\u001c\u001e"+
		" \"$\u0000\u0004\u0002\u0000\u0016\u0016\u0018\u0019\u0001\u0000\u001e"+
		"\u001f\u0001\u0000 $\u0001\u0000(+\u010f\u0000&\u0001\u0000\u0000\u0000"+
		"\u00024\u0001\u0000\u0000\u0000\u0004;\u0001\u0000\u0000\u0000\u0006?"+
		"\u0001\u0000\u0000\u0000\bS\u0001\u0000\u0000\u0000\nd\u0001\u0000\u0000"+
		"\u0000\fl\u0001\u0000\u0000\u0000\u000er\u0001\u0000\u0000\u0000\u0010"+
		"\u0081\u0001\u0000\u0000\u0000\u0012\u0087\u0001\u0000\u0000\u0000\u0014"+
		"\u008f\u0001\u0000\u0000\u0000\u0016\u0091\u0001\u0000\u0000\u0000\u0018"+
		"\u00ae\u0001\u0000\u0000\u0000\u001a\u00b0\u0001\u0000\u0000\u0000\u001c"+
		"\u00c5\u0001\u0000\u0000\u0000\u001e\u00da\u0001\u0000\u0000\u0000 \u00e8"+
		"\u0001\u0000\u0000\u0000\"\u00f2\u0001\u0000\u0000\u0000$\u00fd\u0001"+
		"\u0000\u0000\u0000&,\u0003\u0004\u0002\u0000\'+\u0003\u0006\u0003\u0000"+
		"(+\u0003\b\u0004\u0000)+\u0003\u0002\u0001\u0000*\'\u0001\u0000\u0000"+
		"\u0000*(\u0001\u0000\u0000\u0000*)\u0001\u0000\u0000\u0000+.\u0001\u0000"+
		"\u0000\u0000,*\u0001\u0000\u0000\u0000,-\u0001\u0000\u0000\u0000-/\u0001"+
		"\u0000\u0000\u0000.,\u0001\u0000\u0000\u0000/0\u0005\u0000\u0000\u0001"+
		"0\u0001\u0001\u0000\u0000\u000013\u0005/\u0000\u000021\u0001\u0000\u0000"+
		"\u000036\u0001\u0000\u0000\u000042\u0001\u0000\u0000\u000045\u0001\u0000"+
		"\u0000\u000059\u0001\u0000\u0000\u000064\u0001\u0000\u0000\u00007:\u0003"+
		"\u001a\r\u00008:\u0003\u001c\u000e\u000097\u0001\u0000\u0000\u000098\u0001"+
		"\u0000\u0000\u0000:\u0003\u0001\u0000\u0000\u0000;<\u0005\r\u0000\u0000"+
		"<=\u0005.\u0000\u0000=>\u0005\u0006\u0000\u0000>\u0005\u0001\u0000\u0000"+
		"\u0000?@\u0005\u000e\u0000\u0000@M\u0005.\u0000\u0000AB\u0005\u0001\u0000"+
		"\u0000BC\u0005.\u0000\u0000CD\u0005\u0005\u0000\u0000DI\u0003\u0014\n"+
		"\u0000EF\u0005\t\u0000\u0000FG\u0005.\u0000\u0000GH\u0005\u0005\u0000"+
		"\u0000HJ\u0003\u0014\n\u0000IE\u0001\u0000\u0000\u0000IJ\u0001\u0000\u0000"+
		"\u0000JK\u0001\u0000\u0000\u0000KL\u0005\u0002\u0000\u0000LN\u0001\u0000"+
		"\u0000\u0000MA\u0001\u0000\u0000\u0000MN\u0001\u0000\u0000\u0000NO\u0001"+
		"\u0000\u0000\u0000OP\u0005\u0010\u0000\u0000PQ\u0005.\u0000\u0000QR\u0005"+
		"\u0006\u0000\u0000R\u0007\u0001\u0000\u0000\u0000ST\u0005\u001a\u0000"+
		"\u0000TU\u0005.\u0000\u0000UV\u0005\u0001\u0000\u0000V_\u0003\n\u0005"+
		"\u0000W[\u0005\t\u0000\u0000X\\\u0003\n\u0005\u0000Y\\\u0003\f\u0006\u0000"+
		"Z\\\u0003\u000e\u0007\u0000[X\u0001\u0000\u0000\u0000[Y\u0001\u0000\u0000"+
		"\u0000[Z\u0001\u0000\u0000\u0000\\^\u0001\u0000\u0000\u0000]W\u0001\u0000"+
		"\u0000\u0000^a\u0001\u0000\u0000\u0000_]\u0001\u0000\u0000\u0000_`\u0001"+
		"\u0000\u0000\u0000`b\u0001\u0000\u0000\u0000a_\u0001\u0000\u0000\u0000"+
		"bc\u0005\u0002\u0000\u0000c\t\u0001\u0000\u0000\u0000de\u0005.\u0000\u0000"+
		"ei\u0003\u0016\u000b\u0000fh\u0003\u0018\f\u0000gf\u0001\u0000\u0000\u0000"+
		"hk\u0001\u0000\u0000\u0000ig\u0001\u0000\u0000\u0000ij\u0001\u0000\u0000"+
		"\u0000j\u000b\u0001\u0000\u0000\u0000ki\u0001\u0000\u0000\u0000lm\u0005"+
		"-\u0000\u0000mn\u0007\u0000\u0000\u0000no\u0005\u0007\u0000\u0000op\u0003"+
		"\u0012\t\u0000pq\u0005\b\u0000\u0000q\r\u0001\u0000\u0000\u0000rs\u0005"+
		"\u001c\u0000\u0000st\u0005\u0007\u0000\u0000tu\u0003\u0012\t\u0000uv\u0005"+
		"\b\u0000\u0000vw\u0005\u001d\u0000\u0000wx\u0005.\u0000\u0000xy\u0005"+
		"\u0007\u0000\u0000yz\u0003\u0012\t\u0000z~\u0005\b\u0000\u0000{}\u0003"+
		"\u0010\b\u0000|{\u0001\u0000\u0000\u0000}\u0080\u0001\u0000\u0000\u0000"+
		"~|\u0001\u0000\u0000\u0000~\u007f\u0001\u0000\u0000\u0000\u007f\u000f"+
		"\u0001\u0000\u0000\u0000\u0080~\u0001\u0000\u0000\u0000\u0081\u0083\u0007"+
		"\u0001\u0000\u0000\u0082\u0084\u0005%\u0000\u0000\u0083\u0082\u0001\u0000"+
		"\u0000\u0000\u0083\u0084\u0001\u0000\u0000\u0000\u0084\u0085\u0001\u0000"+
		"\u0000\u0000\u0085\u0086\u0007\u0002\u0000\u0000\u0086\u0011\u0001\u0000"+
		"\u0000\u0000\u0087\u008c\u0005.\u0000\u0000\u0088\u0089\u0005\t\u0000"+
		"\u0000\u0089\u008b\u0005.\u0000\u0000\u008a\u0088\u0001\u0000\u0000\u0000"+
		"\u008b\u008e\u0001\u0000\u0000\u0000\u008c\u008a\u0001\u0000\u0000\u0000"+
		"\u008c\u008d\u0001\u0000\u0000\u0000\u008d\u0013\u0001\u0000\u0000\u0000"+
		"\u008e\u008c\u0001\u0000\u0000\u0000\u008f\u0090\u0007\u0003\u0000\u0000"+
		"\u0090\u0015\u0001\u0000\u0000\u0000\u0091\u0094\u0005.\u0000\u0000\u0092"+
		"\u0093\u0005\u0003\u0000\u0000\u0093\u0095\u0005\u0004\u0000\u0000\u0094"+
		"\u0092\u0001\u0000\u0000\u0000\u0094\u0095\u0001\u0000\u0000\u0000\u0095"+
		"\u0017\u0001\u0000\u0000\u0000\u0096\u0097\u0005\u0011\u0000\u0000\u0097"+
		"\u0098\u0005\u0007\u0000\u0000\u0098\u0099\u0005(\u0000\u0000\u0099\u00af"+
		"\u0005\b\u0000\u0000\u009a\u009b\u0005\u0012\u0000\u0000\u009b\u009c\u0005"+
		"\u0007\u0000\u0000\u009c\u009d\u0005(\u0000\u0000\u009d\u00af\u0005\b"+
		"\u0000\u0000\u009e\u009f\u0005\u0013\u0000\u0000\u009f\u00a0\u0005\u0007"+
		"\u0000\u0000\u00a0\u00a1\u0005(\u0000\u0000\u00a1\u00af\u0005\b\u0000"+
		"\u0000\u00a2\u00a3\u0005\u0014\u0000\u0000\u00a3\u00a4\u0005\u0007\u0000"+
		"\u0000\u00a4\u00a5\u0005(\u0000\u0000\u00a5\u00af\u0005\b\u0000\u0000"+
		"\u00a6\u00af\u0005\u0015\u0000\u0000\u00a7\u00af\u0005\u0016\u0000\u0000"+
		"\u00a8\u00a9\u0005\u0017\u0000\u0000\u00a9\u00aa\u0005\u0007\u0000\u0000"+
		"\u00aa\u00ab\u0003\u0014\n\u0000\u00ab\u00ac\u0005\b\u0000\u0000\u00ac"+
		"\u00af\u0001\u0000\u0000\u0000\u00ad\u00af\u0005\u0018\u0000\u0000\u00ae"+
		"\u0096\u0001\u0000\u0000\u0000\u00ae\u009a\u0001\u0000\u0000\u0000\u00ae"+
		"\u009e\u0001\u0000\u0000\u0000\u00ae\u00a2\u0001\u0000\u0000\u0000\u00ae"+
		"\u00a6\u0001\u0000\u0000\u0000\u00ae\u00a7\u0001\u0000\u0000\u0000\u00ae"+
		"\u00a8\u0001\u0000\u0000\u0000\u00ae\u00ad\u0001\u0000\u0000\u0000\u00af"+
		"\u0019\u0001\u0000\u0000\u0000\u00b0\u00b1\u0005&\u0000\u0000\u00b1\u00b2"+
		"\u0005?\u0000\u0000\u00b2\u00bb\u00056\u0000\u0000\u00b3\u00b8\u0005="+
		"\u0000\u0000\u00b4\u00b5\u00058\u0000\u0000\u00b5\u00b7\u0005=\u0000\u0000"+
		"\u00b6\u00b4\u0001\u0000\u0000\u0000\u00b7\u00ba\u0001\u0000\u0000\u0000"+
		"\u00b8\u00b6\u0001\u0000\u0000\u0000\u00b8\u00b9\u0001\u0000\u0000\u0000"+
		"\u00b9\u00bc\u0001\u0000\u0000\u0000\u00ba\u00b8\u0001\u0000\u0000\u0000"+
		"\u00bb\u00b3\u0001\u0000\u0000\u0000\u00bb\u00bc\u0001\u0000\u0000\u0000"+
		"\u00bc\u00bd\u0001\u0000\u0000\u0000\u00bd\u00bf\u00057\u0000\u0000\u00be"+
		"\u00c0\u0005>\u0000\u0000\u00bf\u00be\u0001\u0000\u0000\u0000\u00c0\u00c1"+
		"\u0001\u0000\u0000\u0000\u00c1\u00bf\u0001\u0000\u0000\u0000\u00c1\u00c2"+
		"\u0001\u0000\u0000\u0000\u00c2\u00c3\u0001\u0000\u0000\u0000\u00c3\u00c4"+
		"\u00054\u0000\u0000\u00c4\u001b\u0001\u0000\u0000\u0000\u00c5\u00c6\u0005"+
		"\'\u0000\u0000\u00c6\u00c7\u0005?\u0000\u0000\u00c7\u00c9\u00056\u0000"+
		"\u0000\u00c8\u00ca\u0003 \u0010\u0000\u00c9\u00c8\u0001\u0000\u0000\u0000"+
		"\u00c9\u00ca\u0001\u0000\u0000\u0000\u00ca\u00cb\u0001\u0000\u0000\u0000"+
		"\u00cb\u00cd\u00057\u0000\u0000\u00cc\u00ce\u0005>\u0000\u0000\u00cd\u00cc"+
		"\u0001\u0000\u0000\u0000\u00ce\u00cf\u0001\u0000\u0000\u0000\u00cf\u00cd"+
		"\u0001\u0000\u0000\u0000\u00cf\u00d0\u0001\u0000\u0000\u0000\u00d0\u00d6"+
		"\u0001\u0000\u0000\u0000\u00d1\u00d4\u0005:\u0000\u0000\u00d2\u00d5\u0003"+
		"\"\u0011\u0000\u00d3\u00d5\u0003\u001e\u000f\u0000\u00d4\u00d2\u0001\u0000"+
		"\u0000\u0000\u00d4\u00d3\u0001\u0000\u0000\u0000\u00d5\u00d7\u0001\u0000"+
		"\u0000\u0000\u00d6\u00d1\u0001\u0000\u0000\u0000\u00d6\u00d7\u0001\u0000"+
		"\u0000\u0000\u00d7\u00d8\u0001\u0000\u0000\u0000\u00d8\u00d9\u00054\u0000"+
		"\u0000\u00d9\u001d\u0001\u0000\u0000\u0000\u00da\u00db\u0005;\u0000\u0000"+
		"\u00db\u00dc\u00056\u0000\u0000\u00dc\u00dd\u0005?\u0000\u0000\u00dd\u00e3"+
		"\u0003$\u0012\u0000\u00de\u00df\u00058\u0000\u0000\u00df\u00e0\u0005?"+
		"\u0000\u0000\u00e0\u00e2\u0003$\u0012\u0000\u00e1\u00de\u0001\u0000\u0000"+
		"\u0000\u00e2\u00e5\u0001\u0000\u0000\u0000\u00e3\u00e1\u0001\u0000\u0000"+
		"\u0000\u00e3\u00e4\u0001\u0000\u0000\u0000\u00e4\u00e6\u0001\u0000\u0000"+
		"\u0000\u00e5\u00e3\u0001\u0000\u0000\u0000\u00e6\u00e7\u00057\u0000\u0000"+
		"\u00e7\u001f\u0001\u0000\u0000\u0000\u00e8\u00e9\u0005=\u0000\u0000\u00e9"+
		"\u00ef\u0003$\u0012\u0000\u00ea\u00eb\u00058\u0000\u0000\u00eb\u00ec\u0005"+
		"=\u0000\u0000\u00ec\u00ee\u0003$\u0012\u0000\u00ed\u00ea\u0001\u0000\u0000"+
		"\u0000\u00ee\u00f1\u0001\u0000\u0000\u0000\u00ef\u00ed\u0001\u0000\u0000"+
		"\u0000\u00ef\u00f0\u0001\u0000\u0000\u0000\u00f0!\u0001\u0000\u0000\u0000"+
		"\u00f1\u00ef\u0001\u0000\u0000\u0000\u00f2\u00f3\u00056\u0000\u0000\u00f3"+
		"\u00f8\u0003$\u0012\u0000\u00f4\u00f5\u00058\u0000\u0000\u00f5\u00f7\u0003"+
		"$\u0012\u0000\u00f6\u00f4\u0001\u0000\u0000\u0000\u00f7\u00fa\u0001\u0000"+
		"\u0000\u0000\u00f8\u00f6\u0001\u0000\u0000\u0000\u00f8\u00f9\u0001\u0000"+
		"\u0000\u0000\u00f9\u00fb\u0001\u0000\u0000\u0000\u00fa\u00f8\u0001\u0000"+
		"\u0000\u0000\u00fb\u00fc\u00057\u0000\u0000\u00fc#\u0001\u0000\u0000\u0000"+
		"\u00fd\u00ff\u0005?\u0000\u0000\u00fe\u0100\u0005<\u0000\u0000\u00ff\u00fe"+
		"\u0001\u0000\u0000\u0000\u00ff\u0100\u0001\u0000\u0000\u0000\u0100%\u0001"+
		"\u0000\u0000\u0000\u0019*,49IM[_i~\u0083\u008c\u0094\u00ae\u00b8\u00bb"+
		"\u00c1\u00c9\u00cf\u00d4\u00d6\u00e3\u00ef\u00f8\u00ff";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}