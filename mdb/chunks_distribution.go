// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

// ChunkStats stores chunk stats
type ChunkStats struct {
	Empty   int64
	Jumbo   int64
	Objects int64
	Shard   string
	Size    int64
	Total   int64
}
