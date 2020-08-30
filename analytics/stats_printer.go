// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var loc, _ = time.LoadLocation("Local")

// PrintAllStats print all stats
func PrintAllStats(docs []ServerStatusDoc, span int) string {
	var lines []string

	if span < 0 {
		span = int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	}

	lines = append(lines, printStatsDetails(docs, span))
	lines = append(lines, printGlobalLockDetails(docs, span))
	lines = append(lines, printLatencyDetails(docs, span))
	lines = append(lines, printMetricsDetails(docs, span))
	lines = append(lines, printWiredTigerCacheDetails(docs, span))
	lines = append(lines, printWiredTigerConcurrentTransactionsDetails(docs, span))
	return strings.Join(lines, "")
}

// printStatsDetails -
func printStatsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	var iops uint64
	var dur uint64
	if span < 0 {
		span = 60
	}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	lines = append(lines, "\n--- Analytic Summary ---")
	lines = append(lines, "+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+")
	lines = append(lines, "| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |")
	lines = append(lines, "|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		dur = uint64(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
		if cnt == 0 {
			stat1 = stat2
		} else if cnt == 1 {
			iops = stat2.OpCounters.Command - stat1.OpCounters.Command +
				stat2.OpCounters.Delete - stat1.OpCounters.Delete +
				stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
				stat2.OpCounters.Insert - stat1.OpCounters.Insert +
				stat2.OpCounters.Query - stat1.OpCounters.Query +
				stat2.OpCounters.Update - stat1.OpCounters.Update
			if dur > 0 {
				iops = iops / dur
			} else {
				iops = 0
			}
			if (stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults) >= 0 &&
				(stat2.OpCounters.Command-stat1.OpCounters.Command) >= 0 &&
				iops >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.Mem.Resident,
					stat2.Mem.Virtual,
					stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
					stat2.OpCounters.Command-stat1.OpCounters.Command,
					stat2.OpCounters.Delete-stat1.OpCounters.Delete,
					stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
					stat2.OpCounters.Insert-stat1.OpCounters.Insert,
					stat2.OpCounters.Query-stat1.OpCounters.Query,
					stat2.OpCounters.Update-stat1.OpCounters.Update, iops))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")

			}
			stat1 = stat2
		} else if stat2.Host == stat1.Host {
			if cnt == len(docs)-1 || dur >= uint64(span) {
				iops := stat2.OpCounters.Command - stat1.OpCounters.Command +
					stat2.OpCounters.Delete - stat1.OpCounters.Delete +
					stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
					stat2.OpCounters.Insert - stat1.OpCounters.Insert +
					stat2.OpCounters.Query - stat1.OpCounters.Query +
					stat2.OpCounters.Update - stat1.OpCounters.Update
				if dur > 0 {
					iops = iops / dur
				} else {
					iops = 0
				}

				if (stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults) >= 0 &&
					(stat2.OpCounters.Command-stat1.OpCounters.Command) >= 0 &&
					iops >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						stat2.Mem.Resident,
						stat2.Mem.Virtual,
						stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
						stat2.OpCounters.Command-stat1.OpCounters.Command,
						stat2.OpCounters.Delete-stat1.OpCounters.Delete,
						stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
						stat2.OpCounters.Insert-stat1.OpCounters.Insert,
						stat2.OpCounters.Query-stat1.OpCounters.Query,
						stat2.OpCounters.Update-stat1.OpCounters.Update, iops))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|")
				}
				stat1 = stat2
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+")
	return strings.Join(lines, "\n")
}

// printLatencyDetails -
func printLatencyDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	var r, w, c float64
	var stat1, stat2 ServerStatusDoc
	if span < 0 {
		span = 60
	}
	lines = append(lines, "\n--- Latencies Summary (ms) ---")
	lines = append(lines, "+-------------------------+----------+----------+----------+")
	lines = append(lines, "| Date/Time               | reads    | writes   | commands |")
	lines = append(lines, "|-------------------------|----------|----------|----------|")
	for cnt, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)

		if cnt == 0 {
			stat1 = stat2
			continue
		}

		d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
		if d >= span {
			r = 0
			if stat2.OpLatencies.Reads.Ops > 0 {
				r = float64(stat2.OpLatencies.Reads.Latency) / float64(stat2.OpLatencies.Reads.Ops) / 1000
			}
			w = 0
			if stat2.OpLatencies.Writes.Ops > 0 {
				w = float64(stat2.OpLatencies.Writes.Latency) / float64(stat2.OpLatencies.Writes.Ops) / 1000
			}
			c = 0
			if stat2.OpLatencies.Commands.Ops > 0 {
				c = float64(stat2.OpLatencies.Commands.Latency) / float64(stat2.OpLatencies.Commands.Ops) / 1000
			}
			lines = append(lines, fmt.Sprintf("|%-25s|%10.1f|%10.1f|%10.1f|",
				stat2.LocalTime.In(loc).Format(time.RFC3339), r, w, c))
			stat1 = stat2
		}
	}
	lines = append(lines, "+-------------------------+----------+----------+----------+")
	return strings.Join(lines, "\n")
}

// printMetricsDetails -
func printMetricsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	var stat1, stat2 ServerStatusDoc
	if span < 0 {
		span = 60
	}

	lines = append(lines, "\n--- Metrics ---")
	lines = append(lines, "+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+")
	lines = append(lines, "| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |")
	lines = append(lines, "|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if stat1.Host != "" {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if d >= span {
				if stat2.Uptime > stat1.Uptime {
					lines = append(lines, fmt.Sprintf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned,
						stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects,
						stat2.Metrics.Operation.ScanAndOrder-stat1.Metrics.Operation.ScanAndOrder,
						stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts,
						stat2.Metrics.Document.Deleted-stat1.Metrics.Document.Deleted,
						stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted,
						stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned,
						stat2.Metrics.Document.Updated-stat1.Metrics.Document.Updated))
				} else {
					lines = append(lines, "+-- REBOOT ---------------+----------+------------+------------+--------------+----------+----------+----------+----------+")
				}
				stat1 = stat2
			}
		} else {
			stat1 = stat2
		}
	}
	lines = append(lines, "+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+")
	return strings.Join(lines, "\n")
}

// printGlobalLockDetails prints globalLock stats
func printGlobalLockDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat := ServerStatusDoc{}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	acm := uint64(0)
	lines = append(lines, "\n--- Global Locks Summary ---")
	lines = append(lines, "+-------------------------+--------------+--------------------------------------------+--------------------------------------------+")
	lines = append(lines, "|                         | Total Time   | Active Clients                             | Current Queue                              |")
	lines = append(lines, "| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if cnt == 0 {
			stat1 = stat
			stat2.Host = stat1.Host
		} else if cnt == 1 {
			if (stat.GlobalLock.TotalTime - stat1.GlobalLock.TotalTime) >= 0 {
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|",
					stat.LocalTime.In(loc).Format(time.RFC3339),
					(stat.GlobalLock.TotalTime-stat1.GlobalLock.TotalTime)/1000,
					stat.GlobalLock.CurrentQueue.Total,
					stat.GlobalLock.CurrentQueue.Readers,
					stat.GlobalLock.CurrentQueue.Writers,
					stat.GlobalLock.CurrentQueue.Total,
					stat.GlobalLock.CurrentQueue.Readers,
					stat.GlobalLock.CurrentQueue.Writers))
			} else {
				cnt = 0
				lines = append(lines, "|-- REBOOT ---------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
			}
			stat1 = stat
			stat2 = ServerStatusDoc{}
			stat2.Host = stat1.Host
		} else if stat2.Host == stat.Host {
			d := int(stat.LocalTime.Sub(stat1.LocalTime).Seconds())
			acm++
			stat2.LocalTime = stat.LocalTime
			stat2.GlobalLock.TotalTime = stat.GlobalLock.TotalTime
			stat2.GlobalLock.CurrentQueue.Total = stat.GlobalLock.CurrentQueue.Total
			stat2.GlobalLock.CurrentQueue.Readers += stat.GlobalLock.CurrentQueue.Readers
			stat2.GlobalLock.CurrentQueue.Writers += stat.GlobalLock.CurrentQueue.Writers
			stat2.GlobalLock.CurrentQueue.Total += stat.GlobalLock.CurrentQueue.Total
			stat2.GlobalLock.CurrentQueue.Readers += stat.GlobalLock.CurrentQueue.Readers
			stat2.GlobalLock.CurrentQueue.Writers += stat.GlobalLock.CurrentQueue.Writers
			if cnt == len(docs)-1 || d >= span {
				if (stat.GlobalLock.TotalTime - stat1.GlobalLock.TotalTime) >= 0 {
					lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|",
						stat2.LocalTime.In(loc).Format(time.RFC3339),
						(stat2.GlobalLock.TotalTime-stat1.GlobalLock.TotalTime)/1000,
						stat2.GlobalLock.CurrentQueue.Total/acm,
						stat2.GlobalLock.CurrentQueue.Readers/acm,
						stat2.GlobalLock.CurrentQueue.Writers/acm,
						stat2.GlobalLock.CurrentQueue.Total/acm,
						stat2.GlobalLock.CurrentQueue.Readers/acm,
						stat2.GlobalLock.CurrentQueue.Writers/acm))
				} else {
					cnt = 0
					lines = append(lines, "|-- REBOOT ---------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
				}
				acm = 0
				stat1 = stat2
				stat2 = ServerStatusDoc{}
				stat2.Host = stat1.Host
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}

// printWiredTigerCacheDetails prints wiredTiger cache stats
func printWiredTigerCacheDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	var stat1, stat2 ServerStatusDoc
	if span < 0 {
		span = 60
	}

	lines = append(lines, "\n--- WiredTiger Cache Summary ---")
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	lines = append(lines, "|                         |              |              |              | Modified     | Unmodified   | BytesRead    | BytesWritten |")
	lines = append(lines, "|                         | MaxBytes     | Currently    | Tracked      | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |")
	lines = append(lines, "| Date/Time               | Configured   | InCache      | DirtyBytes   | per Minute   | per Minute   | per Minute   | per Minute   |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for cnt, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt == 0 || stat2.Uptime < stat1.Uptime {
			stat1 = stat2
		} else {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if d >= span {
				minutes := stat2.LocalTime.Sub(stat1.LocalTime).Minutes()
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14.0f|%14.0f|%14.0f|%14.0f|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.Cache.MaxBytesConfigured,
					stat2.WiredTiger.Cache.CurrentlyInCache,
					stat2.WiredTiger.Cache.TrackedDirtyBytes,
					float64(stat2.WiredTiger.Cache.ModifiedPagesEvicted-stat1.WiredTiger.Cache.ModifiedPagesEvicted)/minutes,
					float64(stat2.WiredTiger.Cache.UnmodifiedPagesEvicted-stat1.WiredTiger.Cache.UnmodifiedPagesEvicted)/minutes,
					float64(stat2.WiredTiger.Cache.BytesReadIntoCache-stat1.WiredTiger.Cache.BytesReadIntoCache)/minutes,
					float64(stat2.WiredTiger.Cache.BytesWrittenFromCache-stat1.WiredTiger.Cache.BytesWrittenFromCache)/minutes))
				stat1 = stat2
			}
		}
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}

// printWiredTigerConcurrentTransactionsDetails prints wiredTiger concurrentTransactions stats
func printWiredTigerConcurrentTransactionsDetails(docs []ServerStatusDoc, span int) string {
	var lines []string
	if span < 0 {
		span = 60
	}
	stat := ServerStatusDoc{}
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	acm := uint64(0)
	lines = append(lines, "\n--- WiredTiger Concurrent Transactions Summary ---")
	lines = append(lines, "+-------------------------+--------------------------------------------+--------------------------------------------+")
	lines = append(lines, "|                         | Read Ticket                                | Write Ticket                               |")
	lines = append(lines, "| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |")
	lines = append(lines, "|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if cnt == 0 {
			stat1 = stat
			stat2.Host = stat1.Host
		} else {
			d := int(stat.LocalTime.Sub(stat1.LocalTime).Seconds())
			acm++
			stat2.LocalTime = stat.LocalTime
			stat2.WiredTiger.ConcurrentTransactions.Read.Available += stat.WiredTiger.ConcurrentTransactions.Read.Available
			stat2.WiredTiger.ConcurrentTransactions.Read.Out += stat.WiredTiger.ConcurrentTransactions.Read.Out
			stat2.WiredTiger.ConcurrentTransactions.Read.TotalTickets += stat.WiredTiger.ConcurrentTransactions.Read.TotalTickets
			stat2.WiredTiger.ConcurrentTransactions.Write.Available += stat.WiredTiger.ConcurrentTransactions.Write.Available
			stat2.WiredTiger.ConcurrentTransactions.Write.Out += stat.WiredTiger.ConcurrentTransactions.Write.Out
			stat2.WiredTiger.ConcurrentTransactions.Write.TotalTickets += stat.WiredTiger.ConcurrentTransactions.Write.TotalTickets
			if cnt == len(docs)-1 || d >= span {
				lines = append(lines, fmt.Sprintf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.ConcurrentTransactions.Read.Available/acm,
					stat2.WiredTiger.ConcurrentTransactions.Read.Out/acm,
					stat2.WiredTiger.ConcurrentTransactions.Read.TotalTickets/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.Available/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.Out/acm,
					stat2.WiredTiger.ConcurrentTransactions.Write.TotalTickets/acm))
				acm = 0
				stat1 = stat2
				stat2 = ServerStatusDoc{}
				stat2.Host = stat1.Host
			}
		}
		cnt++
	}
	lines = append(lines, "+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+")
	return strings.Join(lines, "\n")
}
