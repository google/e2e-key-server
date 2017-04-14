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

package appender

import (
	"github.com/google/keytransparency/core/transaction"

	"golang.org/x/net/context"
)

// Appender is an append only interface into a data structure.
type Appender interface {
	// Adds an object to the append-only data structure.
	Append(ctx context.Context, txn transaction.Txn, epoch int64, obj interface{}) error

	// Epoch retrieves a specific object.
	// Returns obj and a serialized ct.SignedCertificateTimestamp
	Epoch(ctx context.Context, epoch int64, obj interface{}) ([]byte, error)

	// Latest returns the latest object.
	// Returns epoch, obj, and a serialized ct.SignedCertificateTimestamp
	Latest(ctx context.Context, obj interface{}) (int64, []byte, error)
}

// Local stores a list of items that have been sequenced.
type Local interface {
	// Write writes an object at a given epoch.
	Write(txn transaction.Txn, logID, epoch int64, obj interface{}) error

	// Read retrieves a specific object at a given epoch.
	Read(txn transaction.Txn, logID, epoch int64, obj interface{}) error

	// Latest returns the latest object and its epoch.
	Latest(txn transaction.Txn, logID int64, obj interface{}) (int64, error)
}

// Remote stores a list of items in a remote service.
type Remote interface {
	// Write writes an object at a given epoch.
	Write(ctx context.Context, logID, epoch int64, obj interface{}) error

	// Read retrieves a specific object at a given epoch.
	Read(ctx context.Context, logID, epoch int64, obj interface{}) error

	// Latest returns the latest object and its epoch.
	Latest(ctx context.Context, logID int64, obj interface{}) (int64, error)
}
