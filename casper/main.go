package main

import (
	"CasperQA/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Configuration struct {
	Admin   SSHInfo
	Master  SSHInfo
	Worker1 SSHInfo
	Worker2 SSHInfo
	Key     string
	Dir     string
}

type SSHInfo struct {
	User string
	Pass string
	IP   string
}

var Dir string

func Keyloader(configuration Configuration) *exec.Cmd {

	out, err := exec.Command("eval", "`ssh-agent`").CombinedOutput()
	if err != nil {
		log.Fatalf("Eval didn't work...%s", err)
	}

	if !strings.Contains(fmt.Sprintf("%s", string(out)), "pid") {
		log.Fatalf("Bad ssh-agent check!")
	}

	err = exec.Command("ssh-add").Run()
	if err != nil {
		log.Fatalf("ssh-add didn't work...%s", err)
	}

	templ, err := template.ParseFiles(Dir + "utils/keyload.template")
	if err != nil {
		log.Fatalf("Search didn't work...%s", err)
	}
	var f *os.File
	f, err = os.Create(Dir + "utils/keyloader.sh")
	if err != nil {
		log.Fatalf("couldn't create the file...%s", err)
	}
	err = templ.Execute(f, configuration)
	f.Close()
	err = exec.Command("chmod", "+x", Dir+"utils/keyloader.sh").Run()
	if err != nil {
		log.Fatalf("couldn't execute command...%s", err)
	}
	out, err = exec.Command("ls", "-alh", Dir+"utils/").CombinedOutput()
	if err != nil {
		log.Fatalf("couldn't execute command...%s", err)
	}
	tmp := strings.Split(fmt.Sprintf("%s", string(out)), "\n")
	for i := 0; i < len(tmp); i++ {
		if strings.Contains(tmp[i], "keyloader.sh") && strings.Contains(tmp[i], "rwxr") {
			fmt.Println("Goood! Keyloader file loaded!", string(out))
		}
	}
	time.Sleep(5 * time.Second)
	return exec.Command(Dir + "utils/keyloader.sh")
}

func (s *SSHInfo) Command(cmd ...string) *exec.Cmd {

	arg := append(
		[]string{"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", s.User, s.IP),
		},
		cmd...,
	)
	return exec.Command("ssh", arg...)
}

func SimpleShellExec(Node Configuration, cmd []string, flag string) string {
	var out []byte
	if flag == "alias" {
		withalias := append(
			[]string{"docker", "exec", "$(docker ps -q --filter name=salt-master)", "salt", "-P", "\"roles:admin|kube-master|kube-minion\""}, cmd...)
		out, err := Node.Admin.Command(withalias...).CombinedOutput()
		if err != nil {
			fmt.Println("This is bad! ", cmd, "didn't work:", err)
		}
		fmt.Printf("%s", string(out))
	} else {
		if flag == "admin" {
			out, err := Node.Admin.Command(cmd...).CombinedOutput()
			if err != nil {
				fmt.Println("This is bad! ", cmd, "didn't work:", err)
			}
			fmt.Printf("%s", string(out))
		}
		if flag == "master" {
			out, err := Node.Master.Command(cmd...).CombinedOutput()
			if err != nil {
				fmt.Println("This is bad! ", cmd, "didn't work:", err)
			}
			fmt.Printf("%s", string(out))
		}
		if flag == "worker1" {
			out, err := Node.Worker1.Command(cmd...).CombinedOutput()
			if err != nil {
				fmt.Println("This is bad! ", cmd, "didn't work:", err)
			}
			fmt.Printf("%s", string(out))
		}
		if flag == "worker2" {
			out, err := Node.Worker2.Command(cmd...).CombinedOutput()
			if err != nil {
				fmt.Println("This is bad! ", cmd, "didn't work:", err)
			}
			fmt.Printf("%s", string(out))
		}

	}
	return fmt.Sprintf("%s", out)
}

func TransUpdConfigChecker(Node Configuration) {
	a := []string{"admin", "master", "worker1", "worker2"}
	for i := 0; i < len(a); i++ {
		out := SimpleShellExec(Node, []string{"cat", "/etc/transactional-update.conf"}, a[i])
		var mux sync.Mutex
		if !strings.Contains(out, "ZYPPER_AUTO_IMPORT_KEYS=1") {
			mux.Lock()
			SimpleShellExec(Node, []string{"echo", "'ZYPPER_AUTO_IMPORT_KEYS=1'", ">>", "/etc/transactional-update.conf"}, a[i])
		}
		mux.Unlock()
		if !strings.Contains(SimpleShellExec(Node, []string{"cat", "/etc/transactional-update.conf"}, a[i]), "ZYPPER_AUTO_IMPORT_KEYS=1") {
			fmt.Printf("Bad! Config Didn't Work at %s\n", a[i])
		}
	}
}

func AdminOrchestrator(Node Configuration) {

	if len(os.Args) <= 1 || (len(os.Args) > 1 && os.Args[1] == "new") {
		var mux sync.Mutex
		//------------Refresh Salt Grains
		mux.Lock()
		out1 := SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
		fmt.Println(out1)
		mux.Unlock()
		//------------Registering Nodes
		mux.Lock()
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'transactional-update", "register", "-r", Node.Key + "'"}, "alias")
		fmt.Println(out1)
		mux.Unlock()
		//------------Disabling Update.Timer
		mux.Lock()
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'systemctl", "disable", "--now", "transactional-update.timer'"}, "alias")
		fmt.Println(out1)
		mux.Unlock()
		//----------------Updating With Salt
		mux.Lock()
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'/usr/sbin/transactional-update", "cleanup", "dup", "salt'"}, "alias")
		fmt.Println(out1)
		mux.Unlock()
		//------------------Refreshing Grains Again...
		mux.Lock()
		out1 = SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
		fmt.Println(out1)
		mux.Unlock()
	} else {

		for i := 0; i < len(os.Args); i++ {
			if os.Args[i] == "ref" {
				out1 := SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
				fmt.Println(out1)
			}
			if os.Args[i] == "reg" {
				out1 := SimpleShellExec(Node, []string{"cmd.run", "'transactional-update", "register", "-r", Node.Key + "'"}, "alias")
				fmt.Println(out1)
			}
			if os.Args[i] == "dis" {
				out1 := SimpleShellExec(Node, []string{"cmd.run", "'systemctl", "disable", "--now", "transactional-update.timer'"}, "alias")
				fmt.Println(out1)
			}
			if os.Args[i] == "supd" {
				out1 := SimpleShellExec(Node, []string{"cmd.run", "'/usr/sbin/transactional-update", "cleanup", "dup", "salt'"}, "alias")
				fmt.Println(out1)
			}
			if os.Args[i] == "ar" && len(os.Args) > i+1 && strings.Contains(os.Args[i+1], ".repo") {
				out1 := SimpleShellExec(Node, []string{"cmd.run", "'zypper", "ar", os.Args[i+1] + "'"}, "alias")
				fmt.Println(out1)
			}
			if os.Args[i] == "pupd" {
				TransUpdConfigChecker(Node)
				if len(os.Args) > i+1 {
					out1 := SimpleShellExec(Node, []string{"cmd.run", "'/usr/sbin/transactional-update", "reboot", "pkg", "install", "-y", os.Args[i+1] + "'"}, "alias")
					fmt.Println(out1)
				}
			}
			if os.Args[i] == "cmd" {
				temp := os.Args[2:]
				temp[0] = "'" + temp[0]
				temp[len(temp)-1] = temp[len(temp)-1] + "'"
				temp = append(temp, temp[len(temp)-1])
				for i := len(temp) - 1; i > 0; i-- {
					temp[i] = temp[i-1]
				}

				temp[0] = "cmd.run"
				out1 := SimpleShellExec(Node, temp, "alias")
				fmt.Println(out1)
			}
		}
	}
}

func main() {
	var out []byte
	configuration := Configuration{}
	file, _ := os.Open(Dir + "utils/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("This is bad! .json decoding didn't work:", err)
	}

	if configuration.Dir == "" {
		out, err := exec.Command("pwd").CombinedOutput()
		if err != nil {
			fmt.Printf("Bad!...%s", err)
		}
		Dir = strings.Replace(fmt.Sprintf("%s", string(out)), "\n", "", -1)
		out, _ = exec.Command("ls", "-alh", Dir).CombinedOutput()
		if !(strings.Contains(fmt.Sprintf("%s", string(out)), "utils")) {
			fmt.Println("This is bad! please fix your running folder!--->config.json--->Dir")
		}
	} else {
		Dir = configuration.Dir
	}

	// KEYLOADER PART----------------------------------------

	if len(os.Args) > 1 {
		if os.Args[1] == "new" {
			out, err = Keyloader(configuration).CombinedOutput()
			if err != nil {
				fmt.Println("This is bad! keyloader script didn't work:", err)
			}

			if strings.Contains(fmt.Sprintf("%s", string(out)), "try logging") {
				fmt.Println("Successfully uploaded the keys!")
			}
		}
		for i := 0; i < len(os.Args); i++ {
			if os.Args[i] == "ui" {
				utils.Runner(configuration.Admin.IP)
			}
		}
	}
	// END OF KEYLOADER PART----------------------------------
	AdminOrchestrator(configuration)
}
