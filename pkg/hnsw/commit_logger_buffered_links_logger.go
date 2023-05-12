//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2023 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package hnsw

// type bufferedLinksLogger struct {
// 	base *hnswCommitLogger
// 	buf  *bytes.Buffer
// }

// func (b *bufferedLinksLogger) ReplaceLinksAtLevel(nodeid uint64,
// 	Level int, targets []uint64) error {
// 	b.base.writeCommitType(b.buf, ReplaceLinksAtLevel)
// 	b.base.writeUint64(b.buf, nodeid)
// 	b.base.writeUint16(b.buf, uint16(Level))
// 	targetLength := len(targets)
// 	if targetLength > math.MaxUint16 {
// 		// TODO: investigate why we get such massive Connections
// 		targetLength = math.MaxUint16
// 		b.base.logger.WithField("action", "hnsw_current_commit_log").
// 			WithField("ID", b.base.ID).
// 			WithField("original_length", len(targets)).
// 			WithField("maximum_length", targetLength).
// 			Warning("condensor length of Connections would overflow uint16, cutting off")
// 	}
// 	b.base.writeUint16(b.buf, uint16(targetLength))
// 	b.base.writeUint64Slice(b.buf, targets[:targetLength])

// 	return nil
// }

// func (b *bufferedLinksLogger) AddLinkAtLevel(nodeid uint64, Level int,
// 	target uint64) error {
// 	ec := &errorCompounder{}
// 	ec.add(b.base.writeCommitType(b.buf, AddLinkAtLevel))
// 	ec.add(b.base.writeUint64(b.buf, nodeid))
// 	ec.add(b.base.writeUint16(b.buf, uint16(Level)))
// 	ec.add(b.base.writeUint64(b.buf, target))

// 	if err := ec.toError(); err != nil {
// 		return errors.Wrapf(err, "write link at Level %d->%d (%d) to commit log",
// 			nodeid, target, Level)
// 	}

// 	return nil
// }

// func (b *bufferedLinksLogger) Close() error {
// 	b.base.Lock()
// 	defer b.base.Unlock()

// 	_, err := b.base.logWriter.Write(b.buf.Bytes())
// 	if err != nil {
// 		return errors.Wrap(err, "flush link buffer to commit logger")
// 	}

// 	return nil
// }
