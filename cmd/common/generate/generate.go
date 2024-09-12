package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spf13/cobra/doc"
)

type writeFunc func(string) (*os.File, error)

func WriteDocs(cmd *cobra.Command, dir string) error {
	return writeCmd(cmd, ".", 0, func(s string) (*os.File, error) {
		fullPath := filepath.Join(dir, s)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			return nil, err
		}

		return os.Create(fullPath)
	}, true)
}

func writeCmd(cmd *cobra.Command, dir string, idx int, write writeFunc, first bool) error {
	subCommands := cmd.Commands()
	if len(subCommands) == 0 {
		return createCmdFile(cmd, dir, idx, write)
	}

	err := createIndexFile(cmd, dir, write, first)
	if err != nil {
		return err
	}

	subDir := dir + "/" + cmd.Name()
	for i, subCmd := range subCommands {
		err := writeCmd(subCmd, subDir, i+1, write, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// createIndexFile creates a file for a command that has subcommands.
// If "first" is true, then this is the top-level command.
func createIndexFile(cmd *cobra.Command, dir string, write writeFunc, first bool) error {
	dir = dir + "/" + cmd.Name()

	header := docusaursHeader{
		sidebar_position: 99, // we want these to always be at the bottom of the sidebar
		sidebar_label:    cmd.Name(),
		id:               strings.ReplaceAll(cmd.CommandPath(), " ", "-"),
		title:            cmd.CommandPath(),
		description:      cmd.Short,
		slug:             getSlug(cmd),
	}
	if first {
		header.sidebar_label = "Reference"
	}

	file, err := write(dir + "/index.md")
	if err != nil {
		return err
	}

	file.WriteString(header.String() + "\n\n")

	err = doc.GenMarkdownCustom(cmd, file, linkHandler(dir))
	if err != nil {
		return err
	}

	return nil
}

// createCmdFile creates a file for the command, and writes the command's documentation to it.
// it does not call subcommands.
func createCmdFile(cmd *cobra.Command, dir string, idx int, write writeFunc) error {
	file, err := write(dir + "/" + cmd.Name() + ".mdx")
	if err != nil {
		return err
	}

	header := docusaursHeader{
		sidebar_position: idx,
		sidebar_label:    cmd.Name(),
		id:               strings.ReplaceAll(cmd.CommandPath(), " ", "-"),
		title:            cmd.CommandPath(),
		description:      cmd.Short,
		slug:             getSlug(cmd),
	}

	file.WriteString(header.String() + "\n\n")

	err = doc.GenMarkdownCustom(cmd, file, linkHandler(dir))
	if err != nil {
		return err
	}

	return nil
}

func linkHandler(dir string) func(string) string {
	return func(s string) string {
		// trying just linking ids??
		s = strings.TrimSuffix(s, ".md")
		s = strings.ReplaceAll(s, "_", "/")
		return "/docs/ref/" + s
	}
}

// getSlug gets a slug for the command
func getSlug(cmd *cobra.Command) string {
	return "/ref/" + strings.ReplaceAll(cmd.CommandPath(), " ", "/")
}

// docusaurusHeader is a page header for a doc page in a docusaurus site.
type docusaursHeader struct {
	sidebar_position int
	sidebar_label    string
	id               string
	title            string
	description      string
	slug             string
}

// String returns the header as a string.
func (d *docusaursHeader) String() string {
	return fmt.Sprintf(headerString, d.sidebar_position, d.sidebar_label, d.id, d.title, d.description, d.slug)
}

var headerString = `---
sidebar_position: %d
sidebar_label: "%s"
id: "%s"
title: "%s"
description: "%s"
slug: %s
---`
