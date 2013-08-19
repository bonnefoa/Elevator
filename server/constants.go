package server

// Command line parsing default values
const (
	DefaultConfigFile = "/etc/elevator/elevator.conf"
	DefaultDaemonMode = false
	DefaultEndpoint   = "tcp://127.0.0.1:4141"
	DefaultLogLevel   = "INFO"
	DefaultPidfile    = "/var/run/elevator.pid"
	DefaultWorkers    = 5
)
