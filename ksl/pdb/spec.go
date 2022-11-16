package pdb

import "ksl/spec"

const (
	Ascending  string = "Asc"
	Descending string = "Desc"
)

var Spec = spec.Schema(
	spec.OptSchema.WithDirectives(
		spec.Annot("backend", spec.OptAnnot.Single(), spec.OptAnnot.Args(
			spec.RequiredArg("name", spec.OneOf(spec.Constant("postgres")), spec.OptArg.Default()),
		)),
	),
	spec.OptSchema.WithModelBlockAnnotations(
		spec.Annot("ignore"),
		spec.Annot("id", spec.OptAnnot.Args(
			spec.RequiredArg("fields", spec.FieldRefs(), spec.OptArg.Default()),
			spec.Arg("name", spec.String()),
		)),
		spec.Annot("unique", spec.OptAnnot.Args(
			spec.RequiredArg("fields", spec.FieldRefs(), spec.OptArg.Default()),
			spec.Arg("name", spec.String()),
		)),
		spec.Annot("index", spec.OptAnnot.Args(
			spec.RequiredArg("fields", spec.FieldRefs(), spec.OptArg.Default()),
			spec.Arg("name", spec.String()),
			spec.Arg("type", spec.Enum("BTree", "Hash", "Gist", "Gin", "SpGist", "Brin")),
		)),
		spec.Annot("map", spec.OptAnnot.Args(
			spec.RequiredArg("name", spec.String(), spec.OptArg.Default()),
		)),
	),
	spec.OptSchema.WithModelFieldRelationAnnotations(
		spec.Annot("ignore"),
		spec.Annot("ref", spec.OptAnnot.Args(
			spec.Arg("name", spec.String(), spec.OptArg.Default()),
			spec.Arg("fields", spec.FieldRefs()),
			spec.Arg("references", spec.FieldRefs()),
			spec.Arg("onDelete", spec.Enum("NoAction", "Restrict", "Cascade", "SetNull", "SetDefault")),
			spec.Arg("onUpdate", spec.Enum("NoAction", "Restrict", "Cascade", "SetNull", "SetDefault")),
		)),
	),
	spec.OptSchema.WithModelFieldScalarAnnotations(
		spec.Annot("ignore"),
		spec.Annot("id", spec.OptAnnot.Args(
			spec.Arg("name", spec.String(), spec.OptArg.Default()),
			spec.Arg("sort", spec.Enum(Ascending, Descending)),
		)),
		spec.Annot("unique", spec.OptAnnot.Args(
			spec.Arg("name", spec.String(), spec.OptArg.Default()),
			spec.Arg("sort", spec.Enum(Ascending, Descending)),
		)),
		spec.Annot("default", spec.OptAnnot.Args(
			spec.RequiredArg("value",
				spec.OneOf(
					spec.DbGenerated(),
					spec.List(spec.AnyScalar()),
					spec.AnyScalar(),
				),
				spec.OptArg.Default(),
			),
		)),
		spec.Annot("map", spec.OptAnnot.Args(
			spec.RequiredArg("name", spec.String(), spec.OptArg.Default()),
		)),
	),
	spec.OptSchema.WithEnumBlockAnnotations(
		spec.Annot("map", spec.OptAnnot.Args(
			spec.RequiredArg("name", spec.String(), spec.OptArg.Default()),
		)),
	),
	spec.OptSchema.WithEnumFieldAnnotations(
		spec.Annot("map", spec.OptAnnot.Args(
			spec.RequiredArg("name", spec.String(), spec.OptArg.Default()),
		)),
	),
)
