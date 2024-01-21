package sqlparser

import (
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/stretchr/testify/assert"
)

type mockParserRuleContext struct {
	*antlr.BaseParserRuleContext
}

type mockToken struct {
	*antlr.CommonToken
}

func (m mockToken) GetLine() int {
	return 1
}

func (m mockToken) GetColumn() int {
	return 2
}

func (m mockParserRuleContext) GetStart() antlr.Token {
	return mockToken{}
}

func (m mockParserRuleContext) GetStop() antlr.Token {
	return mockToken{}
}

func TestAstBuilder_getPos(t *testing.T) {
	mockCtx := mockParserRuleContext{}

	ab := newAstBuilder(false, false)
	pos := ab.getPos(mockCtx)
	assert.Nil(t, pos)

	ab = newAstBuilder(false, true)
	pos = ab.getPos(mockCtx)
	assert.NotNil(t, pos)
	assert.Equal(t, 1, pos.StartLine)
	assert.Equal(t, 2, pos.StartColumn)
	assert.Equal(t, 1, pos.EndLine)
	assert.Equal(t, 2, pos.EndColumn)

}
