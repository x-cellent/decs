package decs

type versionTracker struct {
	cbusID string
	state  map[string]*version
}

type version struct {
	current     uint
	lastHandled uint
}

func newVersion() *version {
	return &version{}
}

func newVersionTracker(cbusID string) *versionTracker {
	return &versionTracker{
		cbusID: cbusID,
		state:  make(map[string]*version),
	}
}

func (vt *versionTracker) version(cbusID string) *version {
	v := vt.state[cbusID]
	if v == nil {
		v = newVersion()
		vt.state[cbusID] = v
	}
	return v
}

func (vt *versionTracker) next() uint {
	v := vt.version(vt.cbusID)
	v.current++
	return v.current
}

func (vt *versionTracker) alreadyHandled(cmd *command) bool {
	v := vt.version(cmd.BusId)
	handled := v.lastHandled >= cmd.Version_
	if !handled {
		v.lastHandled = cmd.Version_
	}
	return handled
}

func (vt *versionTracker) update(cbusID string, version uint) {
	v := vt.state[cbusID]
	v.lastHandled = version
}
