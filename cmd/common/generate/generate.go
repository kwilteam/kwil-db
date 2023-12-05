package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
	})
}

func writeCmd(cmd *cobra.Command, dir string, idx int, write writeFunc) error {
	subCommands := cmd.Commands()
	if len(subCommands) == 0 {
		return createCmdFile(cmd, dir, idx, write)
	}

	err := createIndexFile(cmd, dir, write)
	if err != nil {
		return err
	}

	subDir := dir + "/" + cmd.Name()
	for i, subCmd := range subCommands {
		err := writeCmd(subCmd, subDir, i+1, write)
		if err != nil {
			return err
		}
	}

	return nil
}

// createIndexFile creates a file for a command that has subcommands.
func createIndexFile(cmd *cobra.Command, dir string, write writeFunc) error {
	dir = dir + "/" + cmd.Name()

	header := docusaursHeader{
		sidebar_position: 0,
		sidebar_label:    cmd.Name(),
		id:               strings.ReplaceAll(cmd.CommandPath(), " ", "-"),
		title:            cmd.CommandPath(),
		description:      cmd.Short,
	}

	file, err := write(dir + "/index.md")
	if err != nil {
		return err
	}

	file.WriteString(header.String() + "\n\n")

	err = genMarkdownCustom(cmd, file, linkHandler(dir))
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
	}

	file.WriteString(header.String() + "\n\n")

	err = genMarkdownCustom(cmd, file, linkHandler(dir))
	if err != nil {
		return err
	}

	return nil
}

// linkHandler creates the relative file path for links
// the value passed to the return func (the link) is passed as root-cmd_sub-cmd_sub-sub-cmd
// dir is passed as ./root-cmd/sub-cmd/sub-sub-cmd
// if the link is shorter than the dir, such as root-cmd_sub-cmd and dir is root-cmd/sub-cmd/sub-sub-cmd
// then the result should be ../ (since it is relative to the current directory)
// If the link is longer than the dir, such as root-cmd_sub-cmd_sub-sub-cmd and dir is root-cmd/sub-cmd
// then the result should be ./sub-sub-cmd
// If the link is the same as the dir, such as root-cmd_sub-cmd_sub-sub-cmd and dir is root-cmd/sub-cmd/sub-sub-cmd
// then the result should be .
func linkHandler(dir string) func(string) string {
	return func(s string) string {
		s = strings.Trim(s, ".md")
		s = strings.ReplaceAll(s, "_", "/")

		currentDir := strings.Split(dir, "/")
		linkDir := strings.Split(s, "/")

		if len(currentDir) > 0 && currentDir[0] == "." {
			currentDir = currentDir[1:]
		}

		// if the link is shorter than the dir, such as root-cmd_sub-cmd and dir is root-cmd/sub-cmd/sub-sub-cmd
		// then the result should be ../ (since it is relative to the current directory)
		if len(linkDir) < len(currentDir) {
			return strings.Repeat("../", len(currentDir)-len(linkDir))
		}

		// If the link is longer than the dir, such as root-cmd_sub-cmd_sub-sub-cmd and dir is root-cmd/sub-cmd
		// then the result should be ./sub-sub-cmd/
		if len(linkDir) > len(currentDir) {
			return "./" + strings.Join(linkDir[len(currentDir):], "/")
		}

		// If the link is the same as the dir, such as root-cmd_sub-cmd_sub-sub-cmd and dir is root-cmd/sub-cmd/sub-sub-cmd
		// then the result should be .
		return "."
	}
}

// docusaurusHeader is a page header for a doc page in a docusaurus site.
type docusaursHeader struct {
	sidebar_position int
	sidebar_label    string
	id               string
	title            string
	description      string
}

// String returns the header as a string.
func (d *docusaursHeader) String() string {
	return fmt.Sprintf(headerString, d.sidebar_position, d.sidebar_label, d.id, d.title, d.description)
}

var headerString = `---
sidebar_position: %d
sidebar_label: "%s"
id: "%s"
title: "%s"
description: "%s"
---`
