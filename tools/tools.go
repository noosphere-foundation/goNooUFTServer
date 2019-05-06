package tools

type ConnectionSettings struct {
	PWS  int    // packet window size
	UUID string // unique client identifier
}

type DataStructure struct {
	ID   int
	DATA string
}

type CommandStructure struct {
	COMMAND string
	DATA    []int
}

const (
	NodeTCPPort string = "3333"
	NodeUDPPort string = "4444"
)
