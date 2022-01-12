// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import "fmt"

// CompareClusters compares two clusters, source and target
func CompareClusters(cfg *Config) error {
	if cfg.IsDeepCompare {
		var err error
		var comp *Comparator
		if comp, err = NewComparator(cfg.SourceURI, cfg.TargetURI); err != nil {
			return err
		}
		return comp.Compare(cfg.Filters, cfg.SampleSize)
	}
	return CompareMetadata(cfg)
}

// CompareMetadata compares two clusters' metadata
func CompareMetadata(cfg *Config) error {
	var err error
	comp := NewComparison(cfg.Signature)
	comp.SetNoColor(cfg.NoColor)
	comp.SetVerbose(cfg.Verbose)
	if err = comp.Compare(cfg.SourceURI, cfg.TargetURI); err != nil {
		return err
	}
	if err = comp.OutputBSON(); err != nil {
		return err
	}
	return err
}

// PrintCompareHelp prints help message
func PrintCompareHelp() {
	message := `deprecated, use keyhole -config <config>
{
  "action": "compare_clusters",
  "deep_compare: false,
  "namespaces: [],
  "source_uri": "mongodb+srv://user:password@source.xxxxxx.mongodb.net/",
  "target_uri": "mongodb+srv://user:password@target.xxxxxx.mongodb.net/"
}`
	fmt.Println(message)
}
