package wayland

// The screen's output name and description.
const (
	outputNameText        = "WASIVM-1"
	outputDescriptionText = "WasiVM display"
)

// output is a client's wl_output: the one screen.
type output struct {
	// c is the client.
	c *client
	// resource is the binding.
	resource
}

// request handles a wl_output request: release, from version 3.
func (o output) request(op uint16, _ *message) error {
	if op == outputRelease {
		o.c.destroy(o.id)
	}
	return nil
}

// bindOutput binds wl_output and describes the screen.
func (c *client) bindOutput(r resource) object {
	// Describe the screen: its position, maker and model, its one mode, and
	// from version 4 its name, as the xdg output names it.
	size := c.comp.desktop.area.Size()
	c.send(newEvent(r.id, outputGeometry).int(0).int(0).int(0).int(0).int(0).
		string("WasiVM").string("display").int(0))
	c.send(newEvent(r.id, outputMode).uint(outputModeCurrentPreferred).
		int(int32(size.X)).int(int32(size.Y)).int(60000)) //nolint:gosec // screen sizes are small.
	if r.version >= 2 {
		c.send(newEvent(r.id, outputScale).int(1))
	}
	if r.version >= 4 {
		c.send(newEvent(r.id, outputName).string(outputNameText))
		c.send(newEvent(r.id, outputDescription).string(outputDescriptionText))
	}
	if r.version >= 2 {
		c.send(newEvent(r.id, outputDone))
	}
	return output{c, r}
}

// bindXdgOutputManager binds zxdg_output_manager_v1, which describes an
// output in the desktop's logical coordinates. Desktop clients such as panels
// read the output's name and logical size from it.
func (c *client) bindXdgOutputManager(r resource) object {
	return requestFunc(func(op uint16, m *message) error {
		switch op {
		case xdgOutputManagerDestroy:
			c.destroy(r.id)
		case xdgOutputManagerGetXdgOutput:
			id, oid := m.uint(), m.uint()
			if m.err != nil {
				return nil
			}
			o, ok := c.objects[oid].(output)
			if !ok {
				return c.fail(r.id, errInvalidObject, "bad output")
			}
			if err := c.add(id, c.inert(id, xdgOutputDestroy)); err != nil {
				return err
			}
			c.describeXdgOutput(resource{id: id, version: r.version}, o)
		}
		return nil
	})
}

// describeXdgOutput sends the screen's logical position, size, name and
// description to the xdg output x of o, then ends the update. From version
// 3 the output's done event ends it in place of the xdg output's.
func (c *client) describeXdgOutput(x resource, o output) {
	// Describe the screen at its full size, as the scale is 1.
	size := c.comp.desktop.area.Size()
	c.send(newEvent(x.id, xdgOutputLogicalPosition).int(0).int(0))
	c.send(newEvent(x.id, xdgOutputLogicalSize).int(int32(size.X)).int(int32(size.Y))) //nolint:gosec // screen sizes are small.
	if x.version >= 2 {
		c.send(newEvent(x.id, xdgOutputName).string(outputNameText))
		c.send(newEvent(x.id, xdgOutputDescription).string(outputDescriptionText))
	}

	// End the update.
	switch {
	case x.version < 3:
		c.send(newEvent(x.id, xdgOutputDone))
	case o.version >= 2:
		c.send(newEvent(o.id, outputDone))
	}
}
