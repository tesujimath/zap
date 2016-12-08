// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zap_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

func Example() {
	// Log in JSON, using zap's reflection-free JSON encoder. By default, loggers
	// write all InfoLevel and above logs to standard out.
	logger := zap.Neo(zap.WriterFacility(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		nil, // defaults to standard out
		zap.InfoLevel,
	))

	logger.Warn("Log without structured data...")
	logger.Warn(
		"Or use strongly-typed wrappers to add structured context.",
		zap.String("library", "zap"),
		zap.Duration("latency", time.Nanosecond),
	)

	// Avoid re-serializing the same data repeatedly by creating a child logger
	// with some attached context. That context is added to all the child's
	// log output, but doesn't affect the parent.
	child := logger.With(
		zap.String("user", "jane@test.com"),
		zap.Int("visits", 42),
	)
	child.Error("Oh no!")

	// Output:
	// {"level":"warn","msg":"Log without structured data..."}
	// {"level":"warn","msg":"Or use strongly-typed wrappers to add structured context.","library":"zap","latency":1}
	// {"level":"error","msg":"Oh no!","user":"jane@test.com","visits":42}
}

func Example_fileOutput() {
	// Create a temporary file to output logs to.
	f, err := ioutil.TempFile("", "log")
	if err != nil {
		panic("failed to create temporary file")
	}
	defer os.Remove(f.Name())

	logger := zap.Neo(
		// Write the logging output to the specified file instead of stdout.
		// Any type implementing zap.Facility can be used.
		zap.WriterFacility(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			f,
			zap.InfoLevel,
		),
	)

	logger.Info("This is an info log.", zap.Int("foo", 42))

	// Sync the file so logs are written to disk, and print the file contents.
	// zap will call Sync automatically when logging at FatalLevel or PanicLevel.
	f.Sync()
	contents, err := ioutil.ReadFile(f.Name())
	if err != nil {
		panic("failed to read temporary file")
	}

	fmt.Println(string(contents))
	// Output:
	// {"level":"info","msg":"This is an info log.","foo":42}
}

func ExampleNest() {
	logger := zap.Neo(zap.WriterFacility(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		nil, // defaults to standard out
		zap.InfoLevel,
	))

	// We'd like the logging context to be {"outer":{"inner":42}}
	nest := zap.Nest("outer", zap.Int("inner", 42))
	logger.Info("Logging a nested field.", nest)

	// Output:
	// {"level":"info","msg":"Logging a nested field.","outer":{"inner":42}}
}

func ExampleNew() {
	// The default logger outputs to standard out and only writes logs that are
	// Info level or higher.
	logger := zap.Neo(zap.WriterFacility(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		nil, // defaults to standard out
		zap.InfoLevel,
	))

	// The default logger does not print Debug logs.
	logger.Debug("This won't be printed.")
	logger.Info("This is an info log.")

	// Output:
	// {"level":"info","msg":"This is an info log."}
}

func ExampleNew_textEncoder() {
	// For more human-readable output in the console, use a TextEncoder.
	textLogger := zap.Neo(zap.WriterFacility(
		zap.NewTextEncoder(zap.TextNoTime()), // drop timestamps in tests.
		nil, // defaults to standard out
		zap.InfoLevel,
	))

	textLogger.Info("This is a text log.", zap.Int("foo", 42))

	// Output:
	// [I] This is a text log. foo=42
}

func ExampleTee() {
	// Multiple loggers can be combine using Tee.
	output := zap.Output(os.Stdout)
	logger := zap.Tee(
		zap.New(zap.NewTextEncoder(zap.TextNoTime()), output),
		zap.New(zap.NewJSONEncoder(zap.NoTime()), output),
	)

	logger.Info("this log gets encoded twice, differently", zap.Int("foo", 42))
	// Output:
	// [I] this log gets encoded twice, differently foo=42
	// {"level":"info","msg":"this log gets encoded twice, differently","foo":42}
}

func ExampleMultiWriteSyncer() {
	// To send output to multiple outputs, use MultiWriteSyncer.
	textLogger := zap.Neo(zap.WriterFacility(
		zap.NewTextEncoder(zap.TextNoTime()),
		zap.MultiWriteSyncer(os.Stdout, os.Stdout),
		zap.InfoLevel,
	))

	textLogger.Info("One becomes two")
	// Output:
	// [I] One becomes two
	// [I] One becomes two
}

func ExampleNew_options() {
	// We can pass multiple options to the New method to configure the logging
	// level, output location, or even the initial context.
	logger := zap.Neo(
		zap.WriterFacility(
			zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
			nil, // defaults to standard out
			zap.DebugLevel,
		),
		zap.Fields(zap.Int("count", 1)),
	)

	logger.Debug("This is a debug log.")
	logger.Info("This is an info log.")

	// Output:
	// {"level":"debug","msg":"This is a debug log.","count":1}
	// {"level":"info","msg":"This is an info log.","count":1}
}

func ExampleCheckedMessage() {
	logger := zap.Neo(zap.WriterFacility(
		zap.NewJSONEncoder(zap.NoTime()), // drop timestamps in tests
		nil, // defaults to standard out
		zap.InfoLevel,
	))

	// By default, the debug logging level is disabled. However, calls to
	// logger.Debug will still allocate a slice to hold any passed fields.
	// Particularly performance-sensitive applications can avoid paying this
	// penalty by using checked messages.
	if cm := logger.Check(zap.DebugLevel, "This is a debug log."); cm.OK() {
		// Debug-level logging is disabled, so we won't get here.
		cm.Write(zap.Int("foo", 42), zap.Stack())
	}

	if cm := logger.Check(zap.InfoLevel, "This is an info log."); cm.OK() {
		// Since info-level logging is enabled, we expect to write out this message.
		cm.Write()
	}

	// Output:
	// {"level":"info","msg":"This is an info log."}
}

func ExampleLevel_MarshalText() {
	level := zap.ErrorLevel
	s := struct {
		Level *zap.Level `json:"level"`
	}{&level}
	bytes, _ := json.Marshal(s)
	fmt.Println(string(bytes))

	// Output:
	// {"level":"error"}
}

func ExampleLevel_UnmarshalText() {
	var s struct {
		Level zap.Level `json:"level"`
	}
	// The zero value for a zap.Level is zap.InfoLevel.
	fmt.Println(s.Level)

	json.Unmarshal([]byte(`{"level":"error"}`), &s)
	fmt.Println(s.Level)

	// Output:
	// info
	// error
}

func ExampleNewJSONEncoder() {
	// An encoder with the default settings.
	zap.NewJSONEncoder()

	// Dropping timestamps is often useful in tests.
	zap.NewJSONEncoder(zap.NoTime())

	// In production, customize the encoder to work with your log aggregation
	// system.
	zap.NewJSONEncoder(
		zap.RFC3339Formatter("@timestamp"), // human-readable timestamps
		zap.MessageKey("@message"),         // customize the message key
		zap.LevelString("@level"),          // stringify the log level
	)
}

func ExampleNewTextEncoder() {
	// A text encoder with the default settings.
	zap.NewTextEncoder()

	// Dropping timestamps is often useful in tests.
	zap.NewTextEncoder(zap.TextNoTime())

	// If you don't like the default timestamp formatting, choose another.
	zap.NewTextEncoder(zap.TextTimeFormat(time.RFC822))
}
