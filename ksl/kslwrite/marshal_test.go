package kslwrite_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"ksl/kslwrite"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name   string
		file   *kslwrite.File
		result string
	}{
		{
			name: "simple",
			file: &kslwrite.File{
				Directives: []*kslwrite.Directive{
					{Name: "import", Value: "\"ksl\""},
					{Name: "import", Value: "\"test_value\""},
					{Name: "option", Key: "test", Value: "\"test_value\""},
				},
				Blocks: []*kslwrite.Block{
					{
						Type:     "message",
						Name:     "TestMessage",
						Modifier: "extends",
						Target:   "TestMessageSuper",
						Labels: []*kslwrite.Kwarg{
							{Key: "test", Value: "\"test_value\""},
							{Key: "default"},
						},
						Body: &kslwrite.BlockBody{
							Attributes: []*kslwrite.Attribute{
								{Key: "optional", Value: "false"},
								{Key: "name", Value: "some_name"},
							},
							Blocks: []*kslwrite.Block{
								{
									Type: "message",
									Name: "ChildMessage",
									Body: &kslwrite.BlockBody{},
								},
								{
									Type: "message",
									Name: "ChildMessage2",
									Body: &kslwrite.BlockBody{
										Annotations: []*kslwrite.Annotation{
											{Name: "test", Args: &kslwrite.ArgList{Args: []string{"\"test_value\""}}},
										},
									},
								},
							},
							Definitions: []*kslwrite.Definition{
								{
									Name: "test",
									Type: "string",
									Annotations: []*kslwrite.Annotation{
										{Name: "size", Args: &kslwrite.ArgList{Args: []string{"1024"}}},
									},
								},
							},
							Annotations: []*kslwrite.Annotation{
								{Name: "foreign_key", Args: &kslwrite.ArgList{Args: []string{"1024"}, Kwargs: []*kslwrite.Kwarg{{Key: "test", Value: "\"test_value\""}}}},
							},
						},
					},
				},
			},
			result: `@import "ksl"
@import "test_value"

@option test = "test_value"

message TestMessage extends TestMessageSuper [test="test_value", default] {
    test: string @size(1024)
    optional = false
    name = some_name
    message ChildMessage {}
    message ChildMessage2 {
        @@test("test_value")
    }
    @@foreign_key(1024, test="test_value")
}
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf strings.Builder
			err := kslwrite.Marshal(&buf, test.file)
			require.NoError(t, err)
			t.Log(buf.String())
			require.Equal(t, test.result, buf.String())
		})
	}
}
