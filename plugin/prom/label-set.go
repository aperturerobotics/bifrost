package prom_plugin

import (
	"github.com/prometheus/prometheus/model/labels"
)

// NewLabelSet converts a Prometheus label set to its stored form.
func NewLabelSet(lset labels.Labels) *LabelSet {
	out := &LabelSet{Labels: make([]*Label, 0, lset.Len())}
	lset.Range(func(l labels.Label) {
		out.Labels = append(out.Labels, &Label{Name: l.Name, Value: l.Value})
	})
	return out
}

// ToLabels converts a stored label set to a Prometheus label set.
// The stored labels are already sorted by name.
func (s *LabelSet) ToLabels() labels.Labels {
	b := labels.NewScratchBuilder(len(s.GetLabels()))
	for _, l := range s.GetLabels() {
		b.Add(l.GetName(), l.GetValue())
	}
	b.Sort()
	return b.Labels()
}

// mergeLabels returns the union of the target's labels and a series' labels.
// The two sets share no names, which Store.Append enforces.
func mergeLabels(target, series labels.Labels) labels.Labels {
	// Collect both sets into one builder.
	b := labels.NewScratchBuilder(target.Len() + series.Len())
	target.Range(func(l labels.Label) { b.Add(l.Name, l.Value) })
	series.Range(func(l labels.Label) { b.Add(l.Name, l.Value) })

	// Sort by name, as a label set requires.
	b.Sort()
	return b.Labels()
}
