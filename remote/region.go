package remote

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/process"
)

// Region is a management interface for region processes
type Region interface {
	WriteRegionINI(mgm.Region) error
	WriteOpensimINI([]mgm.ConfigOption) error
	Start()
	Kill()
}

type regionCmd struct {
	command string
	success string
}

type region struct {
	UUID     uuid.UUID
	cmds     chan regionCmd
	log      logger.Log
	dir      string
	hostName string
	rStat    chan<- mgm.RegionStat
}

// NewRegion constructs a Region for use
func NewRegion(rID uuid.UUID, path string, hostname string, rStat chan<- mgm.RegionStat, log logger.Log) Region {
	reg := region{}
	reg.UUID = rID
	reg.cmds = make(chan regionCmd, 8)
	reg.log = logger.Wrap(rID.String(), log)
	reg.dir = path
	reg.rStat = rStat
	reg.hostName = hostname

	go reg.communicate()

	return reg
}

func (r region) communicate() {

	//collect region statistics
	ticker := time.NewTicker(5 * time.Second)

	//object holding process reference
	var exe *exec.Cmd
	exe = nil
	var start time.Time
	var proc *process.Process

	//process communication
	terminated := make(chan bool)

	for {
		select {
		case <-terminated:
			//the process exited for some Reason
			exe = nil
		case cmd := <-r.cmds:
			switch cmd.command {
			case "start":
				//if already running, exit
				if exe != nil {
					r.log.Error("Region is already running", r.UUID)
					continue
				}
				//execute binaries
				os.Chdir(r.dir)
				cmdName := "/usr/bin/mono"
				cmdArgs := []string{"OpenSim.exe", "-console", "rest"}
				exe = exec.Command(cmdName, cmdArgs...)
				err := exe.Start()
				if err != nil {
					errMsg := fmt.Sprintf("Error starting process: %s", err.Error())
					r.log.Error(errMsg)
					continue
				}
				r.log.Info("Started Successfully")
				start = time.Now()
				proc, _ = process.NewProcess(int32(exe.Process.Pid))
				go func() {
					//wait for process, ignoring process-specific errors
					_ = exe.Wait()
					r.log.Error("Terminated")
					exe = nil
					terminated <- true
				}()
			case "kill":
				//if not running, exit
				if exe == nil {
					errMsg := fmt.Sprintf("Kill region %v failed, region is not running", r.UUID.String())
					r.log.Error(errMsg)
					continue
				}
				if err := exe.Process.Kill(); err != nil {
					errMsg := fmt.Sprintf("Error killing process: %s", err.Error())
					r.log.Error(errMsg)
				}
			default:
				r.log.Info("Received unexpected command: %v", cmd.command)
			}
		case <-ticker.C:
			stat := mgm.RegionStat{UUID: r.UUID}
			if exe == nil {
				//trivially halted if we never started
				r.rStat <- stat
				continue
			}
			stat.Running = true

			cpuPercent, err := proc.CPUPercent(0)
			if err != nil {
				errMsg := fmt.Sprintf("Error getting cpu for pid: %s", err.Error())
				r.log.Error(errMsg)
			} else {
				stat.CPUPercent = cpuPercent
			}
			//memory info from this module may not be correct....
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				errMsg := fmt.Sprintf("Error getting memory for pid: %s", err.Error())
				r.log.Error(errMsg)
			} else {
				stat.MemKB = (float64(memInfo.RSS) / 1024.0)
			}

			elapsed := time.Since(start)
			stat.Uptime = elapsed

			r.rStat <- stat
		}
	}
}

func (r region) Start() {
	cmd := regionCmd{command: "start"}
	r.cmds <- cmd
}

func (r region) Kill() {
	cmd := regionCmd{command: "kill"}
	r.cmds <- cmd
}
