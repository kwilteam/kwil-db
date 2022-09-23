package hcl_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"kwil/x/schemadef/hcl"
)

func TestReferences(t *testing.T) {
	f := `
backend "app" {
	image = "ariga/app:1.2.3"
	addr = "127.0.0.1:8081"
}
backend "admin" {
	image = "ariga/admin:1.2.3"
	addr = "127.0.0.1:8082"
}
endpoint "home" {
	path = "/"
	addr = backend.app.addr
	timeout_ms = config.defaults.timeout_ms
	retry = config.defaults.retry
	description = "default: ${config.defaults.description}"
}
endpoint "admin" {
	path = "/admin"
	addr = backend.admin.addr
}
config "defaults" {
	timeout_ms = 10
	retry = false
	description = "generic"
}
`
	type (
		Backend struct {
			Name  string `spec:",name"`
			Image string `spec:"image"`
			Addr  string `spec:"addr"`
		}
		Endpoint struct {
			Name      string `spec:",name"`
			Path      string `spec:"path"`
			Addr      string `spec:"addr"`
			TimeoutMs int    `spec:"timeout_ms"`
			Retry     bool   `spec:"retry"`
			Desc      string `spec:"description"`
		}
	)
	var test struct {
		Backends  []*Backend  `spec:"backend"`
		Endpoints []*Endpoint `spec:"endpoint"`
	}
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, []*Endpoint{
		{
			Name:      "home",
			Path:      "/",
			Addr:      "127.0.0.1:8081",
			Retry:     false,
			TimeoutMs: 10,
			Desc:      "default: generic",
		},
		{
			Name: "admin",
			Path: "/admin",
			Addr: "127.0.0.1:8082",
		},
	}, test.Endpoints)
}

func TestUnlabeledBlockReferences(t *testing.T) {
	f := `
country "israel" {
    metadata {
        phone_prefix = "972"
    }
    metadata {
        phone_prefix = "123"
    }
    metadata "geo" {
		continent = "asia"
    }
}

metadata  = country.israel.metadata.0
phone_prefix = country.israel.metadata.0.phone_prefix
phone_prefix_2 = country.israel.metadata.1.phone_prefix
continent = country.israel.metadata.geo.continent
`
	type (
		Metadata struct {
			PhonePrefix string `spec:"phone_prefix"`
			Continent   string `spec:"continent"`
		}
		Country struct {
			Metadata []*Metadata `spec:"metadata"`
		}
		Test struct {
			Countries    []*Country `spec:"country"`
			MetadataRef  *hcl.Ref   `spec:"metadata"`
			PhonePrefix  string     `spec:"phone_prefix"`
			PhonePrefix2 string     `spec:"phone_prefix_2"`
			Continent    string     `spec:"continent"`
		}
	)
	var test Test
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, test, Test{
		Countries: []*Country{
			{
				Metadata: []*Metadata{
					{PhonePrefix: "972"},
					{PhonePrefix: "123"},
					{Continent: "asia"},
				},
			},
		},
		MetadataRef:  &hcl.Ref{V: "$country.israel.$metadata.0"},
		PhonePrefix:  "972",
		PhonePrefix2: "123",
		Continent:    "asia",
	})
}

func TestNestedReferences(t *testing.T) {
	f := `
country "israel" {
	city "tel_aviv" {
		phone_area_code = "03"
	}
	city "jerusalem" {
		phone_area_code = "02"
	}
	city "givatayim" {
		phone_area_code = country.israel.city.tel_aviv.phone_area_code
	}
}
`
	type (
		City struct {
			Name          string `spec:",name"`
			PhoneAreaCode string `spec:"phone_area_code"`
		}
		Country struct {
			Name   string  `spec:",name"`
			Cities []*City `spec:"city"`
		}
	)
	var test struct {
		Countries []*Country `spec:"country"`
	}
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	israel := &Country{
		Name: "israel",
		Cities: []*City{
			{Name: "tel_aviv", PhoneAreaCode: "03"},
			{Name: "jerusalem", PhoneAreaCode: "02"},
			{Name: "givatayim", PhoneAreaCode: "03"},
		},
	}
	require.NoError(t, err)
	require.EqualValues(t, israel, test.Countries[0])
}

func TestBlockReference(t *testing.T) {
	f := `person "jon" {
}
pet "garfield" {
  type  = "cat"
  owner = person.jon
}
`
	type (
		Person struct {
			Name string `spec:",name"`
		}
		Pet struct {
			Name  string   `spec:",name"`
			Type  string   `spec:"type"`
			Owner *hcl.Ref `spec:"owner"`
		}
	)
	var test struct {
		People []*Person `spec:"person"`
		Pets   []*Pet    `spec:"pet"`
	}
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, &Pet{
		Name:  "garfield",
		Type:  "cat",
		Owner: &hcl.Ref{V: "$person.jon"},
	}, test.Pets[0])
	marshal, err := hcl.Marshal(&test)
	require.NoError(t, err)
	require.EqualValues(t, f, string(marshal))
}

func TestListRefs(t *testing.T) {
	f := `
user "simba" {

}
user "mufasa" {

}
group "lion_kings" {
	members = [
		user.simba,
		user.mufasa,
	]
}
`
	type (
		User struct {
			Name string `spec:",name"`
		}
		Group struct {
			Name    string     `spec:",name"`
			Members []*hcl.Ref `spec:"members"`
		}
	)
	var test struct {
		Users  []*User  `spec:"user"`
		Groups []*Group `spec:"group"`
	}
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, &Group{
		Name: "lion_kings",
		Members: []*hcl.Ref{
			{V: "$user.simba"},
			{V: "$user.mufasa"},
		},
	}, test.Groups[0])
	_, err = hcl.Marshal(&test)
	require.NoError(t, err)
}

func TestNestedDifference(t *testing.T) {
	f := `
person "john" {
	nickname = "jonnie"
	hobby "hockey" {
		active = true
	}
}
person "jane" {
	nickname = "janie"
	hobby "football" {
		budget = 1000
	}
	car "ferrari" {
		year = 1960
	}
}
`
	type (
		Hobby struct {
			Name   string `spec:",name"`
			Active bool   `spec:"active"`
			Budget int    `spec:"budget"`
		}
		Car struct {
			Name string `spec:",name"`
			Year int    `spec:"year"`
		}
		Person struct {
			Name     string   `spec:",name"`
			Nickname string   `spec:"nickname"`
			Hobbies  []*Hobby `spec:"hobby"`
			Car      *Car     `spec:"car"`
		}
	)
	var test struct {
		People []*Person `spec:"person"`
	}
	err := hcl.New().EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	john := &Person{
		Name:     "john",
		Nickname: "jonnie",
		Hobbies: []*Hobby{
			{Name: "hockey", Active: true},
		},
	}
	require.EqualValues(t, john, test.People[0])
	jane := &Person{
		Name:     "jane",
		Nickname: "janie",
		Hobbies: []*Hobby{
			{Name: "football", Budget: 1000},
		},
		Car: &Car{
			Name: "ferrari",
			Year: 1960,
		},
	}
	require.EqualValues(t, jane, test.People[1])
}

func TestSchemaRefParse(t *testing.T) {
	type Point struct {
		Z []*hcl.Ref `spec:"z"`
	}
	var test = struct {
		Points []*Point `spec:"point"`
	}{
		Points: []*Point{
			{Z: []*hcl.Ref{{V: "$a"}}},
			{Z: []*hcl.Ref{{V: "b"}}},
		},
	}
	b, err := hcl.Marshal(&test)
	require.NoError(t, err)
	expected :=
		`point {
  z = [a]
}
point {
  z = [b]
}
`
	require.Equal(t, expected, string(b))
}

func TestWithTypes(t *testing.T) {
	f := `first    = int
second   = bool
third    = int(10)
sized    = varchar(255)
variadic = enum("a","b","c")
`
	s := hcl.New(
		hcl.WithTypes(
			[]*hcl.TypeSpec{
				{Name: "bool", T: "bool"},
				{
					Name: "int",
					T:    "int",
					Attributes: []*hcl.TypeAttr{
						{Name: "size", Kind: reflect.Int, Required: false},
						{Name: "unsigned", Kind: reflect.Bool, Required: false},
					},
				},
				{
					Name: "varchar",
					T:    "varchar",
					Attributes: []*hcl.TypeAttr{
						{Name: "size", Kind: reflect.Int, Required: false},
					},
				},
				{
					Name: "enum",
					T:    "enum",
					Attributes: []*hcl.TypeAttr{
						{Name: "values", Kind: reflect.Slice, Required: false},
					},
				},
			},
		),
	)
	var test struct {
		First    *hcl.Type `spec:"first"`
		Second   *hcl.Type `spec:"second"`
		Third    *hcl.Type `spec:"third"`
		Varchar  *hcl.Type `spec:"sized"`
		Variadic *hcl.Type `spec:"variadic"`
	}
	err := s.EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, "int", test.First.T)
	require.EqualValues(t, "bool", test.Second.T)
	require.EqualValues(t, &hcl.Type{
		T: "varchar",
		Attrs: []*hcl.Attr{
			{K: "size", V: &hcl.LiteralValue{V: "255"}},
		},
	}, test.Varchar)
	require.EqualValues(t, &hcl.Type{
		T: "enum",
		Attrs: []*hcl.Attr{
			{
				K: "values",
				V: &hcl.ListValue{
					V: []hcl.Value{
						&hcl.LiteralValue{V: `"a"`},
						&hcl.LiteralValue{V: `"b"`},
						&hcl.LiteralValue{V: `"c"`},
					},
				},
			},
		},
	}, test.Variadic)
	require.EqualValues(t, &hcl.Type{
		T: "int",
		Attrs: []*hcl.Attr{
			{K: "size", V: &hcl.LiteralValue{V: "10"}},
		},
	}, test.Third)
	after, err := s.MarshalSpec(&test)
	require.NoError(t, err)
	require.EqualValues(t, f, string(after))
}

func TestEmptyStrSQL(t *testing.T) {
	s := hcl.New(hcl.WithTypes(nil))
	h := `x = sql("")`
	err := s.EvalBytes([]byte(h), &struct{}{}, nil)
	require.ErrorContains(t, err, "empty expression")
}

func TestOptionalArgs(t *testing.T) {
	s := hcl.New(
		hcl.WithTypes([]*hcl.TypeSpec{
			{
				T:    "float",
				Name: "float",
				Attributes: []*hcl.TypeAttr{
					{Name: "precision", Kind: reflect.Int, Required: false},
					{Name: "scale", Kind: reflect.Int, Required: false},
				},
			},
		}),
	)
	f := `arg_0 = float
arg_1 = float(10)
arg_2 = float(10,2)
`
	var test struct {
		Arg0 *hcl.Type `spec:"arg_0"`
		Arg1 *hcl.Type `spec:"arg_1"`
		Arg2 *hcl.Type `spec:"arg_2"`
	}
	err := s.EvalBytes([]byte(f), &test, nil)
	require.NoError(t, err)
	require.Nil(t, test.Arg0.Attrs)
	require.EqualValues(t, []*hcl.Attr{
		hcl.LitAttr("precision", "10"),
	}, test.Arg1.Attrs)
	require.EqualValues(t, []*hcl.Attr{
		hcl.LitAttr("precision", "10"),
		hcl.LitAttr("scale", "2"),
	}, test.Arg2.Attrs)
}

func TestQualifiedRefs(t *testing.T) {
	h := `user "kwil" "cli" {
	version = "v0.3.9"
}
v = user.kwil.cli.version
r = user.kwil.cli
`
	var test struct {
		V string   `spec:"v"`
		R *hcl.Ref `spec:"r"`
	}
	err := hcl.New().EvalBytes([]byte(h), &test, nil)
	require.NoError(t, err)
	require.EqualValues(t, "v0.3.9", test.V)
	require.EqualValues(t, "$user.kwil.cli", test.R.V)
}

func TestInputValues(t *testing.T) {
	h := `
variable "name" {
  type = string
}

variable "default" {
  type = string
  default = "hello"
}

variable "int" {
  type = int
}

variable "bool" {
  type = bool
}

name = var.name
default = var.default
int = var.int
bool = var.bool
`
	state := hcl.New()
	var test struct {
		Name    string `spec:"name"`
		Default string `spec:"default"`
		Int     int    `spec:"int"`
		Bool    bool   `spec:"bool"`
	}
	err := state.EvalBytes([]byte(h), &test, map[string]string{
		"name": "rotemtam",
		"int":  "42",
		"bool": "true",
	})
	require.NoError(t, err)
	require.EqualValues(t, "rotemtam", test.Name)
	require.EqualValues(t, "hello", test.Default)
	require.EqualValues(t, 42, test.Int)
	require.EqualValues(t, true, test.Bool)
}
