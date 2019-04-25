package utils

import (
	"fmt" //---replace it with https://github.com/atighineanu/ssher; >>> make sure you write your id_rsa password, line #20;
	"log" //--- also, export the client's ssh-key(s) to the server you are going to test (see README notes)
	"strings"
	"time"

	"github.com/sclevine/agouti"
)

const (
	NULL            = "\uE000"
	CANCEL          = "\uE001"
	HELP            = "\uE002"
	BACK_SPACE      = "\uE003"
	TAB             = "\uE004"
	CLEAR           = "\uE005"
	RETURN          = "\uE006"
	ENTER           = "\uE007"
	SHIFT           = "\uE008"
	LEFT_SHIFT      = "\uE008"
	CONTROL         = "\uE009"
	LEFT_CONTROL    = "\uE009"
	ALT             = "\uE00A"
	LEFT_ALT        = "\uE00A"
	PAUSE           = "\uE00B"
	ESCAPE          = "\uE00C"
	SPACE           = "\uE00D"
	PAGE_UP         = "\uE00E"
	PAGE_DOWN       = "\uE00F"
	END             = "\uE010"
	HOME            = "\uE011"
	LEFT            = "\uE012"
	ARROW_LEFT      = "\uE012"
	UP              = "\uE013"
	ARROW_UP        = "\uE013"
	RIGHT           = "\uE014"
	ARROW_RIGHT     = "\uE014"
	DOWN            = "\uE015"
	ARROW_DOWN      = "\uE015"
	INSERT          = "\uE016"
	DELETE          = "\uE017"
	SEMICOLON       = "\uE018"
	EQUALS          = "\uE019"
	NUMPAD0         = "\uE01A"
	NUMPAD1         = "\uE01B"
	NUMPAD2         = "\uE01C"
	NUMPAD3         = "\uE01D"
	NUMPAD4         = "\uE01E"
	NUMPAD5         = "\uE01F"
	NUMPAD6         = "\uE020"
	NUMPAD7         = "\uE021"
	NUMPAD8         = "\uE022"
	NUMPAD9         = "\uE023"
	MULTIPLY        = "\uE024"
	ADD             = "\uE025"
	SEPARATOR       = "\uE026"
	SUBTRACT        = "\uE027"
	DECIMAL         = "\uE028"
	DIVIDE          = "\uE029"
	F1              = "\uE031"
	F2              = "\uE032"
	F3              = "\uE033"
	F4              = "\uE034"
	F5              = "\uE035"
	F6              = "\uE036"
	F7              = "\uE037"
	F8              = "\uE038"
	F9              = "\uE039"
	F10             = "\uE03A"
	F11             = "\uE03B"
	F12             = "\uE03C"
	META            = "\uE03D"
	COMMAND         = "\uE03D"
	ZENKAKU_HANKAKU = "\uE040"
)

var VERSION int /// CaaSP 3 or 4

func ErrorChecker(err error, place string) {
	if err != nil {
		fmt.Printf("test encountered an error at\t"+place+"\n%s\n", err)
	}
}

func Clicker(button string, page *agouti.Page) (*agouti.Selection, error) {
	time.Sleep(2 * time.Second)
	var element *agouti.Selection
	if strings.Contains(button, "!") {
		element = page.FindByButton(strings.Replace(button, "!", "", -1))
	} else {
		element = page.FindByXPath(button)
	}
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

	element, err = Clicker("//*[@class=\"btn btn-success btn-block\"]", page)
	place = "clicking \"LOGIN\" "
	ErrorChecker(err, place)
}

func PageRefresher(linku string, Driver *agouti.WebDriver) *agouti.Page {
	page, err := Driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		fmt.Printf("Error New Page!...%s", err)
	}

	if err := page.Navigate(linku); err != nil {
		fmt.Printf("Error Navigate!...%s", err)
	}
	return page
}

func Runner(ip string) {
	linku := "https://" + ip
	Driver := agouti.ChromeDriver() //(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"})
	if err := Driver.Start(); err != nil {
		log.Fatal(err)
	}
	page, err := Driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		fmt.Printf("Error New Page!...%s", err)
	}

	if err := page.Navigate(linku); err != nil {
		fmt.Printf("Error Navigate!...%s", err)
	}
	//time.Sleep(10 * time.Second)
	Login(linku, page)
}
