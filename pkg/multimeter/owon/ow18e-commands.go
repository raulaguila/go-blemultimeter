package owon

type OW18ECommands struct{}

// Buttons command
func (m *OW18ECommands) Select() []byte {
	return []byte{1, 0}
}

func (m *OW18ECommands) Auto() []byte {
	return []byte{2, 0}
}

func (m *OW18ECommands) Range() []byte {
	return []byte{2, 1}
}

func (m *OW18ECommands) Light() []byte {
	return []byte{3, 0}
}

func (m *OW18ECommands) Hold() []byte {
	return []byte{3, 1}
}

func (m *OW18ECommands) Relative() []byte {
	return []byte{4, 0}
}
