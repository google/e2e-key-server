// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/google/key-transparency/core/proto/ctmap"
)

var (
	start, end int64
)

// histCmd fetches the account history for a user
var histCmd = &cobra.Command{
	Use:   "history [user email]",
	Short: "Retrieve and verify all keys used for this account",
	Long: `Retrieve all user profiles for this account from the key server 
and verify that the results are consistent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("user email needs to be provided")
		}
		userID := args[0]
		timeout := viper.GetDuration("timeout")

		c, err := GetClient("")
		if err != nil {
			return fmt.Errorf("Error connecting: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if end == 0 {
			// Get the current epoch.
			_, smh, err := c.GetEntry(ctx, userID)
			if err != nil {
				return fmt.Errorf("GetEntry failed: %v", err)
			}
			if verbose {
				fmt.Printf("Got current epoch: %v\n", smh.Epoch)
			}
			end = smh.Epoch
		}

		profiles, err := c.ListHistory(ctx, userID, start, end)
		if err != nil {
			return fmt.Errorf("ListHistory failed: %v", err)
		}

		// Sort map heads.
		keys := make([]*ctmap.MapHead, 0, len(profiles))
		for k := range profiles {
			keys = append(keys, k)
		}
		sort.Sort(mapHeads(keys))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Epoch\tTimestamp\tProfile")
		for _, m := range keys {
			t, err := ptypes.Timestamp(m.IssueTime)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "%v\t%v\t%v\n", m.Epoch, t.Format(time.UnixDate), profiles[m])
		}
		if err := w.Flush(); err != nil {
			return nil
		}
		return nil
	},
}

// mapHeads satisfies sort.Interface to allow sorting []MapHead by epoch.
type mapHeads []*ctmap.MapHead

func (m mapHeads) Len() int           { return len(m) }
func (m mapHeads) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m mapHeads) Less(i, j int) bool { return m[i].Epoch < m[j].Epoch }

func init() {
	RootCmd.AddCommand(histCmd)

	histCmd.PersistentFlags().Int64Var(&start, "start", 0, "Start epoch")
	histCmd.PersistentFlags().Int64Var(&end, "end", 0, "End epoch")
}
