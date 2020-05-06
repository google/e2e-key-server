// Copyright 2020 Google Inc. All Rights Reserved.
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

package spanner

import (
	"cloud.google.com/go/spanner/spansql"
)

//go:generate sh -c "echo '// Code generated by impl/spanner/ddl.go. DO NOT EDIT.' > keytransparency.ddl.go"
//go:generate sh -c "echo '\npackage spanner\n' >> keytransparency.ddl.go"
//go:generate sh -c "echo 'const ddlString = `' >> keytransparency.ddl.go"
//go:generate sh -c "cat ./keytransparency.ddl >> keytransparency.ddl.go"
//go:generate sh -c "echo '`' >> keytransparency.ddl.go"

// ReadDDL returns a list of DDL statements for the KT database schema
func ReadDDL() ([]string, error) {
	ddl, err := spansql.ParseDDL(ddlString)
	if err != nil {
		return nil, err
	}
	stmts := make([]string, 0, len(ddl.List))
	for _, s := range ddl.List {
		stmts = append(stmts, s.SQL())
	}
	return stmts, nil
}
