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

	"sigs.k8s.io/krew/pkg/receiptsmigration"

	"github.com/spf13/cobra"
)

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system receipts-upgrade",
	Short: "Perform krew maintenance tasks",
	Long: `Perform krew maintenance tasks such as migrating to a new krew-home layout.

This command will be removed without further notice from future versions of krew.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "receipts-upgrade" {
			return fmt.Errorf("only subcommand `receipts-upgrade` is supported")
		}
		return receiptsmigration.Migrate(paths)
	},
	PreRunE: ensureIndexUpdated,
	Hidden:  true,
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
