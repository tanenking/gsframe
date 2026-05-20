package application

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

// deify
var logo = `


  ██████╗ ███████╗██╗███████╗██╗   ██╗
  ██╔══██╗██╔════╝██║██╔════╝╚██╗ ██╔╝
  ██║  ██║█████╗  ██║█████╗   ╚████╔╝ 
  ██║  ██║██╔══╝  ██║██╔══╝    ╚██╔╝  
  ██████╔╝███████╗██║██║        ██║   
  ╚═════╝ ╚══════╝╚═╝╚═╝        ╚═╝   
`
var topLine = `┌───────────────────────────────────────────────────┐`
var borderLine = `│`
var bottomLine = `└───────────────────────────────────────────────────┘`

var sub_pids []*os.Process

func init() {
	rand.Seed(time.Now().Unix())
	sub_pids = []*os.Process{}
}

func notifySubKill() {
	if constants.IsWindowsSystem() {
		return
	}
	//通知所有子进程,退出
	if len(sub_pids) > 0 {
		for _, ps := range sub_pids {
			ps.Signal(syscall.SIGINT)
			ps.Wait()
			fmt.Println(`子进程终止 => `, ps.Pid)
		}
	}
}

func writePid() bool {
	pid_path := os.Getenv(gsinf.Env_PidPath)
	if len(pid_path) <= 0 {
		return true
	}
	if !filepath.IsAbs(pid_path) {
		p, _ := os.Getwd()
		pid_path = p + "/" + pid_path
	}
	var is bool
	var _exists bool = false

	is, _exists = constants.IsDir(pid_path)
	if _exists && !is {
		fmt.Println("env pid_path must need a dir")
		return false
	}
	if !_exists {
		err := os.MkdirAll(pid_path, os.ModePerm)
		if err != nil {
			fmt.Printf("%v\n", err)
			return false
		}
	}

	var f *os.File
	var err1 error
	fpath := fmt.Sprintf("%spid-%s-%s-%d", pid_path, constants.ProjectName, constants.ServiceType, os.Getgid())
	if constants.CheckFileIsExist(fpath) { //如果文件存在
		os.Remove(fpath)
	}
	f, err1 = os.Create(fpath) //创建文件
	constants.Check(err1)
	pidinfo := fmt.Sprintf("%d", os.Getpid())
	f.WriteString(pidinfo)

	return true
}
