package domain

type Multimeter interface {
	ProccessArray(bytearray []byte) (float64, string, []string)
}

type MultimeterCommands interface {
	Select() []byte
	Auto() []byte
	Range() []byte
	Light() []byte
	Relative() []byte
}
