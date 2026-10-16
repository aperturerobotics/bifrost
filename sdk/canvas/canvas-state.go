package s4wave_canvas

// Clone clones the canvas state.
func (s *CanvasState) Clone() *CanvasState {
	if s == nil {
		return nil
	}
	return s.CloneVT()
}
