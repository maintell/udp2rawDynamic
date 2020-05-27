package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)



func runInWindows(cmd string) (string, error) {
	result, err := exec.Command("cmd", "/c", cmd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result)), err
}

func RunCommand(cmd string) (string, error) {
	if runtime.GOOS == "windows" {
		return runInWindows(cmd)
	} else {
		return runInLinux(cmd)
	}
}

func runInLinux(cmd string) (string, error) {
	fmt.Println("Running Linux cmd:" + cmd)
	result, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result)), err
}
//根据进程名判断进程是否运行
func CheckProRunning(serverName string) (bool,string, error) {
	a := `ps ux | awk '/` + serverName + `/ && !/awk/ {print $2}'`
	pid, err := RunCommand(a)
	if err != nil {
		return false,"", err
	}
	return pid != "",pid, nil
}
//根据进程名称获取进程ID
func GetPid(serverName string) (string, error) {
	a := `ps ux | awk '/` + serverName + `/ && !/awk/ {print $2}'`
	pid, err := RunCommand(a)
	return pid , err
}

// findAndKillProcess walks iterative through the /process directory tree
// looking up the process name found in each /proc/<pid>/status file. If
// the name matches the name in the argument the process with the corresponding
// <pid> will be killed.
func findAndKillProcess(pid string) error {
	fmt.Printf("PID: %s, will be killed.\n", pid)
	pidInt,_ := strconv.Atoi(pid)
	proc, err := os.FindProcess(pidInt)
	if err != nil {
		log.Println(err)
	}
	// Kill the process
	killerr := proc.Kill()
	return killerr
}

func getIpAddr(Hostname string) string{
	names, err := net.LookupIP(Hostname)
	if err != nil {
		panic(err)
	}
	if len(names) == 0 {
		fmt.Printf("no record")
	}

	return names[0].String()

}

func main(){
	lastTimes := 0
	ipAddress := getIpAddr("file.rlhd.net");
	for true {
		ipAddressNow := getIpAddr("file.rlhd.net");
		if ipAddress != ipAddressNow {
			fmt.Printf("ip address change from %s to %s, restarting process\n",ipAddress,ipAddressNow)
			ipAddress = ipAddressNow
			runState, pid,_ := CheckProRunning("udp2rawserver")
			if runState {
				_ = findAndKillProcess(pid);
			}
			_, _ = runInLinux("/root/udp2raw/udp2rawserver -c -l 0.0.0.0:27000 -r "+ipAddressNow+":27015 -k maintell --raw-mode faketcp -a > /dev/null 2>&1 &")
		}else{
			runState, _,_ := CheckProRunning("udp2rawserver")
			if !runState {
				_, _ = runInLinux("/root/udp2raw/udp2rawserver -c -l 0.0.0.0:27000 -r "+ipAddressNow+":27015 -k maintell --raw-mode faketcp -a > /dev/null 2>&1 &")
			}
			if lastTimes > 30 {
				fmt.Printf("ip address not changed @ %s to %s\n", ipAddress)
				lastTimes =0;
			}
		}
		lastTimes++
		time.Sleep(time.Duration(1)*time.Second)
	}
}