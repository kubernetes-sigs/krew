// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"

	"sigs.k8s.io/krew/pkg/installation"
)

type Entry struct {
	Name    string
	Version string
}

func sortByName(e []Entry) []Entry {
	sort.Slice(e, func(a, b int) bool {
		return e[a].Name < e[b].Name
	})
	return e
}

// Consume produces a junk GroupVersionKin for obj.GetObjectKind().GroupVersionKind().Empty() check to eat in PrintObj()
type Consume struct {
	APIVersion string
	Kind       string
}

func (c Consume) SetGroupVersionKind(kind schema.GroupVersionKind) {
	c.APIVersion, c.Kind = kind.ToAPIVersionAndKind()
}

func (c Consume) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(c.APIVersion, c.Kind)
}

func (e Entry) GetObjectKind() schema.ObjectKind {
	return Consume{"_", "_"}
}

func (e Entry) DeepCopyObject() runtime.Object {
	newObj := Entry{e.Name, e.Version}
	return newObj
}

type ListFlags struct {
	JSONYamlPrintFlags *genericclioptions.JSONYamlPrintFlags
	OutputFormat       *string
}

func NewListFlags() *ListFlags {
	outputFormat := ""
	return &ListFlags{
		JSONYamlPrintFlags: genericclioptions.NewJSONYamlPrintFlags(),
		OutputFormat:       &outputFormat,
	}
}

func (f *ListFlags) AllowedFormats() []string {
	formats := f.JSONYamlPrintFlags.AllowedFormats()
	return formats
}

func (f *ListFlags) AddFlags(c *cobra.Command) {}

func (f *ListFlags) ToPrinter() (printers.ResourcePrinter, error) {
	outputFormat := *f.OutputFormat
	var printer printers.ResourcePrinter

	switch outputFormat {
	case "json":
		printer = &printers.JSONPrinter{}
	case "yaml":
		printer = &printers.YAMLPrinter{}
	default:
		return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: f.OutputFormat, AllowedFormats: f.AllowedFormats()}
	}

	return printer, nil
}

func init() {
	outputFormat := ""
	output := NewListFlags()

	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list [(-o|--output=)json|yaml|wide]",
		Short: "List installed kubectl plugins",
		Long: `Show a list of installed kubectl plugins and their versions.

Remarks:
  Redirecting the output of this command to a program or file will only print
  the names of the plugins installed. This output can be piped back to the
  "install" command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}

			// return sorted list of plugin names when piped to other commands or file
			if *output.OutputFormat == "" && !isTerminal(os.Stdout) {
				var names []string
				for name := range plugins {
					names = append(names, name)
				}
				sort.Strings(names)
				fmt.Fprintln(os.Stdout, strings.Join(names, "\n"))
				return nil
			}

			if *output.OutputFormat == "wide" {
				// print table
				var rows [][]string
				for p, version := range plugins {
					rows = append(rows, []string{p, version})
				}
				rows = sortByFirstColumn(rows)
				return printTable(os.Stdout, []string{"PLUGIN", "VERSION"}, rows)
			}

			if inArray(*output.OutputFormat, output.AllowedFormats()) {
				objs := []Entry{}
				for plugin, version := range plugins {
					obj := Entry{plugin, version}
					objs = append(objs, obj)
				}
				objs = sortByName(objs)
				for _, obj := range objs {
					p, err := output.ToPrinter()
					if err != nil {
						return err
					}
					err = p.PrintObj(obj, os.Stdout)
					if err != nil {
						return err
					}
				}
				return nil
			}
			return errors.New("unsupported output format specified")
		},
		PreRunE: checkIndex,
	}

	listCmd.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "_")

	output.OutputFormat = &outputFormat
	rootCmd.AddCommand(listCmd)
}

func inArray(s string, arr []string) bool {
	for _, a := range arr {
		if s == a {
			return true
		}
	}
	return false
}

func printTable(out io.Writer, columns []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprint(w, strings.Join(columns, "\t"))
	fmt.Fprintln(w)
	for _, values := range rows {
		fmt.Fprint(w, strings.Join(values, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func sortByFirstColumn(rows [][]string) [][]string {
	sort.Slice(rows, func(a, b int) bool {
		return rows[a][0] < rows[b][0]
	})
	return rows
}
