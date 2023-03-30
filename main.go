package main

import (
	"os"

	"github.com/kubefirst/git-helper/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	cmd.Execute()
}
