// Copyright The Linux Foundation
// SPDX-License-Identifier: Apache-2.0

package spdxtvmanager

import (
	"fmt"
	"strings"
	"unicode"
)

type tagvalue struct {
	tag   string
	value string
}

type spdxTVReader struct {
	midtext      bool
	tvList       []tagvalue
	currentLine  int
	currentTag   string
	currentValue string
}

func (reader *spdxTVReader) finalize() ([]tagvalue, error) {
	if reader.midtext {
		return nil, fmt.Errorf("finalize called while still midtext parsing a text tag")
	}
	return reader.tvList, nil
}

func (reader *spdxTVReader) readNextLine(line string) error {
	reader.currentLine++

	if reader.midtext {
		return reader.readNextLineFromMidtext(line)
	}

	return reader.readNextLineFromReady(line)
}

func (reader *spdxTVReader) readNextLineFromReady(line string) error {
	// strip whitespace from beginning of line
	line2 := strings.TrimLeftFunc(line, func(r rune) bool {
		return unicode.IsSpace(r)
	})

	// ignore empty lines
	if line2 == "" {
		return nil
	}

	// ignore comment lines
	if strings.HasPrefix(line2, "#") {
		return nil
	}

	// split at colon
	substrings := strings.SplitN(line2, ":", 2)
	if len(substrings) == 1 {
		// error if a colon isn't found
		return fmt.Errorf("no colon found in '%s'", line)
	}

	// the first substring is the tag
	reader.currentTag = strings.TrimSpace(substrings[0])

	// determine whether the value contains (or starts) a <text> line
	substrings = strings.SplitN(substrings[1], "<text>", 2)
	if len(substrings) == 1 {
		// no <text> tag found means this is a single-line value
		// strip whitespace and use as a single line
		reader.currentValue = strings.TrimSpace(substrings[0])
	} else {
		// there was a <text> tag; now decide whether it's multi-line
		substrings = strings.SplitN(substrings[1], "</text>", 2)
		if len(substrings) > 1 {
			// there is also a </text> tag; take the middle part and
			// set as value
			reader.currentValue = substrings[0]
		} else {
			// there is no </text> tag on this line; switch to midtext
			reader.currentValue = substrings[0] + "\n"
			reader.midtext = true
			return nil
		}
	}

	// if we got here, the value was on a single line
	// so go ahead and add it to the tag-value list
	tv := tagvalue{reader.currentTag, reader.currentValue}
	reader.tvList = append(reader.tvList, tv)

	// and reset
	reader.currentTag = ""
	reader.currentValue = ""

	return nil
}

func (reader *spdxTVReader) readNextLineFromMidtext(line string) error {
	// look for whether the line closes here
	substrings := strings.SplitN(line, "</text>", 2)
	if len(substrings) == 1 {
		// doesn't contain </text>, so keep building the current value
		reader.currentValue += line + "\n"
		return nil
	}

	// contains </text>, so end and record this pair
	reader.currentValue += substrings[0]
	tv := tagvalue{reader.currentTag, reader.currentValue}
	reader.tvList = append(reader.tvList, tv)

	// and reset
	reader.midtext = false
	reader.currentTag = ""
	reader.currentValue = ""

	return nil
}
