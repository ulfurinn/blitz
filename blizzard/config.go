package blizzard

import (
	"fmt"
	"log/syslog"
	"strings"

	"bitbucket.org/ulfurinn/blitz"
)

type BlizzardConfig struct {
	Executables []ExeConfig
	Logger      LoggerConfig
}

type LoggerConfig struct {
	Type           string
	SyslogSeverity syslog.Priority
	SyslogFacility syslog.Priority
}

func (c *LoggerConfig) UpdateFromCommand(cmd *blitz.ConfigLoggerCommand) error {
	switch strings.ToLower(cmd.LogType) {
	case "stderr":
		c.Type = "stderr"
		return nil
	case "syslog":
		var severity syslog.Priority
		var facility syslog.Priority

		switch strings.ToLower(cmd.Syslog.Severity) {
		case "emerg":
			severity = syslog.LOG_EMERG
		case "alert":
			severity = syslog.LOG_ALERT
		case "crit":
			severity = syslog.LOG_CRIT
		case "err":
			severity = syslog.LOG_ERR
		case "warning":
			severity = syslog.LOG_WARNING
		case "notice":
			severity = syslog.LOG_NOTICE
		case "info":
			severity = syslog.LOG_INFO
		case "debug":
			severity = syslog.LOG_DEBUG
		default:
			return fmt.Errorf("invalid syslog severity: %s", cmd.Syslog.Severity)
		}

		switch strings.ToLower(cmd.Syslog.Facility) {
		case "kern":
			facility = syslog.LOG_KERN
		case "user":
			facility = syslog.LOG_USER
		case "mail":
			facility = syslog.LOG_MAIL
		case "daemon":
			facility = syslog.LOG_DAEMON
		case "auth":
			facility = syslog.LOG_AUTH
		case "syslog":
			facility = syslog.LOG_SYSLOG
		case "lpr":
			facility = syslog.LOG_LPR
		case "news":
			facility = syslog.LOG_NEWS
		case "uucp":
			facility = syslog.LOG_UUCP
		case "cron":
			facility = syslog.LOG_CRON
		case "authpriv":
			facility = syslog.LOG_AUTHPRIV
		case "ftp":
			facility = syslog.LOG_FTP
		case "local0":
			facility = syslog.LOG_LOCAL0
		case "local1":
			facility = syslog.LOG_LOCAL1
		case "local2":
			facility = syslog.LOG_LOCAL2
		case "local3":
			facility = syslog.LOG_LOCAL3
		case "local4":
			facility = syslog.LOG_LOCAL4
		case "local5":
			facility = syslog.LOG_LOCAL5
		case "local6":
			facility = syslog.LOG_LOCAL6
		case "local7":
			facility = syslog.LOG_LOCAL7
		default:
			return fmt.Errorf("invalid syslog facility: %s", cmd.Syslog.Facility)
		}

		c.Type = "syslog"
		c.SyslogSeverity = severity
		c.SyslogFacility = facility
		return nil
	default:
		return fmt.Errorf("unsupported log type: %s", cmd.LogType)
	}
}
