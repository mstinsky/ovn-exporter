package ovnmonitor

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kubeovn/ovsdb"
)

const metricNamespace = "ovn"

var (
	appName          = "ovn-exporter"
	isClusterEnabled = true
	tryConnectCnt    = 0
	checkNbDbCnt     = 0
	checkSbDbCnt     = 0
)

// Exporter collects OVN data from the given server and exports them using
// the prometheus metrics package.
type Exporter struct {
	sync.RWMutex
	Client       *ovsdb.OvnClient
	timeout      int
	pollInterval int
	errors       int64
	errorsLocker sync.RWMutex
}

// OVNDBClusterStatus contains information about a cluster.
type OVNDBClusterStatus struct {
	cid             string
	sid             string
	status          string
	role            string
	leader          string
	vote            string
	term            float64
	electionTimer   float64
	logIndexStart   float64
	logIndexNext    float64
	logNotCommitted float64
	logNotApplied   float64
	connIn          float64
	connOut         float64
	connInErr       float64
	connOutErr      float64
}

// NewExporter returns an initialized Exporter.
func NewExporter(cfg *Configuration) *Exporter {
	e := Exporter{}
	e.Client = ovsdb.NewOvnClient()
	e.initParas(cfg)
	return &e
}

func (e *Exporter) initParas(cfg *Configuration) {
	e.timeout = cfg.PollTimeout
	e.pollInterval = cfg.PollInterval

	e.Client.Timeout = cfg.PollTimeout
	e.Client.System.Hostname = os.Getenv("KUBE_NODE_NAME")

	e.Client.Database.Northbound.Name = cfg.DatabaseNorthboundName
	e.Client.Database.Northbound.Socket.Remote = cfg.DatabaseNorthboundSocketRemote
	e.Client.Database.Northbound.Socket.Control = cfg.DatabaseNorthboundSocketControl
	e.Client.Database.Northbound.File.Data.Path = cfg.DatabaseNorthboundFileDataPath
	e.Client.Database.Northbound.File.Pid.Path = cfg.DatabaseNorthboundFilePidPath
	e.Client.Database.Northbound.Port.Default = cfg.DatabaseNorthboundPortDefault
	e.Client.Database.Northbound.Port.Ssl = cfg.DatabaseNorthboundPortSsl
	e.Client.Database.Northbound.Port.Raft = cfg.DatabaseNorthboundPortRaft

	e.Client.Database.Southbound.Name = cfg.DatabaseSouthboundName
	e.Client.Database.Southbound.Socket.Remote = cfg.DatabaseSouthboundSocketRemote
	e.Client.Database.Southbound.Socket.Control = cfg.DatabaseSouthboundSocketControl
	e.Client.Database.Southbound.File.Data.Path = cfg.DatabaseSouthboundFileDataPath
	e.Client.Database.Southbound.File.Pid.Path = cfg.DatabaseSouthboundFilePidPath
	e.Client.Database.Southbound.Port.Default = cfg.DatabaseSouthboundPortDefault
	e.Client.Database.Southbound.Port.Ssl = cfg.DatabaseSouthboundPortSsl
	e.Client.Database.Southbound.Port.Raft = cfg.DatabaseSouthboundPortRaft

	e.Client.Service.Vswitchd.File.Pid.Path = cfg.ServiceVswitchdFilePidPath
	e.Client.Service.Northd.File.Pid.Path = cfg.ServiceNorthdFilePidPath
}

// StartConnection connect to database socket
func (e *Exporter) StartConnection() error {
	if err := e.Client.Connect(); err != nil {
		return err
	}
	slog.Info("Exporter connected successfully to database socket")
	return nil
}

// TryClientConnection try to connect to database socket after init exporter
func (e *Exporter) TryClientConnection() {
	for {
		if tryConnectCnt > 5 {
			slog.Error("ovn-exporter failed to reconnect to db socket")
			os.Exit(1)
		}

		if err := e.StartConnection(); err != nil {
			tryConnectCnt++
			slog.Error(fmt.Sprintf("ovn-exporter failed to reconnect db socket %v times", tryConnectCnt))
		} else {
			slog.Info("ovn-exporter reconnected to db successfully")
			break
		}

		time.Sleep(5 * time.Second)
	}
}

var registerOvnMetricsOnce sync.Once

// StartOvnMetrics register and start to update ovn metrics
func (e *Exporter) StartOvnMetrics() {
	registerOvnMetricsOnce.Do(func() {
		registerOvnMetrics()

		// OVN metrics updater
		go e.ovnMetricsUpdate()
	})
}

// ovnMetricsUpdate updates the ovn metrics for every 30 sec
func (e *Exporter) ovnMetricsUpdate() {
	for {
		e.exportOvnStatusGauge()
		e.exportOvnDBFileSizeGauge()
		e.exportOvnRequestErrorGauge()
		e.exportOvnDBStatusGauge()

		e.exportOvnChassisGauge()
		e.exportLogicalSwitchGauge()
		e.exportLogicalSwitchPortGauge()

		e.exportOvnClusterEnableGauge()
		if isClusterEnabled {
			e.exportOvnClusterInfoGauge()
		}

		time.Sleep(time.Duration(e.pollInterval) * time.Second)
	}
}

// GetExporterName returns exporter name.
func GetExporterName() string {
	return appName
}

func (e *Exporter) exportOvnStatusGauge() {
	metricOvnHealthyStatus.Reset()
	result := e.getOvnStatus()
	for k, v := range result {
		metricOvnHealthyStatus.WithLabelValues(e.Client.System.Hostname, k).Set(float64(v))
	}

	metricOvnHealthyStatusContent.Reset()
	statusResult := e.getOvnStatusContent()
	for k, v := range statusResult {
		metricOvnHealthyStatusContent.WithLabelValues(e.Client.System.Hostname, k, v).Set(float64(1))
	}
}

func (e *Exporter) exportOvnDBFileSizeGauge() {
	metricDBFileSize.Reset()
	nbPath := e.Client.Database.Northbound.File.Data.Path
	sbPath := e.Client.Database.Southbound.File.Data.Path
	dirDbMap := map[string]string{
		nbPath: "OVN_Northbound",
		sbPath: "OVN_Southbound",
	}
	for dbFile, database := range dirDbMap {
		fileInfo, err := os.Stat(dbFile)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get the DB size for database %s", database), "error", err)
			return
		}
		metricDBFileSize.WithLabelValues(e.Client.System.Hostname, database).Set(float64(fileInfo.Size()))
	}
}

func (e *Exporter) exportOvnRequestErrorGauge() {
	metricRequestErrorNums.WithLabelValues(e.Client.System.Hostname).Set(float64(e.errors))
}

func (e *Exporter) exportOvnChassisGauge() {
	metricChassisInfo.Reset()
	if vteps, err := e.Client.GetChassis(); err != nil {
		slog.Error(fmt.Sprintf("%s", e.Client.Database.Southbound.Name), "error", err)
		e.IncrementErrorCounter()
	} else {
		for _, vtep := range vteps {
			metricChassisInfo.WithLabelValues(vtep.Hostname, vtep.UUID, vtep.Name, vtep.IPAddress.String()).Set(1)
		}
	}
}

func (e *Exporter) exportLogicalSwitchGauge() {
	resetLogicalSwitchMetrics()
	e.setLogicalSwitchInfoMetric()
}

func (e *Exporter) exportLogicalSwitchPortGauge() {
	resetLogicalSwitchPortMetrics()
	e.setLogicalSwitchPortInfoMetric()
}

func (e *Exporter) exportOvnClusterEnableGauge() {
	metricClusterEnabled.Reset()
	isClusterEnabled, err := getClusterEnableState(e.Client.Database.Northbound.File.Data.Path)
	if err != nil {
		slog.Error("failed to get output of cluster status", "error", err)
	}
	if isClusterEnabled {
		metricClusterEnabled.WithLabelValues(e.Client.System.Hostname, e.Client.Database.Northbound.File.Data.Path).Set(1)
	} else {
		metricClusterEnabled.WithLabelValues(e.Client.System.Hostname, e.Client.Database.Northbound.File.Data.Path).Set(0)
	}
}

func (e *Exporter) exportOvnClusterInfoGauge() {
	resetOvnClusterMetrics()
	dirDbMap := map[string]string{
		"nb": "OVN_Northbound",
		"sb": "OVN_Southbound",
	}
	for direction, database := range dirDbMap {
		clusterStatus, err := getClusterInfo(direction, database)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get Cluster Info for database %s", database), "error", err)
			return
		}
		e.setOvnClusterInfoMetric(clusterStatus, database)
	}
}

func (e *Exporter) exportOvnDBStatusGauge() {
	metricDBStatus.Reset()
	dbList := []string{"OVN_Northbound", "OVN_Southbound"}
	for _, database := range dbList {
		ok, err := getDBStatus(database)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get DB status for %s", database), "error", err)
			return
		}
		if ok {
			metricDBStatus.WithLabelValues(e.Client.System.Hostname, database).Set(1)
		} else {
			metricDBStatus.WithLabelValues(e.Client.System.Hostname, database).Set(0)

			switch database {
			case "OVN_Northbound":
				checkNbDbCnt++
				if checkNbDbCnt < 6 {
					slog.Warn(fmt.Sprintf("Failed to get OVN NB DB status for %v times", checkNbDbCnt))
					return
				}
				slog.Warn(fmt.Sprintf("Failed to get OVN NB DB status for %v times, ready to restore OVN DB", checkNbDbCnt))
				checkNbDbCnt = 0
			case "OVN_Southbound":
				checkSbDbCnt++
				if checkSbDbCnt < 6 {
					slog.Warn(fmt.Sprintf("Failed to get OVN SB DB status for %v times", checkSbDbCnt))
					return
				}
				slog.Warn(fmt.Sprintf("Failed to get OVN SB DB status for %v times, ready to restore OVN DB", checkSbDbCnt))
				checkSbDbCnt = 0
			}
		}
	}
}
