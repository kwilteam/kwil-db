package hcl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttributes(t *testing.T) {
	f := `i  = 1
b  = true
s  = "hello, world"
sl = ["hello", "world"]
bl = [true, false]
`
	var test struct {
		Int        int      `spec:"i"`
		Bool       bool     `spec:"b"`
		Str        string   `spec:"s"`
		StringList []string `spec:"sl"`
		BoolList   []bool   `spec:"bl"`
	}
	err := New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, test.Int)
	require.EqualValues(t, true, test.Bool)
	require.EqualValues(t, "hello, world", test.Str)
	require.EqualValues(t, []string{"hello", "world"}, test.StringList)
	require.EqualValues(t, []bool{true, false}, test.BoolList)
	marshal, err := Marshal(&test)
	require.NoError(t, err)
	require.EqualValues(t, f, string(marshal))
}

func TestResource(t *testing.T) {
	f := `endpoint "/hello" {
  description = "the hello handler"
  timeout_ms  = 100
  handler {
    active = true
    addr   = ":8080"
  }
}
`
	type (
		Handler struct {
			Active bool   `spec:"active"`
			Addr   string `spec:"addr"`
		}

		Endpoint struct {
			Name        string   `spec:",name"`
			Description string   `spec:"description"`
			TimeoutMs   int      `spec:"timeout_ms"`
			Handler     *Handler `spec:"handler"`
		}
		File struct {
			Endpoints []*Endpoint `spec:"endpoint"`
		}
	)
	var test File
	err := New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.Len(t, test.Endpoints, 1)
	expected := &Endpoint{
		Name:        "/hello",
		Description: "the hello handler",
		TimeoutMs:   100,
		Handler: &Handler{
			Active: true,
			Addr:   ":8080",
		},
	}
	require.EqualValues(t, expected, test.Endpoints[0])
	buf, err := Marshal(&test)
	require.NoError(t, err)
	require.EqualValues(t, f, string(buf))
}

func TestInterface(t *testing.T) {
	type (
		Animal interface {
			animal()
		}
		Parrot struct {
			Animal
			Name string `spec:",name"`
			Boss string `spec:"boss"`
		}
		Lion struct {
			Animal
			Name   string `spec:",name"`
			Friend string `spec:"friend"`
		}
		Zoo struct {
			Animals []Animal `spec:""`
		}
		Cast struct {
			Animal Animal `spec:""`
		}
	)
	Register("lion", &Lion{})
	Register("parrot", &Parrot{})
	t.Run("single", func(t *testing.T) {
		f := `
cast "lion_king" {
	lion "simba" {
		friend = "rafiki"
	}
}
`
		var test struct {
			Cast *Cast `spec:"cast"`
		}
		err := New().EvalBytes([]byte(f), &test, nil)
		require.NoError(t, err)
		require.EqualValues(t, &Cast{
			Animal: &Lion{
				Name:   "simba",
				Friend: "rafiki",
			},
		}, test.Cast)
	})
	t.Run("slice", func(t *testing.T) {
		f := `
zoo "ramat_gan" {
	lion "simba" {
		friend = "rafiki"
	}
	parrot "iago" {
		boss = "jafar"
	}
}
`
		var test struct {
			Zoo *Zoo `spec:"zoo"`
		}
		err := New().EvalBytes([]byte(f), &test, nil)
		require.NoError(t, err)
		require.EqualValues(t, &Zoo{
			Animals: []Animal{
				&Lion{
					Name:   "simba",
					Friend: "rafiki",
				},
				&Parrot{
					Name: "iago",
					Boss: "jafar",
				},
			},
		}, test.Zoo)
	})
}

func TestQualified(t *testing.T) {
	type Person struct {
		Name  string `spec:",name"`
		Title string `spec:",qualifier"`
	}
	var test struct {
		Person *Person `spec:"person"`
	}
	h := `person "dr" "jekyll" {
}
`
	err := New().EvalBytes([]byte(h), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, test.Person, &Person{
		Title: "dr",
		Name:  "jekyll",
	})
	out, err := Marshal(&test)
	require.NoError(t, err)
	require.EqualValues(t, h, string(out))
}

func TestNameAttr(t *testing.T) {
	h := `
named "block_id" {
  name = "kwil"
}
ref = named.block_id.name
`
	type Named struct {
		Name string `spec:"name,name"`
	}
	var test struct {
		Named *Named `spec:"named"`
		Ref   string `spec:"ref"`
	}
	err := New().EvalBytes([]byte(h), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, &Named{
		Name: "kwil",
	}, test.Named)
	require.EqualValues(t, "kwil", test.Ref)
}

func TestRefPatch(t *testing.T) {
	type (
		Family struct {
			Name string `spec:"name,name"`
		}
		Person struct {
			Name   string `spec:",name"`
			Family *Ref   `spec:"family"`
		}
	)
	Register("family", &Family{})
	Register("person", &Person{})
	var test struct {
		Families []*Family `spec:"family"`
		People   []*Person `spec:"person"`
	}
	h := `
variable "family_name" {
  type = string
}

family "default" {
	name = var.family_name
}

person "rotem" {
	family = family.default
}
`
	err := New().EvalBytes([]byte(h), &test, map[string]string{"family_name": "tam"})
	require.NoError(t, err)
	require.EqualValues(t, "$family.tam", test.People[0].Family.V)
}
