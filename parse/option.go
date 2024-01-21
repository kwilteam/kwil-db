package parse

import "github.com/antlr4-go/antlr/v4"

// parseConfig is the configuration for parser(sql&action).
type parseConfig struct {
	Trace    bool
	TrackPos bool

	LexerErrorListener  antlr.ErrorListener
	ParserErrorListener antlr.ErrorListener
}

func DefaultOpt() *parseConfig {
	return new(parseConfig)
}

type Option interface {
	ApplyOption(*parseConfig)
}

type optionFunc func(*parseConfig)

func (fn optionFunc) ApplyOption(opt *parseConfig) {
	fn(opt)
}

func WithTrace(trace bool) Option {
	return optionFunc(func(cfg *parseConfig) {
		cfg.Trace = trace
	})
}

func WithTrackPos(trackPos bool) Option {
	return optionFunc(func(cfg *parseConfig) {
		cfg.TrackPos = trackPos
	})
}

func WithLexerErrorListener(listener antlr.ErrorListener) Option {
	return optionFunc(func(cfg *parseConfig) {
		cfg.LexerErrorListener = listener
	})
}

func WithParserErrorListener(listener antlr.ErrorListener) Option {
	return optionFunc(func(cfg *parseConfig) {
		cfg.ParserErrorListener = listener
	})
}
