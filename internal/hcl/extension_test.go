package hcl_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/hcl"
	"github.com/stretchr/testify/require"
)

type OwnerBlock struct {
	hcl.DefaultExtension
	ID        string            `spec:",name"`
	FirstName string            `spec:"first_name"`
	Born      int               `spec:"born"`
	Active    bool              `spec:"active"`
	BoolPtr   *bool             `spec:"bool_ptr"`
	OmitBool1 bool              `spec:"omit_bool1,omitempty"`
	OmitBool2 bool              `spec:"omit_bool2,omitempty"`
	Lit       *hcl.LiteralValue `spec:"lit"`
}

type PetBlock struct {
	hcl.DefaultExtension
	ID        string        `spec:",name"`
	Breed     string        `spec:"breed"`
	Born      int           `spec:"born"`
	Owners    []*OwnerBlock `spec:"owner"`
	RoleModel *PetBlock     `spec:"role_model"`
}

func TestInvalidExt(t *testing.T) {
	r := &hcl.Resource{}
	err := r.As(1)
	require.EqualError(t, err, "schemahcl: expected target to be a pointer")
	var p *string
	err = r.As(p)
	require.EqualError(t, err, "schemahcl: expected target to be a pointer to a struct")
}

func TestExtension(t *testing.T) {
	hcl.Register("owner", &OwnerBlock{})
	original := &hcl.Resource{
		Name: "name",
		Type: "owner",
		Attrs: []*hcl.Attr{
			hcl.StrLitAttr("first_name", "tzuri"),
			hcl.LitAttr("born", "2019"),
			hcl.LitAttr("active", "true"),
			hcl.LitAttr("bool_ptr", "true"),
			hcl.LitAttr("omit_bool1", "true"),
			hcl.LitAttr("lit", "1000"),
			hcl.LitAttr("extra", "true"),
		},
		Children: []*hcl.Resource{
			{
				Name: "extra",
				Type: "extra",
			},
		},
	}
	owner := OwnerBlock{}
	err := original.As(&owner)
	require.NoError(t, err)
	require.EqualValues(t, "tzuri", owner.FirstName)
	require.EqualValues(t, "name", owner.ID)
	require.EqualValues(t, 2019, owner.Born)
	require.EqualValues(t, true, owner.Active)
	require.NotNil(t, owner.BoolPtr)
	require.EqualValues(t, true, *owner.BoolPtr)
	require.EqualValues(t, hcl.LitAttr("lit", "1000").V, owner.Lit)
	attr, ok := owner.Remain().Attr("extra")
	require.True(t, ok)
	eb, err := attr.Bool()
	require.NoError(t, err)
	require.True(t, eb)

	scan := &hcl.Resource{}
	err = scan.Scan(&owner)
	require.NoError(t, err)
	require.EqualValues(t, original, scan)
}

func TestNested(t *testing.T) {
	hcl.Register("pet", &PetBlock{})
	pet := &hcl.Resource{
		Name: "donut",
		Type: "pet",
		Attrs: []*hcl.Attr{
			hcl.StrLitAttr("breed", "golden retriever"),
			hcl.LitAttr("born", "2002"),
		},
		Children: []*hcl.Resource{
			{
				Name: "rotemtam",
				Type: "owner",
				Attrs: []*hcl.Attr{
					hcl.StrLitAttr("first_name", "rotem"),
					hcl.LitAttr("born", "1985"),
					hcl.LitAttr("active", "true"),
				},
			},
			{
				Name: "gonnie",
				Type: "role_model",
				Attrs: []*hcl.Attr{
					hcl.StrLitAttr("breed", "golden retriever"),
					hcl.LitAttr("born", "1998"),
				},
			},
		},
	}
	pb := PetBlock{}
	err := pet.As(&pb)
	require.NoError(t, err)
	require.EqualValues(t, PetBlock{
		ID:    "donut",
		Breed: "golden retriever",
		Born:  2002,
		Owners: []*OwnerBlock{
			{ID: "rotemtam", FirstName: "rotem", Born: 1985, Active: true},
		},
		RoleModel: &PetBlock{
			ID:    "gonnie",
			Breed: "golden retriever",
			Born:  1998,
		},
	}, pb)
	scan := &hcl.Resource{}
	err = scan.Scan(&pb)
	require.NoError(t, err)
	require.EqualValues(t, pet, scan)
	name, err := pet.FinalName()
	require.NoError(t, err)
	require.EqualValues(t, "donut", name)
}

func TestRef(t *testing.T) {
	type A struct {
		Name string   `spec:",name"`
		User *hcl.Ref `spec:"user"`
	}
	hcl.Register("a", &A{})
	resource := &hcl.Resource{
		Name: "x",
		Type: "a",
		Attrs: []*hcl.Attr{
			{
				K: "user",
				V: &hcl.Ref{V: "$user.rotemtam"},
			},
		},
	}
	var a A
	err := resource.As(&a)
	require.NoError(t, err)
	require.EqualValues(t, &hcl.Ref{V: "$user.rotemtam"}, a.User)
	scan := &hcl.Resource{}
	err = scan.Scan(&a)
	require.NoError(t, err)
	require.EqualValues(t, resource, scan)
}

func TestListRef(t *testing.T) {
	type B struct {
		Name  string     `spec:",name"`
		Users []*hcl.Ref `spec:"users"`
	}
	hcl.Register("b", &B{})
	resource := &hcl.Resource{
		Name: "x",
		Type: "b",
		Attrs: []*hcl.Attr{
			{
				K: "users",
				V: &hcl.ListValue{
					V: []hcl.Value{
						&hcl.Ref{V: "$user.a8m"},
						&hcl.Ref{V: "$user.rotemtam"},
					},
				},
			},
		},
	}

	var b B
	err := resource.As(&b)
	require.NoError(t, err)
	require.Len(t, b.Users, 2)
	require.EqualValues(t, []*hcl.Ref{
		{V: "$user.a8m"},
		{V: "$user.rotemtam"},
	}, b.Users)
	scan := &hcl.Resource{}
	err = scan.Scan(&b)
	require.NoError(t, err)
	require.EqualValues(t, resource, scan)
}

func TestNameAttr(t *testing.T) {
	type Named struct {
		Name string `spec:"name,name"`
	}
	hcl.Register("named", &Named{})
	resource := &hcl.Resource{
		Name: "id",
		Type: "named",
		Attrs: []*hcl.Attr{
			hcl.StrLitAttr("name", "kwil"),
		},
	}
	var tgt Named
	err := resource.As(&tgt)
	require.NoError(t, err)
	require.EqualValues(t, "kwil", tgt.Name)
	name, err := resource.FinalName()
	require.NoError(t, err)
	require.EqualValues(t, name, "kwil")
}
