// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Parse the output of Packer
package packerout

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type PackerOut struct {
	BuilderID  string
	ID         string
	String     string
	FilesCount int
}

func getArtifactLine(
	logger *log.Logger,
	line string,
	builderName string,
	artifactNum int,
) (field string, value string, ok bool) {

	field = ""
	value = ""
	ok = false

	cl := strings.Split(line, ",")

	if len(cl) < 5 {
		return
	}
	if cl[2] == "ui" {
		if logger == nil {
			log.Printf(
				"[INFO] [packer] [%s]: %s",
				cl[3],
				cl[4],
			)
		} else {
			logger.Printf(
				"[INFO] [packer] [%s]: %s",
				cl[3],
				cl[4],
			)
		}
	}
	if len(cl) < 6 {
		return
	}
	if cl[1] != builderName {
		return
	}
	if cl[2] != "artifact" {
		return
	}
	if cl[3] != strconv.Itoa(artifactNum) {
		return
	}

	return cl[4], cl[5], true
}

// Given Packer output on pin, modify pout to contain the data from the first
// build using builder builderName, and log ui messages to logger.
// If logger is nil, use standard logger.
func (pout *PackerOut) ParsePackerOut(
	logger *log.Logger,
	pin io.Reader,
	builderName string,
) error {
	scanner := bufio.NewScanner(pin)
	field := ""
	value := ""
	ok := false
	for scanner.Scan() {
		line := scanner.Text()
		field, value, ok = getArtifactLine(
			logger,
			line,
			builderName,
			0,
		)
		if ok {
			switch field {
			case "builder-id":
				pout.BuilderID = value
			case "id":
				pout.ID = value
			case "string":
				pout.String = value
			case "files-count":
				ifc, err := strconv.Atoi(value)
				if err == nil {
					pout.FilesCount = ifc
				}
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

// Run a Packer comand configured on cmd and ParsePackerOut its output.
func (pout *PackerOut) RunPacker(
	logger *log.Logger,
	cmd *exec.Cmd,
	builderName string,
) error {
	pr, pw := io.Pipe()
	cmd.Stdout = pw

	err := cmd.Start()
	if err != nil {
		return err
	}

	ppchan := make(chan error)
	go func() {
		ppchan <- pout.ParsePackerOut(logger, pr, builderName)
	}()

	cerr := cmd.Wait()
	werr := pw.Close()
	perr := <-ppchan
	rerr := pr.Close()
	if cerr != nil {
		return cerr
	}
	if perr != nil {
		return perr
	}
	if werr != nil {
		return werr
	}

	return rerr
}
