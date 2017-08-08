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

// +build mysql

package mutations

import (
	_ "github.com/go-sql-driver/mysql" // Set database engine.
)

var (
	createStmt = []string{
		`
	CREATE TABLE IF NOT EXISTS Maps (
		MapID   BIGINT NOT NULL,
		PRIMARY KEY(MapID)
	);`,
		`
	CREATE TABLE IF NOT EXISTS Mutations (
		MapID    BIGINT        NOT NULL,
		Sequence INTEGER       NOT NULL PRIMARY KEY AUTO_INCREMENT,
                MIndex   VARBINARY(32) NOT NULL,
		Mutation BLOB          NOT NULL,
		FOREIGN KEY(MapID) REFERENCES Maps(MapID) ON DELETE CASCADE
	);`,
	}
)
