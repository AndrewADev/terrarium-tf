// Package cmd
/*
Copyright © 2022 Robert Schönthal <robert@schoenthal.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/terrarium-tf/cli/lib"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func NewInitCommand(root *cobra.Command) {
	var initCmd = &cobra.Command{
		Use:   "init workspace stack [-state-lock=false]",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Args: lib.ArgsValidator,
		Run: func(cmd *cobra.Command, args []string) {
			tf, ctx := lib.Executor(*cmd, args[0], args[1])
			_, mergedVars := lib.Vars(*cmd, args[0], args[1])

			//init
			_ = tf.Init(ctx, buildInitOptions(*cmd, mergedVars, args)...)
		},
	}

	initCmd.Flags().BoolP("remote-state", "r", true, "initialize with remote state")

	// TODO allow more customizations for remote state
	initCmd.Flags().Bool("state-lock", true, "initialize with state locking")
	initCmd.Flags().String("state-bucket", "", "initialize with state bucket")
	initCmd.Flags().String("state-dynamo", "", "initialize with state dynamo for locking")
	initCmd.Flags().String("state-region", "", "initialize with state region")
	initCmd.Flags().String("state-account", "", "initialize with state aws account")
	initCmd.Flags().String("state-name", "", "initialize with state name")

	root.AddCommand(initCmd)
}

func buildInitOptions(cmd cobra.Command, mergedVars map[string]interface{}, args []string) []tfexec.InitOption {
	var opts []tfexec.InitOption

	// if we want to init with an s3/dynamo remote state
	rs := cmd.Flags().Lookup("remote-state")
	if rs.Value.String() == "true" {
		opts = append(opts,
			configureRegion(cmd, mergedVars),
			configureBucket(cmd, mergedVars),
			configureStateKey(cmd, mergedVars, args),
		)

		sl := cmd.Flags().Lookup("state-lock")
		if sl.Value.String() == "true" {
			opts = append(opts,
				configureStateLock(cmd, mergedVars),
			)
		}
	} else {
		opts = append(opts, tfexec.Backend(false))
	}

	return append(opts, tfexec.Upgrade(true))
}

func configureRegion(cmd cobra.Command, mergedVars map[string]interface{}) tfexec.InitOption {
	region := lib.GetVar("region", cmd, mergedVars, false)

	if region == "" {
		if os.Getenv("AWS_REGION") != "" {
			region = os.Getenv("AWS_REGION")
		}
		if region == "" && os.Getenv("AWS_DEFAULT_REGION") != "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}
	}

	if region == "" {
		log.Fatalf(lib.ErrorColorLine, "unable to configure remote state, 'region' was not found in var files and not provided with '-state-region' nor was AWS_REGION or AWS_DEFAULT_REGION found in global environment")
	}

	return tfexec.BackendConfig(fmt.Sprintf("region=%s", region))
}

func configureBucket(cmd cobra.Command, mergedVars map[string]interface{}) tfexec.InitOption {
	bucket := lib.GetVar("bucket", cmd, mergedVars, false)
	if bucket == "" {
		// no bucket defined, so generate a unique name
		bucket = fmt.Sprintf("tf-state-%s-%s-%s",
			lib.GetVar("project", cmd, mergedVars, false),
			lib.GetVar("region", cmd, mergedVars, false),
			lib.GetVar("account", cmd, mergedVars, true),
		)
	}
	return tfexec.BackendConfig(fmt.Sprintf("bucket=%s", bucket))
}

func configureStateKey(cmd cobra.Command, mergedVars map[string]interface{}, args []string) tfexec.InitOption {
	key := lib.GetVar("name", cmd, mergedVars, false)
	if key == "" {
		// no bucket defined, so generate a unique name
		key = path.Base(args[1])
	}
	return tfexec.BackendConfig(fmt.Sprintf("key=%s.tfstate", key))
}

func configureStateLock(cmd cobra.Command, mergedVars map[string]interface{}) tfexec.InitOption {
	table := lib.GetVar("dynamo", cmd, mergedVars, false)
	if table == "" {
		// no bucket defined, so generate a unique name
		table = fmt.Sprintf("terraform-lock-%s-%s-%s",
			lib.GetVar("project", cmd, mergedVars, false),
			lib.GetVar("region", cmd, mergedVars, false),
			lib.GetVar("account", cmd, mergedVars, true),
		)
	}
	return tfexec.BackendConfig(fmt.Sprintf("dynamodb_table=%s", table))
}
