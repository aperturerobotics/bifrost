package engine

import (
	"bytes"
	"context"
	"slices"
	"sort"
)

// readCatalogue decodes and validates one bounded immutable routing page.
func (e *Engine) readCatalogue(ctx context.Context, name string) (*Catalogue, error) {
	if cached, err := e.cachedMessage(ctx, name); cached != nil || err != nil {
		if err != nil {
			return nil, err
		}
		page, ok := cached.(*Catalogue)
		if !ok {
			return nil, ErrCorrupt
		}
		return page, nil
	}
	data, err := e.readFile(ctx, name)
	if err != nil {
		return nil, err
	}
	page := new(Catalogue)
	if err := decode(data, page); err != nil {
		return nil, err
	}
	if (len(page.Children) == 0) == (len(page.Partitions) == 0) || len(page.Children) > pageFanout || len(page.Partitions) > pageFanout {
		return nil, ErrCorrupt
	}
	var previous []byte
	for i, child := range page.Children {
		if child == nil || child.File == "" || len(child.Lower) > maxKeyBytes || (i > 0 && bytes.Compare(previous, child.Lower) >= 0) {
			return nil, ErrCorrupt
		}
		previous = child.Lower
	}
	for i, partition := range page.Partitions {
		if partition == nil || len(partition.Lower) > maxKeyBytes || len(partition.Runs) > partitionRunLimit || (i > 0 && bytes.Compare(previous, partition.Lower) >= 0) {
			return nil, ErrCorrupt
		}
		previous = partition.Lower
	}
	e.cacheMessage(name, page, len(data)*2+(len(page.Children)+len(page.Partitions))*192)
	return page, nil
}

// readRun decodes a sorted run with bounded records and encoded size.
func (e *Engine) readRun(ctx context.Context, name string) (*Run, error) {
	if cached, err := e.cachedMessage(ctx, name); cached != nil || err != nil {
		if err != nil {
			return nil, err
		}
		run, ok := cached.(*Run)
		if !ok {
			return nil, ErrCorrupt
		}
		return run, nil
	}
	data, err := e.readFile(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(data) > maxRunBytes {
		return nil, ErrCorrupt
	}
	run := new(Run)
	if err := decode(data, run); err != nil {
		return nil, err
	}
	if err := validateRecords(run.Records); err != nil {
		return nil, ErrCorrupt
	}
	e.cacheMessage(name, run, len(data)*2+len(run.Records)*192)
	return run, nil
}

// updateCatalogue rewrites only paths containing this sorted mutation batch.
func (p *publication) updateCatalogue(ctx context.Context, name string, records []*Record) ([]*Child, error) {
	page, err := p.engine.readCatalogue(ctx, name)
	if err != nil {
		return nil, err
	}
	p.retire(name)
	if len(page.Children) != 0 {
		var children []*Child
		for i, child := range page.Children {
			end := len(records)
			if i+1 < len(page.Children) {
				upper := page.Children[i+1].Lower
				end = sort.Search(len(records), func(j int) bool { return bytes.Compare(records[j].Key, upper) >= 0 })
			}
			if end == 0 {
				children = append(children, child)
				continue
			}
			replaced, err := p.updateCatalogue(ctx, child.File, records[:end])
			if err != nil {
				return nil, err
			}
			children = append(children, replaced...)
			records = records[end:]
		}
		return p.writeBranches(children)
	}

	// Each leaf partitions the batch without visiting unrelated run files.
	var partitions []*Partition
	for i, partition := range page.Partitions {
		end := len(records)
		if i+1 < len(page.Partitions) {
			upper := page.Partitions[i+1].Lower
			end = sort.Search(len(records), func(j int) bool { return bytes.Compare(records[j].Key, upper) >= 0 })
		}
		if end == 0 {
			partitions = append(partitions, partition)
			continue
		}
		replaced, err := p.updatePartition(ctx, partition, records[:end])
		if err != nil {
			return nil, err
		}
		partitions = append(partitions, replaced...)
		records = records[end:]
	}
	var children []*Child
	for len(partitions) != 0 {
		count := min(pageFanout, len(partitions))
		name, err := p.add("page", &Catalogue{Partitions: partitions[:count]})
		if err != nil {
			return nil, err
		}
		children = append(children, &Child{Lower: partitions[0].Lower, File: name})
		partitions = partitions[count:]
	}
	return children, nil
}

// updatePartition appends a bounded delta or merges its complete overlap set.
func (p *publication) updatePartition(ctx context.Context, partition *Partition, records []*Record) ([]*Partition, error) {
	lower := partition.Lower
	if bytes.Compare(records[0].Key, lower) < 0 {
		lower = records[0].Key
	}
	var encodedSize int
	var hasDeletion bool
	for _, record := range records {
		hasDeletion = hasDeletion || record.Deleted
		encodedSize += record.SizeVT() + 8
	}
	if !hasDeletion && len(partition.Runs) < partitionRunLimit && len(records) <= maxBatchRecords && encodedSize <= runTargetBytes {
		name, err := p.add("run", &Run{Records: records})
		if err != nil {
			return nil, err
		}
		runs := append(slices.Clone(partition.Runs), name)
		return []*Partition{{Lower: lower, Runs: runs}}, nil
	}

	// Incorporate every older version before discarding deletion records.
	merged := make(map[string]*Record)
	for _, name := range partition.Runs {
		run, err := p.engine.readRun(ctx, name)
		if err != nil {
			return nil, err
		}
		for _, record := range run.Records {
			merged[string(record.Key)] = record
		}
		p.retire(name)
	}
	for _, record := range records {
		merged[string(record.Key)] = record
	}
	all := make([]*Record, 0, len(merged))
	for _, record := range merged {
		if !record.Deleted {
			all = append(all, record)
		}
	}
	sort.Slice(all, func(i, j int) bool { return bytes.Compare(all[i].Key, all[j].Key) < 0 })
	if len(all) == 0 {
		return nil, nil
	}

	// Split complete sorted output so no future merge inherits an unbounded range.
	var partitions []*Partition
	for len(all) != 0 {
		count, size := 0, 0
		for count < len(all) && count < maxBatchRecords {
			next := all[count].SizeVT() + 8
			if count > 0 && size+next > runTargetBytes {
				break
			}
			size += next
			count++
		}
		name, err := p.add("run", &Run{Records: all[:count]})
		if err != nil {
			return nil, err
		}
		boundary := all[0].Key
		if len(partitions) == 0 {
			boundary = lower
		}
		partitions = append(partitions, &Partition{Lower: boundary, Runs: []string{name}})
		all = all[count:]
	}
	return partitions, nil
}

// writeBranches splits routing output into bounded immutable parent pages.
func (p *publication) writeBranches(children []*Child) ([]*Child, error) {
	if len(children) == 1 {
		return children, nil
	}
	var parents []*Child
	for len(children) != 0 {
		count := min(pageFanout, len(children))
		name, err := p.add("page", &Catalogue{Children: children[:count]})
		if err != nil {
			return nil, err
		}
		parents = append(parents, &Child{Lower: children[0].Lower, File: name})
		children = children[count:]
	}
	return parents, nil
}

// finishCatalogue adds root levels only when a split requires them.
func (p *publication) finishCatalogue(children []*Child) (string, error) {
	if len(children) == 0 {
		return p.add("page", &Catalogue{Partitions: []*Partition{{}}})
	}
	for len(children) > 1 {
		var err error
		children, err = p.writeBranches(children)
		if err != nil {
			return "", err
		}
	}
	return children[0].File, nil
}
