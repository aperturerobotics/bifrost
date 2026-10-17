package block_gc

import "fmt"

// ConvertFormat3Record says how the current encoding holds a reference graph
// record of store format 3, which kept a record per direction for every edge,
// edges between blocks and the unreferenced queue's edges to blocks included.
// keep reports whether the record stays as it is. When hasAdd is set, add is
// the edge to apply in the current graph: the child edge of a forward record
// between blocks, or the staging edge that tracks the block an incoming
// record names, since an owner of any kind recorded that its references were
// known. record starts at the kind byte.
func ConvertFormat3Record(record []byte) (keep bool, add RefEdge, hasAdd bool, err error) {
	// Keep any record that is not an edge.
	if len(record) == 0 {
		return false, RefEdge{}, false, fmt.Errorf("decode empty reference graph record")
	}
	kind := record[0]
	if kind != kindForward && kind != kindIncoming {
		return true, RefEdge{}, false, nil
	}

	// Decode the nodes of the edge, subject first.
	first, rest, err := readNode(record[1:])
	if err != nil {
		return false, RefEdge{}, false, fmt.Errorf("decode reference graph record %q: %w", record, err)
	}
	second, _, err := readNode(rest)
	if err != nil {
		return false, RefEdge{}, false, fmt.Errorf("decode reference graph record %q: %w", record, err)
	}
	edge := RefEdge{Subject: first, Object: second}
	if kind == kindIncoming {
		edge = RefEdge{Subject: second, Object: first}
	}

	// An edge to a block from a block or the queue moves into the block's
	// list. An incoming record tracks its block.
	_, child, moves := childEdge(edge)
	switch {
	case kind == kindForward && moves && child != "":
		return false, edge, true, nil
	case kind == kindForward:
		return !moves, RefEdge{}, false, nil
	case isBlockNode(edge.Object):
		return !moves, RefEdge{Subject: NodeUnreferenced, Object: edge.Object}, true, nil
	}
	return true, RefEdge{}, false, nil
}
