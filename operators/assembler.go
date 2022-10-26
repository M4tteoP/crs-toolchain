// Copyright 2022 OWASP Core Rule Set Project
// SPDX-License-Identifier: Apache-2.0

package operators

import (
	"bufio"
	"bytes"
	"regexp"
	"sort"
	"strings"

	"github.com/theseion/crs-toolchain/v2/parser"
	"github.com/theseion/crs-toolchain/v2/processors"
)

const (
	preprocessorStartRegex = `\s*##!}>\s*(.*)`
	preprocessorEndRegex   = `\s*##!}<`
)

var regexes = struct {
	preprocessorStart regexp.Regexp
	preprocessorEnd   regexp.Regexp
}{
	preprocessorStart: *regexp.MustCompile(preprocessorStartRegex),
	preprocessorEnd:   *regexp.MustCompile(preprocessorEndRegex),
}

// Create the processor stack
var processorStack = NewProcessorStack()

func NewAssembler(ctx *processors.Context) *Operator {
	a := &Operator{
		name:    "assemble",
		details: make(map[string]string),
		lines:   []string{},
		ctx:     ctx,
		stats:   NewStats(),
	}
	return a
}

func (a *Operator) Run(input string) (string, error) {
	logger.Trace().Msg("Starting assembler")
	assembleParser := parser.NewParser(a.ctx, strings.NewReader(input))
	lines, _ := assembleParser.Parse()
	logger.Trace().Msgf("Parsed lines: %v", lines)
	return a.Assemble(assembleParser, lines)
}

func (a *Operator) Assemble(assembleParser *parser.Parser, input *bytes.Buffer) (string, error) {
	fileScanner := bufio.NewScanner(bytes.NewReader(input.Bytes()))
	fileScanner.Split(bufio.ScanLines)
	var processor processors.IProcessor = processors.NewAssemble(a.ctx)
	processorStack.push(processor)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		logger.Trace().Msgf("parsing line: %q", line)

		if regexes.preprocessorStart.MatchString(line) {
			logger.Trace().Msg("Found processor start")
			assemble := processors.NewAssemble(a.ctx)
			processorStack.push(assemble)
			processor = assemble
		} else if regexes.preprocessorEnd.MatchString(line) {
			logger.Trace().Msg("Found processor end")
			previousProcessor, err := processorStack.pop()
			if err != nil {
				logger.Error().Err(err).Msg("Mismatched end marker, processor stack is empty")
				return "", err
			}

			lines, err := previousProcessor.Complete()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to complete processor")
				return "", err
			}

			processor, err = processorStack.top()
			if err != nil {
				logger.Error().Err(err).Msg("Mismatched end marker, processor stack is empty")
				return "", err
			}
			a.lines = append(a.lines, lines...)
		} else {
			processor.ProcessLine(line)
		}
	}

	processor, err := processorStack.pop()
	if err != nil {
		logger.Error().Err(err).Msg("Mismatched end marker, processor stack is empty")
		return "", err
	}
	lines, err := processor.Complete()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to complete processor")
		return "", err
	}
	a.lines = append(a.lines, lines...)
	return a.complete(assembleParser), nil
}

func (a *Operator) complete(assembleParser *parser.Parser) string {
	flagsPrefix := ""
	if len(assembleParser.Flags) > 0 {
		flags := make([]string, 0, len(assembleParser.Flags))
		for flag := range assembleParser.Flags {
			flags = append(flags, string(flag))
		}
		sort.Strings(flags)
		flagsPrefix = "(?" + strings.Join(flags, "") + ")"
	}

	result := strings.Join(a.lines, "")
	if len(assembleParser.Prefixes) > 0 && len(assembleParser.Suffixes) > 0 && len(result) > 0 {
		result = "(?:" + result + ")"
	}
	prefixes := strings.Join(assembleParser.Prefixes, "")
	suffixes := strings.Join(assembleParser.Suffixes, "")
	result = prefixes + result + suffixes

	if len(result) > 0 {
		result = a.runSimplificationAssembly(result)
		result = a.escapeDoublequotes(result)
		result = a.useHexBackslashes(result)
		result = a.includeVerticalTabInBackslashS(result)
	}

	if len(flagsPrefix) > 0 {
		result = flagsPrefix + result
	}

	return result
}

func (a *Operator) runSimplificationAssembly(input string) string {
	// TODO port from python
	return input
}

func (a *Operator) escapeDoublequotes(input string) string {
	// TODO port from python
	return input
}

func (a *Operator) useHexBackslashes(input string) string {
	// TODO port from python
	return input
}

func (a *Operator) includeVerticalTabInBackslashS(input string) string {
	// TODO port from python
	return input
}
