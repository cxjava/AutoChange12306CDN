package main

import log "github.com/cihub/seelog"

func init() {
	logger, _ := log.LoggerFromConfigAsFile("./seelog.xml")
	log.ReplaceLogger(logger)
}
