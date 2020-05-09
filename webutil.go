package webutil

import (
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

///////////////////////////////////////////////////
// Global variables to be initialized with Init
///////////////////////////////////////////////////

var (
	WEB_PATH             string
	HEADER_PAGE_TITLE    string
	HEADER_TEMPLATE_FILE string
	ERROR_TEMPLATE_FILE  string
	BACKGROUNDPICSDIR    string
	SESSIONNAME          string
	SESSIONSTORENAME     string
	SESSIONALERTS        string
	Store                *sessions.CookieStore
)

type ConsoleUserT struct {
	Login    string
	Password string
	IsAdmin  bool
}

var ConsoleUsers map[string]ConsoleUserT
var headerTemplate *template.Template

type AlertT struct {
	AlertType string
	Msg       string
}

const (
	ALERT_SUCCESS = "success"
	ALERT_INFO    = "info"
	ALERT_WARNING = "warning"
	ALERT_DANGER  = "danger"
)

func Init(
	_WEB_PATH string,
	_HEADER_PAGE_TITLE string,
	_HEADER_TEMPLATE_FILE string,
	_ERROR_TEMPLATE_FILE string,
	_BACKGROUNDPICSDIR string,
	_SESSIONNAME string,
	_SESSIONSTORENAME string,
	_SESSIONALERTS string,
	_ConsoleUsers map[string]ConsoleUserT,
) {
	WEB_PATH = _WEB_PATH
	HEADER_PAGE_TITLE = _HEADER_PAGE_TITLE
	HEADER_TEMPLATE_FILE = _HEADER_TEMPLATE_FILE
	ERROR_TEMPLATE_FILE = _ERROR_TEMPLATE_FILE
	headerTemplate = template.Must(template.ParseFiles(
		WEB_PATH+HEADER_TEMPLATE_FILE,
		WEB_PATH+ERROR_TEMPLATE_FILE,
	))
	BACKGROUNDPICSDIR = _BACKGROUNDPICSDIR
	SESSIONNAME = _SESSIONNAME
	SESSIONSTORENAME = _SESSIONSTORENAME
	SESSIONALERTS = _SESSIONALERTS
	Store = sessions.NewCookieStore([]byte(SESSIONSTORENAME))
	ConsoleUsers = _ConsoleUsers
}

func Reload(w http.ResponseWriter, req *http.Request, where string) {
	http.Redirect(w, req, where, http.StatusSeeOther)
}

func PlaceHeader(w http.ResponseWriter, req *http.Request) {
	session, err := Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	login, _ := session.Values["login"]
	adminrights, _ := session.Values["adminrights"].(bool)
	passdata := map[string]interface{}{
		"pagetitle":   HEADER_PAGE_TITLE,
		"login":       fmt.Sprintf("%v", login),
		"adminrights": adminrights,
		"alerts":      PopAlerts(w, req),
	}
	headerTemplate.ExecuteTemplate(w, HEADER_TEMPLATE_FILE, passdata)
}

func Logged(w http.ResponseWriter, req *http.Request) bool {
	session, err := Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	_, logged := session.Values["login"]
	return logged
}

func LoggedAsAdmin(w http.ResponseWriter, req *http.Request) bool {
	session, err := Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	adminrights, logged := session.Values["adminrights"]
	return logged && adminrights.(bool)
}

func ShowError(w http.ResponseWriter, req *http.Request, msg string) {
	PlaceHeader(w, req)
	headerTemplate.ExecuteTemplate(w, ERROR_TEMPLATE_FILE, msg)
}

func ShowErrorf(w http.ResponseWriter, req *http.Request, format string, fparams ...interface{}) {
	ShowError(w, req, fmt.Sprintf(format, fparams...))
}

func PushAlert(w http.ResponseWriter, req *http.Request, alerttype string, msg string) {
	session, err := Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.AddFlash(fmt.Sprintf("%v|%v", alerttype, msg))
	session.Save(req, w)
}

func PushAlertf(w http.ResponseWriter, req *http.Request, alerttype string, format string, fparams ...interface{}) {
	PushAlert(w, req, alerttype, fmt.Sprintf(format, fparams...))
}

func PopAlerts(w http.ResponseWriter, req *http.Request) (result []AlertT) {
	session, err := Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	alerts := session.Flashes()
	if len(alerts) == 0 {
		return
	}
	for _, v := range alerts {
		ss := strings.SplitAfterN(v.(string), "|", 2)
		if len(ss) != 2 {
			result = append(result, AlertT{"danger", "ERROR!!! Wrong alert"})
			return
		}
		alertType, msg := ss[0], ss[1]
		alertType = strings.Trim(alertType, "|")

		result = append(result, AlertT{alertType, msg})
	}
	session.Save(req, w)
	return
}

func BackgoundPic() (result string) {
	dir, err := ioutil.ReadDir(BACKGROUNDPICSDIR)
	if err != nil {
		log.Print(err)
	}
	return BACKGROUNDPICSDIR + dir[rand.Intn(len(dir))].Name()

}
