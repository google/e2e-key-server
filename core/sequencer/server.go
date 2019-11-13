// Copyright 2018 Google Inc. All Rights Reserved.
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

package sequencer

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/trillian/monitoring"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/keytransparency/core/directory"
	"github.com/google/keytransparency/core/mutator"
	"github.com/google/keytransparency/core/mutator/entry"
	"github.com/google/keytransparency/core/sequencer/mapper"
	"github.com/google/keytransparency/core/sequencer/metadata"
	"github.com/google/keytransparency/core/sequencer/runner"

	spb "github.com/google/keytransparency/core/sequencer/sequencer_go_proto"
	tpb "github.com/google/trillian"
)

const (
	directoryIDLabel = "directoryid"
	logIDLabel       = "logid"
	reasonLabel      = "reason"
	fnLabel          = "fn"
)

var (
	initMetrics        sync.Once
	knownDirectories   monitoring.Gauge
	logEntryCount      monitoring.Counter
	logEntryUnapplied  monitoring.Gauge
	mapLeafCount       monitoring.Counter
	fnCount            monitoring.Counter
	mapRevisionCount   monitoring.Counter
	watermarkWritten   monitoring.Gauge
	watermarkDefined   monitoring.Gauge
	watermarkApplied   monitoring.Gauge
	mutationFailures   monitoring.Counter
	fnLatency          monitoring.Histogram
	logRootTrail       monitoring.Gauge
	unappliedRevisions monitoring.Gauge
)

// zero is the time to use for the beginning edge of new watermarks.
// When getting the watermark for a new log that is not in the lastMeta set,
// we need to pick a lower bound starting point for reading.  time.Time{} is
// not a valid choice here because it is too low to represent with metadata.New()
// so we need to pick a different, low sentinel value that the tests know about.
var zero = time.Unix(0, 0)

func createMetrics(mf monitoring.MetricFactory) {
	knownDirectories = mf.NewGauge(
		"known_directories",
		"Set to 1 for known directories (whether this instance is master or not)",
		directoryIDLabel)
	logEntryCount = mf.NewCounter(
		"log_entry_count",
		"Total number of log entries read since process start. Duplicates are not removed.",
		directoryIDLabel, logIDLabel)
	logEntryUnapplied = mf.NewGauge(
		"log_entry_unapplied",
		"Total number of log entries still to be processed in the queue.",
		directoryIDLabel)
	mapLeafCount = mf.NewCounter(
		"map_leaf_count",
		"Total number of map leaves written since process start. Duplicates are not removed.",
		directoryIDLabel)
	fnCount = mf.NewCounter(
		"fn_count",
		"Total number of mapping operations that have run since process start",
		directoryIDLabel, fnLabel)
	mapRevisionCount = mf.NewCounter(
		"map_revision_count",
		"Total number of map revisions written since process start.",
		directoryIDLabel)
	watermarkWritten = mf.NewGauge(
		"watermark_written",
		"High watermark of each input log that has been written",
		directoryIDLabel, logIDLabel)
	watermarkDefined = mf.NewGauge(
		"watermark_defined",
		"High watermark of each input log that has been defined in the batch table",
		directoryIDLabel, logIDLabel)
	watermarkApplied = mf.NewGauge(
		"watermark_applied",
		"High watermark of each input log that has been committed in a map revision",
		directoryIDLabel, logIDLabel)
	mutationFailures = mf.NewCounter(
		"mutation_failures",
		"Number of invalid mutations the signer has processed for directoryid since process start",
		directoryIDLabel, reasonLabel)
	fnLatency = mf.NewHistogram(
		"apply_revision_latency",
		"Latency of sequencer apply revision operation in seconds",
		directoryIDLabel, fnLabel)
	logRootTrail = mf.NewGauge(
		"log_root_trail",
		"How many revisions have not been published to the log",
	)
	unappliedRevisions = mf.NewGauge(
		"unapplied_revisions",
		"How many revisions have been defined but haven't been applied to the map",
	)
}

// Watermarks is a map of watermarks by logID.
type Watermarks map[int64]int64

// LogsReader reads messages in multiple logs.
type LogsReader interface {
	// HighWatermark returns the number of items and the highest primary
	// key up to batchSize items after start (exclusive).
	HighWatermark(ctx context.Context, directoryID string, logID int64, start time.Time,
		batchSize int32) (count int32, watermark time.Time, err error)

	// ListLogs returns the logIDs associated with directoryID that have their write bits set,
	// or all logIDs associated with directoryID if writable is false.
	ListLogs(ctx context.Context, directoryID string, writable bool) ([]int64, error)

	// ReadLog returns the lowest messages in the [low, high) range stored in the specified log, up to batchSize.
	ReadLog(ctx context.Context, directoryID string, logID int64, low, high time.Time,
		batchSize int32) ([]*mutator.LogMessage, error)
}

// Batcher writes batch definitions to storage.
type Batcher interface {
	// WriteBatchSources saves the (low, high] boundaries used for each log in making this revision.
	WriteBatchSources(ctx context.Context, dirID string, rev int64, meta *spb.MapMetadata) error
	// ReadBatch returns the batch definitions for a given revision.
	ReadBatch(ctx context.Context, directoryID string, rev int64) (*spb.MapMetadata, error)
	// HighestRev returns the highest defined revision number for directoryID.
	HighestRev(ctx context.Context, directoryID string) (int64, error)
}

// Server implements KeyTransparencySequencerServer.
type Server struct {
	directories         directory.Storage
	batcher             Batcher
	trillian            trillianFactory
	logs                LogsReader
	loopback            spb.KeyTransparencySequencerClient
	BatchSize           int32
	LogPublishBatchSize uint64
}

// NewServer creates a new KeyTransparencySequencerServer.
func NewServer(
	directories directory.Storage,
	tlog tpb.TrillianLogClient,
	tmap tpb.TrillianMapClient,
	twrite tpb.TrillianMapWriteClient,
	batcher Batcher,
	logs LogsReader,
	loopback spb.KeyTransparencySequencerClient,
	metricsFactory monitoring.MetricFactory,
) *Server {
	initMetrics.Do(func() { createMetrics(metricsFactory) })
	return &Server{
		directories: directories,
		trillian: &Trillian{
			directories: directories,
			tmap:        tmap,
			tlog:        tlog,
			twrite:      twrite,
		},
		batcher:             batcher,
		logs:                logs,
		loopback:            loopback,
		BatchSize:           10000,
		LogPublishBatchSize: 10,
	}
}

func (s *Server) UpdateMetrics(ctx context.Context, in *spb.UpdateMetricsRequest) (*spb.UpdateMetricsResponse, error) {
	if err := s.unappliedMetric(ctx, in.DirectoryId, in.MaxUnappliedCount); err != nil {
		glog.Errorf("unappliedMetric(%v): %v", in.DirectoryId, err)
		return nil, err
	}
	return &spb.UpdateMetricsResponse{}, nil
}

// unappliedMetric updates the log_entryunapplied metric for directoryID
func (s *Server) unappliedMetric(ctx context.Context, directoryID string, maxCount int32) error {
	// Get the previous and current high water marks.
	mapClient, err := s.trillian.MapClient(ctx, directoryID)
	if err != nil {
		return err
	}
	_, latestMapRoot, err := mapClient.GetAndVerifyLatestMapRoot(ctx)
	if err != nil {
		return err
	}
	var lastMeta spb.MapMetadata
	if err := proto.Unmarshal(latestMapRoot.Metadata, &lastMeta); err != nil {
		return err
	}
	// Query metadata about outstanding log items.
	count, meta, err := s.HighWatermarks(ctx, directoryID, &lastMeta, maxCount)
	if err != nil {
		return status.Errorf(codes.Internal, "HighWatermarks(): %v", err)
	}
	logEntryUnapplied.Set(float64(count), directoryID)
	for _, source := range meta.Sources {
		watermarkWritten.Set(float64(source.HighestExclusive), directoryID, fmt.Sprintf("%v", source.LogId))
	}
	return nil
}

// DefineRevisions returns the set of outstanding revisions that have not been
// applied, after optionally defining a new revision of outstanding mutations.
func (s *Server) DefineRevisions(ctx context.Context,
	in *spb.DefineRevisionsRequest) (*spb.DefineRevisionsResponse, error) {
	revs, err := s.GetDefinedRevisions(ctx,
		&spb.GetDefinedRevisionsRequest{DirectoryId: in.DirectoryId})
	if err != nil {
		return nil, err
	}

	resp := &spb.DefineRevisionsResponse{
		HighestApplied: revs.HighestApplied,
		HighestDefined: revs.HighestDefined,
	}
	// Do nothing if the highest defined revision is lagging behind for some
	// reason. It will catch up later.
	if resp.HighestDefined < resp.HighestApplied {
		return resp, nil
	}
	// Allow at most MaxUnapplied pending revisions. Having MaxUnapplied != 0
	// enables applying a revision and defining the next one simultaneously.
	if resp.HighestDefined > resp.HighestApplied+int64(in.MaxUnapplied) {
		return resp, nil
	}

	// Query metadata about outstanding log items.
	lastMeta, err := s.batcher.ReadBatch(ctx, in.DirectoryId, resp.HighestDefined)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ReadBatch(): %v", err)
	}
	// Advance the watermarks forward, to define a new batch.
	count, meta, err := s.HighWatermarks(ctx, in.DirectoryId, lastMeta, in.MaxBatch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "HighWatermarks(): %v", err)
	}

	//
	// Rate limit the creation of new batches.
	//

	// TODO(#1057): If time since last map revision > max timeout, define batch.
	// TODO(#1047): If time since oldest queue item > max latency has elapsed, define batch.
	// If count items >= min_batch, define batch.
	if count >= in.MinBatch {
		resp.HighestDefined++
		nextRev := resp.HighestDefined
		if err := s.batcher.WriteBatchSources(ctx, in.DirectoryId, nextRev, meta); err != nil {
			return nil, status.Errorf(codes.Internal, "WriteBatchSources(): %v", err)
		}
		for _, source := range meta.Sources {
			watermarkDefined.Set(float64(source.HighestExclusive),
				in.DirectoryId, fmt.Sprintf("%v", source.LogId))
		}
	}
	// TODO(#1056): If count items == max_batch, should we define the next batch immediately?

	return resp, nil
}

// GetDefinedRevisions returns the range of defined and unapplied revisions.
func (s *Server) GetDefinedRevisions(ctx context.Context,
	in *spb.GetDefinedRevisionsRequest) (*spb.GetDefinedRevisionsResponse, error) {
	// Get the last processed revision number.
	mapClient, err := s.trillian.MapClient(ctx, in.DirectoryId)
	if err != nil {
		return nil, err
	}
	_, root, err := mapClient.GetAndVerifyLatestMapRoot(ctx)
	if err != nil {
		return nil, err
	}
	// Get the highest defined revision number.
	// TODO(pavelkalinnikov): Run this in parallel with getting the root.
	rev, err := s.batcher.HighestRev(ctx, in.DirectoryId)
	if err != nil {
		return nil, err
	}
	return &spb.GetDefinedRevisionsResponse{
		HighestApplied: int64(root.Revision),
		HighestDefined: rev,
	}, nil
}

// ApplyRevisions builds multiple outstanding revisions of a single directory's
// map by integrating the corresponding mutations.
func (s *Server) ApplyRevisions(ctx context.Context, in *spb.ApplyRevisionsRequest) (*empty.Empty, error) {
	resp, err := s.loopback.GetDefinedRevisions(ctx,
		&spb.GetDefinedRevisionsRequest{DirectoryId: in.DirectoryId})
	if err != nil {
		return nil, err
	}

	cnt := resp.HighestDefined - resp.HighestApplied
	unappliedRevisions.Set(float64(cnt))
	if cnt < 0 {
		cnt = 0
	}

	if mx := int64(s.LogPublishBatchSize); cnt > mx {
		glog.Warningf("ApplyRevisions: too many outstanding revisions: %d; applying %d", cnt, mx)
		cnt = mx
	}
	glog.Infof("Applying revisions %d-%d", resp.HighestApplied+1, resp.HighestApplied+cnt)
	for i := int64(1); i <= cnt; i++ {
		req := &spb.ApplyRevisionRequest{
			DirectoryId: in.DirectoryId, Revision: resp.HighestApplied + i}
		if _, err := s.loopback.ApplyRevision(ctx, req); err != nil {
			return nil, err
		}
	}
	return &empty.Empty{}, nil
}

// readMessages returns the full set of EntryUpdates defined by sources.
// chunkSize limits the number of messages to read from a log at one time.
func (s *Server) readMessages(ctx context.Context, source *spb.MapMetadata_SourceSlice,
	directoryID string, chunkSize int32,
	emit func(*mutator.LogMessage)) error {
	ss := metadata.FromProto(source)
	low := ss.StartTime()
	high := ss.EndTime()
	for moreToRead := true; moreToRead; {
		// Request one more item than chunkSize so we can find the next page token.
		batch, err := s.logs.ReadLog(ctx, directoryID, source.LogId, low, high, chunkSize+1)
		if err != nil {
			return fmt.Errorf("logs.ReadLog(): %v", err)
		}
		glog.Infof("ReadLog(dir: %v log: %v, (%v, %v], %v) count: %v",
			directoryID, source.LogId, low, high, chunkSize, len(batch))
		moreToRead = int32(len(batch)) == (chunkSize + 1)
		if moreToRead {
			low = batch[chunkSize].ID // Use the last row as the start of the next read.
			batch = batch[:chunkSize] // Don't emit the next page token.
		}
		logEntryCount.Add(float64(len(batch)), directoryID, fmt.Sprintf("%v", source.LogId))
		for _, m := range batch {
			emit(m)
		}
	}
	return nil
}

// ApplyRevision applies the supplied mutations to the current map revision and creates a new revision.
func (s *Server) ApplyRevision(ctx context.Context, in *spb.ApplyRevisionRequest) (*spb.ApplyRevisionResponse, error) {
	start := time.Now()
	defer func() { fnLatency.Observe(time.Since(start).Seconds(), in.DirectoryId, "ApplyRevision") }()
	meta, err := s.batcher.ReadBatch(ctx, in.DirectoryId, in.Revision)
	fnLatency.Observe(time.Since(start).Seconds(), in.DirectoryId, "ReadBatch")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ReadBatch(%v, %v): %v", in.DirectoryId, in.Revision, err)
	}
	glog.Infof("ApplyRevision(): dir: %v, rev: %v, sources: %v", in.DirectoryId, in.Revision, meta)

	incMetricFn := func(label string) { fnCount.Inc(in.DirectoryId, label) }

	logSlices := runner.DoMapMetaFn(mapper.MapMetaFn, meta, incMetricFn)
	logItems, err := runner.DoReadFn(ctx, s.readMessages, logSlices, in.DirectoryId, s.BatchSize, incMetricFn)
	if err != nil {
		mutationFailures.Inc(err.Error())
		return nil, err
	}

	emitErrFn := func(err error) {
		glog.Warning(err)
		mutationFailures.Inc(in.DirectoryId, status.Code(err).String())
	}
	// Map Log Items
	indexedValues := runner.DoMapLogItemsFn(entry.MapLogItemFn, logItems, emitErrFn, incMetricFn)

	// Collect Indexes.
	groupByIndex := make(map[string]bool)
	for _, iv := range indexedValues {
		groupByIndex[string(iv.Index)] = true
	}
	indexes := make([][]byte, 0, len(groupByIndex))
	for i := range groupByIndex {
		indexes = append(indexes, []byte(i))
	}

	// Read Map.
	mapClient, err := s.trillian.MapWriteClient(ctx, in.DirectoryId)
	if err != nil {
		return nil, err
	}
	verifyLeafStart := time.Now()
	leaves, err := mapClient.GetLeavesByRevision(ctx, in.Revision-1, indexes)
	fnLatency.Observe(time.Since(verifyLeafStart).Seconds(), in.DirectoryId, "GetLeavesByRevision")
	if err != nil {
		return nil, err
	}

	computeStart := time.Now()
	// Convert Trillian map leaves into indexed KT updates.
	indexedLeaves, err := runner.DoMapMapLeafFn(mapper.MapMapLeafFn, leaves, incMetricFn)
	if err != nil {
		return nil, err
	}

	// GroupByIndex.
	joined := runner.Join(indexedLeaves, indexedValues, incMetricFn)

	// Apply mutations to values.
	newIndexedLeaves := runner.DoReduceFn(entry.ReduceFn, joined, emitErrFn, incMetricFn)
	glog.V(2).Infof("DoReduceFn reduced %v values on %v indexes", len(indexedValues), len(joined))

	// Marshal new indexed values back into Trillian Map leaves.
	newLeaves := runner.DoMarshalIndexedValues(newIndexedLeaves, emitErrFn, incMetricFn)
	fnLatency.Observe(time.Since(computeStart).Seconds(), in.DirectoryId, "ProcessMutations")

	// Serialize metadata
	metadata, err := proto.Marshal(meta)
	if err != nil {
		return nil, err
	}

	// Set new leaf values.
	setRevisionStart := time.Now()
	err = mapClient.WriteLeaves(ctx, in.Revision, newLeaves, metadata)
	fnLatency.Observe(time.Since(setRevisionStart).Seconds(), in.DirectoryId, "WriteLeaves")
	if err != nil {
		return nil, err
	}
	glog.V(2).Infof("CreateRevision: WriteLeaves:{Revision: %v}", in.Revision)

	for _, s := range meta.Sources {
		watermarkApplied.Set(float64(s.HighestExclusive), in.DirectoryId, fmt.Sprintf("%v", s.LogId))
	}
	mapLeafCount.Add(float64(len(newLeaves)), in.DirectoryId)
	mapRevisionCount.Inc(in.DirectoryId)
	glog.Infof("ApplyRevision(): dir: %v, rev: %v, mutations: %v, indexes: %v, newleaves: %v",
		in.DirectoryId, in.Revision, len(logItems), len(indexes), len(newLeaves))
	return &spb.ApplyRevisionResponse{
		DirectoryId: in.DirectoryId,
		Revision:    in.Revision,
		Mutations:   int64(len(indexedValues)),
		MapLeaves:   int64(len(newLeaves)),
	}, nil
}

// PublishRevisions copies the MapRoots of all known map revisions into the Log of MapRoots.
func (s *Server) PublishRevisions(ctx context.Context,
	in *spb.PublishRevisionsRequest) (*spb.PublishRevisionsResponse, error) {
	// Create verifying log and map clients.
	logClient, err := s.trillian.LogClient(ctx, in.DirectoryId)
	if err != nil {
		return nil, err
	}
	mapClient, err := s.trillian.MapClient(ctx, in.DirectoryId)
	if err != nil {
		return nil, err
	}

	// Get latest log root and map root.
	logRoot, err := logClient.UpdateRoot(ctx)
	if err != nil {
		return nil, err
	}
	latestRawMapRoot, latestMapRoot, err := mapClient.GetAndVerifyLatestMapRoot(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetAndVerifyLatestMapRoot(): %v", err)
	}

	// Add all unpublished map roots to the log.
	revs := []int64{}
	leaves := make(map[int64][]byte)

	end := latestMapRoot.Revision
	if batch := logRoot.TreeSize + s.LogPublishBatchSize; batch < end {
		// Only publish up to LogPublishBatchSize log roots at a time.
		// TODO: add a metric for delta between log and map roots.
		glog.Errorf("PublishRevisions has too many revisions to catch up on: %d", latestMapRoot.Revision-logRoot.TreeSize)
		end = batch
	}
	logRootTrail.Set(float64(latestMapRoot.Revision - logRoot.TreeSize))
	for rev := logRoot.TreeSize - 1; rev <= end; rev++ {
		rawMapRoot, mapRoot, err := mapClient.GetAndVerifyMapRootByRevision(ctx, int64(rev))
		if err != nil {
			return nil, err
		}
		leaves[int64(mapRoot.Revision)] = rawMapRoot.GetMapRoot()
		revs = append(revs, int64(mapRoot.Revision))
	}
	glog.Infof("Publishing revisions %d-%d", logRoot.TreeSize-1, end)
	if err := logClient.AddSequencedLeaves(ctx, leaves); err != nil {
		glog.Errorf("AddSequencedLeaves(revs: %v): %v", revs, err)
		return nil, err
	}

	if in.Block {
		if err := logClient.WaitForInclusion(ctx, latestRawMapRoot.GetMapRoot()); err != nil {
			return nil, status.Errorf(codes.Internal, "WaitForInclusion(): %v", err)
		}
	}
	return &spb.PublishRevisionsResponse{Revisions: revs}, nil
}

// HighWatermarks returns the total count across all logs and the highest watermark for each log.
// batchSize is a limit on the total number of items represented by the returned watermarks.
// TODO(gbelvin): Block until a minBatchSize has been reached or a timeout has occurred.
func (s *Server) HighWatermarks(ctx context.Context, directoryID string, lastMeta *spb.MapMetadata,
	batchSize int32) (int32, *spb.MapMetadata, error) {
	var total int32

	// Ensure that we do not lose track of end watermarks, even if they are no
	// longer in the active log list, or if they do not move. The sequencer
	// needs them to know where to pick up reading for the next map
	// revision.
	// TODO(gbelvin): Separate end watermarks for the sequencer's needs
	// from ranges of watermarks for the verifier's needs.
	ends := make(map[int64]time.Time)
	starts := make(map[int64]time.Time)
	for _, source := range lastMeta.GetSources() {
		highest := metadata.FromProto(source).EndTime()
		if ends[source.LogId].Before(highest) {
			ends[source.LogId] = highest
			starts[source.LogId] = highest
		}
	}

	filterForWritable := false
	logIDs, err := s.logs.ListLogs(ctx, directoryID, filterForWritable)
	if err != nil {
		return 0, nil, err
	}
	// TODO(gbelvin): Get HighWatermarks in parallel.
	for _, logID := range logIDs {
		low, ok := ends[logID]
		if !ok {
			low = zero // Here be dragons. See comment on zero.
		}
		count, high, err := s.logs.HighWatermark(ctx, directoryID, logID, low, batchSize)
		if err != nil {
			return 0, nil, status.Errorf(codes.Internal,
				"HighWatermark(%v/%v, start: %v, batch: %v): %v",
				directoryID, logID, low, batchSize, err)
		}
		starts[logID], ends[logID] = low, high
		total += count
		batchSize -= count
	}

	meta := &spb.MapMetadata{}
	for logID, end := range ends {
		src, err := metadata.New(logID, starts[logID], end)
		if err != nil {
			return 0, nil, err
		}
		meta.Sources = append(meta.Sources, src.Proto())
	}
	// Deterministic results are nice.
	sort.Slice(meta.Sources, func(a, b int) bool {
		return meta.Sources[a].LogId < meta.Sources[b].LogId
	})
	return total, meta, nil
}
