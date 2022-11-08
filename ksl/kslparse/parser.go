package kslparse

import (
	"fmt"
	"os"
	"path/filepath"

	"ksl"
	"ksl/kslspec"
	"ksl/kslsyntax"
)

type Parser struct {
	files map[string]*kslspec.File
}

func NewParser() *Parser {
	return &Parser{
		files: map[string]*kslspec.File{},
	}
}

func (p *Parser) Parse(src []byte, filename string) (*kslspec.File, ksl.Diagnostics) {
	if existing := p.files[filename]; existing != nil {
		return existing, nil
	}

	doc, diags := kslsyntax.Parse(src, filename, ksl.InitialPos)

	file := &kslspec.File{
		Body:  doc,
		Bytes: src,
	}

	p.files[filename] = file
	return file, diags
}

func (p *Parser) ParseFile(filename string) (*kslspec.File, ksl.Diagnostics) {
	if existing := p.files[filename]; existing != nil {
		return existing, nil
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q could not be read.", filename),
			},
		}
	}

	return p.Parse(src, filename)
}

func (p *Parser) Sources() map[string][]byte {
	ret := make(map[string][]byte)
	for fn, f := range p.files {
		ret[fn] = f.Bytes
	}
	return ret
}

func (p *Parser) Files() map[string]*kslspec.File {
	return p.files
}

func (p *Parser) FileSet() *kslspec.FileSet {
	return &kslspec.FileSet{Files: p.files}
}

func ParseKwilFiles(paths ...string) (*kslspec.FileSet, error) {
	p := NewParser()
	for _, path := range paths {
		switch stat, err := os.Stat(path); {
		case err != nil:
			return nil, err
		case stat.IsDir():
			dir, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			for _, f := range dir {
				if f.IsDir() {
					continue
				}
				if err := mayParse(p, filepath.Join(path, f.Name())); err != nil {
					return nil, err
				}
			}
		default:
			if err := mayParse(p, path); err != nil {
				return nil, err
			}
		}
	}
	if len(p.Files()) == 0 {
		return nil, fmt.Errorf("no schema files found in: %s", paths)
	}

	return &kslspec.FileSet{Files: p.Files()}, nil
}

func mayParse(p *Parser, path string) error {
	if n := filepath.Base(path); filepath.Ext(n) != ".kwil" {
		return nil
	}
	switch _, diag := p.ParseFile(path); {
	case diag.HasErrors():
		return diag
	default:
		return nil
	}
}
