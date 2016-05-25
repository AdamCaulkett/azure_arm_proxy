package config

import (
	"fmt"
	"os"

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

func init() {
	// Parse command line
	app.Version(version)
	app.Parse(os.Args[1:])

	Logger = log15.New()
	logType := os.Getenv("LOG_TYPE")
	var handler log.Handler
	if logType == "" {
		fmt.Errorf("Environment LOG_TYPE is not set")
		return
	} else {
		Logger.Info("config loaded", "LogType", logType)
	}
	switch logType {
	case "stdout":
		handler = log.StreamHandler(os.Stdout, rslog.SimpleFormat(true))
	case "syslog":
		// We use the TCP syslog handler, there is no option!
		h, err := rslog.NewTCPSyslogHandler(SyslogAddr, ApplicationName)
		if err != nil {
			kingpin.Fatalf(err.Error())
		}
		handler = h
	case "none":
		// no handler
	default:
		kingpin.Fatalf("Unknown log type: %s", logType)
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
