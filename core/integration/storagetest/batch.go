package storagetest

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/keytransparency/core/sequencer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spb "github.com/google/keytransparency/core/sequencer/sequencer_go_proto"
)

// Batcher writes batch definitions to storage.
type Batcher = sequencer.Batcher

type BatchStorageFactory func(ctx context.Context, t *testing.T) Batcher

type BatchStorageTest func(ctx context.Context, t *testing.T, b Batcher)

// RunBatchStorageTests runs all the batch storage tests against the provided map storage implementation.
func RunBatchStorageTests(t *testing.T, storageFactory BatchStorageFactory) {
	ctx := context.Background()
	b := &BatchTests{}
	for name, f := range map[string]BatchStorageTest{
		// TODO(gbelvin): Discover test methods via reflection.
		"TestNotFound":   b.TestNotFound,
		"TestWriteBatch": b.TestWriteBatch,
		"TestReadBatch":  b.TestReadBatch,
		"TestHighestRev": b.TestHighestRev,
	} {
		ms := storageFactory(ctx, t)
		t.Run(name, func(t *testing.T) { f(ctx, t, ms) })
	}
}

// BatchTests is a suite of tests to run against
type BatchTests struct{}

func (*BatchTests) TestNotFound(ctx context.Context, t *testing.T, b Batcher) {
	_, err := b.ReadBatch(ctx, "nodir", 0)
	st := status.Convert(err)
	if got, want := st.Code(), codes.NotFound; got != want {
		t.Errorf("ReadBatch(): %v, want %v", err, want)
	}
}

func (*BatchTests) TestWriteBatch(ctx context.Context, t *testing.T, b Batcher) {
	domainID := "writebatchtest"
	for _, tc := range []struct {
		rev     int64
		wantErr bool
		sources []*spb.MapMetadata_SourceSlice
	}{
		// Tests are cumulative.
		{rev: 0, sources: []*spb.MapMetadata_SourceSlice{{LogId: 1, HighestExclusive: 11}}},
		{rev: 0, sources: []*spb.MapMetadata_SourceSlice{{LogId: 1, HighestExclusive: 12}}, wantErr: true},
		{rev: 0, sources: []*spb.MapMetadata_SourceSlice{{LogId: 1, HighestExclusive: 21}}, wantErr: true},
		{rev: 0, sources: []*spb.MapMetadata_SourceSlice{}, wantErr: true},
		{rev: 1, sources: []*spb.MapMetadata_SourceSlice{}},
		{rev: 1, sources: []*spb.MapMetadata_SourceSlice{{LogId: 1, HighestExclusive: 11}}, wantErr: true},
	} {
		err := b.WriteBatchSources(ctx, domainID, tc.rev, &spb.MapMetadata{Sources: tc.sources})
		if got, want := err != nil, tc.wantErr; got != want {
			t.Errorf("WriteBatchSources(%v, %v): err: %v. code: %v, want %v",
				tc.rev, tc.sources, err, got, want)
		}
	}
}

func (*BatchTests) TestReadBatch(ctx context.Context, t *testing.T, b Batcher) {
	domainID := "readbatchtest"
	for _, tc := range []struct {
		rev  int64
		want *spb.MapMetadata
	}{
		{rev: 0, want: &spb.MapMetadata{Sources: []*spb.MapMetadata_SourceSlice{
			{LogId: 1, HighestExclusive: 11},
			{LogId: 2, HighestExclusive: 21},
		}}},
		{rev: 1, want: &spb.MapMetadata{Sources: []*spb.MapMetadata_SourceSlice{
			{LogId: 1, HighestExclusive: 12},
			{LogId: 2, HighestExclusive: 23},
		}}},
	} {
		if err := b.WriteBatchSources(ctx, domainID, tc.rev, tc.want); err != nil {
			t.Fatalf("WriteBatch(%v): %v", tc.rev, err)
		}
		got, err := b.ReadBatch(ctx, domainID, tc.rev)
		if err != nil {
			t.Fatalf("ReadBatch(%v): %v", tc.rev, err)
		}
		if !cmp.Equal(got, tc.want, cmp.Comparer(proto.Equal)) {
			t.Errorf("ReadBatch(%v): %v, want %v", tc.rev, got, tc.want)
		}
	}
}

func (*BatchTests) TestHighestRev(ctx context.Context, t *testing.T, b Batcher) {
	domainID := "writebatchtest"
	for _, tc := range []struct {
		rev     int64
		sources []*spb.MapMetadata_SourceSlice
	}{
		// Tests are cumulative.
		{rev: 0, sources: []*spb.MapMetadata_SourceSlice{{LogId: 1, HighestExclusive: 11}}},
		{rev: 1, sources: []*spb.MapMetadata_SourceSlice{}},
	} {
		err := b.WriteBatchSources(ctx, domainID, tc.rev, &spb.MapMetadata{Sources: tc.sources})
		if err != nil {
			t.Errorf("WriteBatchSources(%v, %v): err: %v", tc.rev, tc.sources, err)
		}
		got, err := b.HighestRev(ctx, domainID)
		if err != nil {
			t.Errorf("HighestRev(): %v", err)
		}
		if got != tc.rev {
			t.Errorf("HighestRev(): %v, want %v", got, tc.rev)
		}
	}
}
