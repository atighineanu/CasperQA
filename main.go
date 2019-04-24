package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Configuration struct {
	Admin   SSHInfo
	Master  SSHInfo
	Worker1 SSHInfo
	Worker2 SSHInfo
	Key     string
}

type SSHInfo struct {
	User string
	Pass string
	IP   string
}

func Keyloader(configuration Configuration) *exec.Cmd {

	templ, err := template.ParseFiles("keyload.template")
	if err != nil {
		log.Fatalf("Search didn't work...%s", err)
	}
	var f *os.File
	f, err = os.Create("keyloader.sh")
	if err != nil {
		log.Fatalf("couldn't create the file...%s", err)
	}
	err = templ.Execute(f, configuration)
	f.Close()
	err = exec.Command("chmod", "+x", "keyloader.sh").Run()
	if err != nil {
		log.Fatalf("couldn't execute command...%s", err)
	}
	out, err := exec.Command("ls", "-alh").CombinedOutput()
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

	return exec.Command("./keyloader.sh")
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
		out, err := Node.Admin.Command(cmd...).CombinedOutput()
		if err != nil {
			fmt.Println("This is bad! ", cmd, "didn't work:", err)
		}
		fmt.Printf("%s", string(out))
	}
	return fmt.Sprintf("%s", out)
}

func AdminOrchestrator(Node Configuration) {

	//------------------when puttin alias into .bashrc/// OBSOLETE-----------------------------------
	/*
			expr := []byte("alias salt-cluster='docker exec $(docker ps -q --filter name=salt-master) salt -P \"roles:admin|kube-master|kube-minion\"'")
			var f *os.File
			f, err := os.Create("temp")
			if err != nil {
				log.Fatalf("couldn't create the file...%s", err)
			}
			f.Write(expr)
			f.Close()
			out, err := exec.Command("scp", "-o", "StrictHostKeyChecking=no", "temp", Node.Admin.User+"@"+Node.Admin.IP+":/root/.bashrc").CombinedOutput()
			fmt.Println(fmt.Sprintf("%s", string(out)))
			err = os.Remove("temp")
			if err != nil {
				fmt.Printf("Bad! couldn't delete the temp file: %s", err)
			}
			fmt.Println(SimpleShellExec(Node, []string{"source", ".bashrc"}))
		out1 := SimpleShellExec(Node, []string{"alias"})
		if !(strings.Contains(out1, "name=salt-master")) {
			fmt.Println("Bad!")
		} else {
			fmt.Println("alias is set fine...")
		}
	*/

	out1 := SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
	fmt.Println(out1)
}

func main() {
	var out []byte
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("This is bad! .json decoding didn't work:", err)
	}

	// KEYLOADER PART----------------------------------------

	if (len(os.Args) > 1) && os.Args[1] == "new" {
		out, err = Keyloader(configuration).CombinedOutput()
		if err != nil {
			fmt.Println("This is bad! keyloader script didn't work:", err)
		}

		if strings.Contains(fmt.Sprintf("%s", string(out)), "try logging") {
			fmt.Println("Successfully uploaded the keys!")
		}
	}
	// END OF KEYLOADER PART----------------------------------

	/*
		cmd := []string{"ls", "-alh"}
		out, _ = configuration.Admin.Command(cmd...).CombinedOutput()
		fmt.Printf("%s", string(out))
		out, _ = configuration.Master.Command(cmd...).CombinedOutput()
		fmt.Printf("%s", string(out))
		out, _ = configuration.Worker1.Command(cmd...).CombinedOutput()
		fmt.Printf("%s", string(out))
		out, _ = configuration.Worker2.Command(cmd...).CombinedOutput()
		fmt.Printf("%s", string(out))
	*/

	AdminOrchestrator(configuration)

}
