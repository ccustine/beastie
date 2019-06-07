// Copyright © 2018 Chris Custine <ccustine@apache.org>
//
// Licensed under the Apache License, version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registrycmd

import (
	"github.com/ccustine/beastie/config"
	"github.com/spf13/cobra"
)

var info *config.BeastInfo

func NewImportCmd(beastInfo *config.BeastInfo) *cobra.Command {
	info = beastInfo
	cmd := &cobra.Command{
		Use:   "import",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
	cmd.AddCommand(NewImportRegistryCmd())
	return cmd
}

func NewFindCmd(beastInfo *config.BeastInfo) *cobra.Command {
	info = beastInfo
	cmd := &cobra.Command{
		Use:   "find",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
	cmd.AddCommand(NewFindRegistryCmd())
	return cmd
}

func NewListCmd(beastInfo *config.BeastInfo) *cobra.Command {
	info = beastInfo
	cmd := &cobra.Command{
		Use:   "list",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
	cmd.AddCommand(NewListRegistryCmd())
	return cmd
}

func NewDownloadCmd(beastInfo *config.BeastInfo) *cobra.Command {
	info = beastInfo
	cmd := &cobra.Command{
		Use:   "download",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
	cmd.AddCommand(NewDownloadRegistryCmd())
	return cmd
}
