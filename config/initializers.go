package config

import (
	"io"
	"log/syslog"
	"os"
	"strings"

	"github.com/rightscale/rslog"

	"gopkg.in/alecthomas/kingpin.v1"
	"gopkg.in/inconshreveable/log15.v2"
	log "gopkg.in/inconshreveable/log15.v2"
)

const (
	version = "0.0.1"
	// APIVersion is a default Azure API version
	// TODO: remove this const or introduce api version per service
	APIVersion = "2014-12-01-Preview"
	// MediaType is default media type for requests to the Azure cloud
	MediaType = "application/json"
	// UserAgent is a RS request sign
	UserAgent = "RightScale Self-Service Plugin"
	// SyslogAddr is the address to use for connecting to syslog.
	SyslogAddr = "syslog:514"
	// ApplicationName is, you know, the name of the application
	ApplicationName = "azure_arm_proxy"
)

var (
	app = kingpin.New("azure_plugin", "Azure V2 RightScale Self-Service plugin.")
	// ListenFlag is a hostname and port to listen
	ListenFlag = app.Flag("listen", "Hostname and port to listen on, e.g. 'localhost:8080' - hostname is optional.").Default("localhost:8080").String()
	// Env is environment name
	Env = app.Flag("env", "Environment name: 'development' (default) or 'production'.").Default("development").String()
	// AppPrefix is URL prefix
	AppPrefix = app.Flag("prefix", "URL prefix.").Default("").String()
	// LogType could be: stdout or syslog
	LogType = app.Flag("log_type", "Type of Logger.").Default("stdout").String()
	// ClientIDCred is the client id of the application that is registered in Azure Active Directory.
	ClientIDCred = app.Arg("client", "The client id of the application that is registered in Azure Active Directory.").String()
	// ClientSecretCred is the client key of the application that is registered in Azure Active Directory.
	ClientSecretCred = app.Arg("secret", "The client key of the application that is registered in Azure Active Directory.").String()
	// SubscriptionIDCred is the client subscription id.
	SubscriptionIDCred = app.Arg("subscription", "The client subscription id.").String()
	// TenantIDCred is Azure Active Directory indentificator.
	TenantIDCred = app.Arg("tenant", "Azure Active Directory indentificator.").String()
	// RefreshTokenCred is the token used for refreshing access token.
	RefreshTokenCred = app.Arg("refresh_token", "The token used for refreshing access token.").String()
	// BaseURL is Azure cloud endpoint...set base url as variable to be able to modify it in the specs
	BaseURL = "https://management.azure.com"
	// GraphURL is the endpoint to Graph Azure service
	GraphURL = "https://graph.windows.net"
	// AuthHost is endpoint to authentication Azure service
	AuthHost = "https://login.windows.net"
	// Logger is Global syslog logger
	Logger log15.Logger
	// DebugMode is used to manage debug mode
	DebugMode = false
)

// Copy/pasted from log15/handler.go so we can specify local0 facility
type closingHandler struct {
	io.WriteCloser
	log15.Handler
}

// Copy/pasted from log15/syslog.go so we can specify local0 facility
func newSyslogNetHandler(net, addr string, tag string, fmtr log15.Format) (log15.Handler, error) {
	wr, err := syslog.Dial(net, addr, syslog.LOG_LOCAL0, tag)
	return newSyslogHandler(fmtr, wr, err)
}

// Copy/pasted from log15/syslog.go so we can specify local0 facility
func newSyslogHandler(fmtr log15.Format, sysWr *syslog.Writer, err error) (log15.Handler, error) {
	if err != nil {
		return nil, err
	}
	h := log15.FuncHandler(func(r *log15.Record) error {
		var syslogFn = sysWr.Info
		switch r.Lvl {
		case log15.LvlCrit:
			syslogFn = sysWr.Crit
		case log15.LvlError:
			syslogFn = sysWr.Err
		case log15.LvlWarn:
			syslogFn = sysWr.Warning
		case log15.LvlInfo:
			syslogFn = sysWr.Info
		case log15.LvlDebug:
			syslogFn = sysWr.Debug
		}

		s := strings.TrimSpace(string(fmtr.Format(r)))
		return syslogFn(s)
	})
	return log15.LazyHandler(&closingHandler{sysWr, h}), nil
}

func init() {
	// Parse command line
	app.Version(version)
	app.Parse(os.Args[1:])

	Logger = log15.New()
	var handler log.Handler
	Logger.Info("config loaded", "LogType", *LogType)
	switch *LogType {
	case "stdout":
		handler = log.StreamHandler(os.Stdout, rslog.SimpleFormat(true))
	case "syslog":
		// We use the TCP syslog handler, there is no option!
		h, err := newSyslogNetHandler("tcp", SyslogAddr, ApplicationName, log15.LogfmtFormat())
		if err != nil {
			kingpin.Fatalf(err.Error())
		}
		handler = h
	default:
		kingpin.Fatalf("Unknown log type: %s", *LogType)
	}

	Logger.SetHandler(handler)

	switch *Env {
	case "development":
		// add development specific settings here
		DebugMode = true
	case "production":
		// add production specific settings here
		// example: *ListenFlag = "rightscale.com:80"
	default:
		panic("Unknown environmental name: " + *Env)
	}

}
