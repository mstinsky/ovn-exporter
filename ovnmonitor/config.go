package ovnmonitor

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/spf13/pflag"
)

// Configuration contains parameters information.
type Configuration struct {
	ListenAddress                   string
	MetricsPath                     string
	PollTimeout                     int
	PollInterval                    int
	DatabaseNorthboundSocketRemote  string
	DatabaseNorthboundSocketControl string
	DatabaseNorthboundFileDataPath  string
	DatabaseNorthboundFilePidPath   string
	DatabaseNorthboundPortDefault   int
	DatabaseNorthboundPortSsl       int
	DatabaseNorthboundPortRaft      int
	DatabaseSouthboundSocketRemote  string
	DatabaseSouthboundSocketControl string
	DatabaseSouthboundFileDataPath  string
	DatabaseSouthboundFilePidPath   string
	DatabaseSouthboundPortDefault   int
	DatabaseSouthboundPortSsl       int
	DatabaseSouthboundPortRaft      int
	ServiceNorthdFilePidPath        string
}

// ParseFlags get parameters information.
func ParseFlags() (*Configuration, error) {
	var (
		argListenAddress = pflag.String("listen-address", ":10661", "Address to listen on for web interface and telemetry.")
		argMetricsPath   = pflag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
		argPollTimeout   = pflag.Int("ovs.timeout", 2, "Timeout on JSON-RPC requests to OVN.")
		argPollInterval  = pflag.Int("ovs.poll-interval", 30, "The minimum interval (in seconds) between collections from OVN server.")

		argDatabaseNorthboundSocketRemote  = pflag.String("database.northbound.socket.remote", "unix:/run/ovn/ovnnb_db.sock", "JSON-RPC unix socket to OVN NB db.")
		argDatabaseNorthboundSocketControl = pflag.String("database.northbound.socket.control", "/run/ovn/ovnnb_db.ctl", "control socket to OVN NB app.")
		argDatabaseNorthboundFileDataPath  = pflag.String("database.northbound.file.data.path", "/etc/ovn/ovnnb_db.db", "OVN NB db file.")
		argDatabaseNorthboundFilePidPath   = pflag.String("database.northbound.file.pid.path", "/run/ovn/ovnnb_db.pid", "OVN NB db process id file.")
		argDatabaseNorthboundPortDefault   = pflag.Int("database.northbound.port.default", 6641, "OVN NB db network socket port.")
		argDatabaseNorthboundPortSsl       = pflag.Int("database.northbound.port.ssl", 6631, "OVN NB db network socket secure port.")
		argDatabaseNorthboundPortRaft      = pflag.Int("database.northbound.port.raft", 6643, "OVN NB db network port for clustering (raft)")

		argDatabaseSouthboundSocketRemote  = pflag.String("database.southbound.socket.remote", "unix:/run/ovn/ovnsb_db.sock", "JSON-RPC unix socket to OVN SB db.")
		argDatabaseSouthboundSocketControl = pflag.String("database.southbound.socket.control", "/run/ovn/ovnsb_db.ctl", "control socket to OVN SB app.")
		argDatabaseSouthboundFileDataPath  = pflag.String("database.southbound.file.data.path", "/etc/ovn/ovnsb_db.db", "OVN SB db file.")
		argDatabaseSouthboundFilePidPath   = pflag.String("database.southbound.file.pid.path", "/run/ovn/ovnsb_db.pid", "OVN SB db process id file.")
		argDatabaseSouthboundPortDefault   = pflag.Int("database.southbound.port.default", 6642, "OVN SB db network socket port.")
		argDatabaseSouthboundPortSsl       = pflag.Int("database.southbound.port.ssl", 6632, "OVN SB db network socket secure port.")
		argDatabaseSouthboundPortRaft      = pflag.Int("database.southbound.port.raft", 6644, "OVN SB db network port for clustering (raft)")

		argServiceNorthdFilePidPath   = pflag.String("service.ovn.northd.file.pid.path", "/var/run/ovn/ovn-northd.pid", "OVN northd daemon process id file.")
	)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	config := &Configuration{
		ListenAddress:                   *argListenAddress,
		MetricsPath:                     *argMetricsPath,
		PollTimeout:                     *argPollTimeout,
		PollInterval:                    *argPollInterval,
		DatabaseNorthboundSocketRemote:  *argDatabaseNorthboundSocketRemote,
		DatabaseNorthboundSocketControl: *argDatabaseNorthboundSocketControl,
		DatabaseNorthboundFileDataPath:  *argDatabaseNorthboundFileDataPath,
		DatabaseNorthboundFilePidPath:   *argDatabaseNorthboundFilePidPath,
		DatabaseNorthboundPortDefault:   *argDatabaseNorthboundPortDefault,
		DatabaseNorthboundPortSsl:       *argDatabaseNorthboundPortSsl,
		DatabaseNorthboundPortRaft:      *argDatabaseNorthboundPortRaft,

		DatabaseSouthboundSocketRemote:  *argDatabaseSouthboundSocketRemote,
		DatabaseSouthboundSocketControl: *argDatabaseSouthboundSocketControl,
		DatabaseSouthboundFileDataPath:  *argDatabaseSouthboundFileDataPath,
		DatabaseSouthboundFilePidPath:   *argDatabaseSouthboundFilePidPath,
		DatabaseSouthboundPortDefault:   *argDatabaseSouthboundPortDefault,
		DatabaseSouthboundPortSsl:       *argDatabaseSouthboundPortSsl,
		DatabaseSouthboundPortRaft:      *argDatabaseSouthboundPortRaft,
		ServiceNorthdFilePidPath:        *argServiceNorthdFilePidPath,
	}

	slog.Info(fmt.Sprintf("ovn monitor config is %+v", config))
	return config, nil
}
