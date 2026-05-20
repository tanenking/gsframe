package gsframe

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type CpuMemInfo_t struct {
	CpuUsage    float64 //cpu使用率
	TotalMemery uint64  //内存总大小 b
	MemRSS      uint64  //进程内存 b
	HeapInuse   uint64  //go堆内存 b
}

var cpuMemInfo = &CpuMemInfo_t{}

func GetCpuMemInfo() *CpuMemInfo_t {
	info := &CpuMemInfo_t{}
	monitor.calcCpuMemInfo()

	info.CpuUsage = cpuMemInfo.CpuUsage
	info.TotalMemery = cpuMemInfo.TotalMemery
	info.MemRSS = cpuMemInfo.MemRSS
	info.HeapInuse = cpuMemInfo.HeapInuse

	return info
}

func GetGoRoroutineInfo() (goroutineNum int, stackInfo string) {
	defer AutoRecover()()
	goroutineNum = runtime.NumGoroutine()
	buf := make([]byte, 1024*1024)
	buf = buf[:runtime.Stack(buf, true)]
	stackInfo = string(buf)
	return goroutineNum, stackInfo

}

type monitor_t struct {
	sync.Mutex
	p        *process.Process
	prevCPU  *cpu.TimesStat
	prevTime time.Time
}

var monitor = &monitor_t{}

func init() {
	var err error
	pid := int32(os.Getpid())
	monitor.p, err = process.NewProcess(pid)
	if err != nil {
		LogErrorF("创建进程实例失败: %v", err)
		return
	}
	monitor.prevCPU, err = monitor.p.Times()
	if err != nil {
		LogErrorF("获取初始CPU耗时失败: %v", err)
		return
	}
	monitor.prevTime = time.Now()
	cpuMemInfo.TotalMemery, err = monitor.getTotalSystemMemory()
	if err != nil {
		fmt.Printf("初始化系统内存信息失败: %v\n", err)
		return
	}
}

func (r *monitor_t) getTotalSystemMemory() (uint64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("获取系统内存信息失败: %v", err)
	}
	return v.Total, nil
}
func (r *monitor_t) calcCpuMemInfo() {
	defer AutoRecover()()
	defer AutoLock(&r.Mutex)()

	currTime := time.Now()
	elapsedTime := currTime.Sub(r.prevTime).Seconds()
	if elapsedTime < 2 {
		return
	}

	var err error
	currCPU, err := r.p.Times()
	if err != nil {
		LogErrorF("获取CPU耗时失败: %v", err)
		return
	}

	elapsedCPU := (currCPU.User - r.prevCPU.User) + (currCPU.System - r.prevCPU.System)
	// cpuMemInfo.CpuUsage = (elapsedCPU / (elapsedTime * float64(runtime.NumCPU()))) * 100
	cpuMemInfo.CpuUsage = (elapsedCPU / elapsedTime) * 100

	memInfo, err := r.p.MemoryInfo()
	if err != nil {
		LogErrorF("获取内存信息失败: %v", err)
		return
	}
	cpuMemInfo.MemRSS = memInfo.RSS

	var rtMem runtime.MemStats
	runtime.ReadMemStats(&rtMem)
	cpuMemInfo.HeapInuse = rtMem.HeapInuse

	r.prevCPU = currCPU
	r.prevTime = currTime
}
