// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// A writer that writes to logger.
package logwriter

import (
	"bufio"
	"io"
	"log"
)

type logWriter struct {
	prefix string
	pr     *io.PipeReader
	pw     *io.PipeWriter
	cc     chan error
}

// Create a writer that logs to logger, prefixing every line with prefix
// If logger is nil, write to standard logger
func New(prefix string, logger *log.Logger) *logWriter {
	lw := &logWriter{
		prefix: prefix,
		cc:     make(chan error),
	}
	lw.pr, lw.pw = io.Pipe()
	go func() {
		scanner := bufio.NewScanner(lw.pr)
		for scanner.Scan() {
			if logger != nil {
				logger.Printf("%s%s", lw.prefix, scanner.Text())
			} else {
				log.Printf("%s%s", lw.prefix, scanner.Text())
			}
		}
		lw.cc <- scanner.Err()
	}()
	return lw
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	return lw.pw.Write(p)
}

// Close the writer. Will not close the underlying io.Writer passed to New
func (lw *logWriter) Close() error {
	pwerr := lw.pw.Close()
	scerr := <-lw.cc
	prerr := lw.pr.Close()
	if pwerr != nil {
		return pwerr
	}
	if prerr != nil {
		return prerr
	}
	return scerr
}
