package common

const (
	SnnName = "snn"
	CnnName = "cnn"

	CmdHelp           = "help"
	CmdServer         = "server"
	CmdClient         = "client"
)

var CMDMap = map[string]bool {
	CmdHelp: true,
	CmdServer: true,
	CmdClient: true,
}