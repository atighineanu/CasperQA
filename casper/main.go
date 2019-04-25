package main

import (
	"CaaSP3_QA_Auto/utils"
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
	Dir     string
}

type SSHInfo struct {
	User string
	Pass string
	IP   string
}

var Dir string

//-----------------------------AGOUTI PART----------------------------------------------
/*
func ErrorChecker(err error, place string) {
	if err != nil {
		fmt.Printf("test encountered an error at\t"+place+"\n%s\n", err)
	}
}

func Clicker(button string, page *agouti.Page) (*agouti.Selection, error) {
	time.Sleep(2 * time.Second)
	element := page.FindByXPath(button)
	err := element.Click()
	return element, err
}

func Login(linku string, page *agouti.Page) {
	element, err := Clicker("//*[@id=\"user_email\"]", page)
	place := "user login"
	ErrorChecker(err, place)

	err = element.Fill("test@test.com")
	place = "typing user name"
	ErrorChecker(err, place)

	element, err = Clicker("//*[@id=\"user_password\"]", page)
	place = "password login"
	ErrorChecker(err, place)

	err = element.Fill("password")
	place = "typing password"
	ErrorChecker(err, place)

	element, err = Clicker("//*[@class=\"Log in\"]/input[3]", page)
	place = "clicking \"LOGIN\" "
	ErrorChecker(err, place)
}

func PageRefresher(linku string, Driver *agouti.WebDriver) *agouti.Page {
	page, err := Driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		t.Fatal("Failed to open page:", err)
	}

	if err := page.Navigate(linku); err != nil {
		fmt.Printf("Error!...%s", err)
	}
	return page
}

func Runner(ip string) {
	linku := "https://" + ip
	Driver := agouti.ChromeDriver()
	page := PageRefresher(linku, Driver)
	Login(linku, page)
}
*/
//------------------END OF AGOUTI PART--------------------------------------------

func Keyloader(configuration Configuration) *exec.Cmd {

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
	out, err := exec.Command("ls", "-alh", Dir+"utils/").CombinedOutput()
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
		out, err := Node.Admin.Command(cmd...).CombinedOutput()
		if err != nil {
			fmt.Println("This is bad! ", cmd, "didn't work:", err)
		}
		fmt.Printf("%s", string(out))
	}
	return fmt.Sprintf("%s", out)
}

func AdminOrchestrator(Node Configuration) {

	//------------------when puttin' alias into .bashrc===> OBSOLETE-----------------------------------
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

	if len(os.Args) <= 1 || (len(os.Args) > 1 && os.Args[1] == "new") {
		//------------Refresh Salt Grains
		out1 := SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
		fmt.Println(out1)

		//------------Registering Nodes
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'transactional-update", "register", "-r", Node.Key + "'"}, "alias")
		fmt.Println(out1)

		//------------Disabling Update.Timer
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'systemctl", "disable", "--now", "transactional-update.timer'"}, "alias")
		fmt.Println(out1)

		//----------------Updating With Salt
		out1 = SimpleShellExec(Node, []string{"cmd.run", "'/usr/sbin/transactional-update", "cleanup", "dup", "salt'"}, "alias")
		fmt.Println(out1)

		//------------------Refreshing Grains Again...
		out1 = SimpleShellExec(Node, []string{"saltutil.refresh_grains"}, "alias")
		fmt.Println(out1)
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
			if os.Args[i] == "upd" {
				out1 := SimpleShellExec(Node, []string{"cmd.run", "'/usr/sbin/transactional-update", "cleanup", "dup", "salt'"}, "alias")
				fmt.Println(out1)
			}
		}
	}
}

func main() {
	Dir = "/home/atighineanu/golang/src/CaaSP3_QA_Auto/"
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
		if !(strings.Contains(fmt.Sprintf("%s", string(out)), "utils/config.json")) {
			fmt.Println("This is bad! please fix your running folder!--->config.json--->Dir")
		}
	} else {
		Dir = configuration.Dir
	}

	// KEYLOADER PART----------------------------------------

	AdminOrchestrator(configuration)
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

}
