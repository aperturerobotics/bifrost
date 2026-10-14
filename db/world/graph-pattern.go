package world

// graphPattern is a graph quad filter. Empty fields match any value.
type graphPattern struct {
	subject, predicate, obj, label string
}

// newGraphPattern builds the pattern of a lookup filter. Each field is spelled
// as the graph stores it, so a filter naming a predicate as a bare string
// matches the quoted string a changed quad reports.
func newGraphPattern(filter GraphQuad) graphPattern {
	return graphPattern{
		subject:   canonicalGraphValue(filter.GetSubject()),
		predicate: canonicalGraphValue(filter.GetPredicate()),
		obj:       canonicalGraphValue(filter.GetObj()),
		label:     canonicalGraphValue(filter.GetLabel()),
	}
}

// MatchGraphQuad checks whether q, spelled as the graph stores it, matches
// every non-empty field of filter. The filter is spelled as for
// LookupGraphQuads, so it may name a predicate as a bare string.
func MatchGraphQuad(filter, q GraphQuad) bool {
	return newGraphPattern(filter).matches(q)
}

// canonicalGraphValue spells a graph quad field as the graph stores it. An
// empty field stays empty.
func canonicalGraphValue(s string) string {
	return GraphValueToString(GraphQuadStringToCayleyValue(s))
}

// matches checks whether q matches every non-empty field of the pattern.
func (p graphPattern) matches(q GraphQuad) bool {
	return matchesField(p.subject, q.GetSubject()) &&
		matchesField(p.predicate, q.GetPredicate()) &&
		matchesField(p.obj, q.GetObj()) &&
		matchesField(p.label, q.GetLabel())
}

// matchesField checks one pattern field. An empty pattern matches any value.
func matchesField(pattern, value string) bool {
	return pattern == "" || pattern == value
}
