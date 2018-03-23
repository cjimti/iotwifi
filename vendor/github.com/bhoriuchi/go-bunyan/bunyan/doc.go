// Package bunyan implements the node.js logging library bunyan in go.
// According to https://github.com/trentm/node-bunyan it is a simple and fast JSON logging library.
//
// See https://github.com/trentm/node-bunyan#log-method-api for log method api
//
// Hello World Example
//
// Create a logger that logs to os.Stdout
//
//    package main
//
//    import (
//        "os"
//        "github.com/bhoriuchi/go-bunyan/bunyan"
//    )
//
//    func main() {
//        config := bunyan.Config{
//            Name: "app",
//            Stream: os.Stdout,
//            Level: bunyan.LogLevelDebug
//        }
//
//        if log, err := bunyan.CreateLogger(config); err == nil {
//            log.Info("Hello %s!", "World")
//        }
//    }
//
// Multi-stream Example
//
// Create a logger that logs to multiple streams
//
//    import (
//        "os"
//        "errors"
//        "github.com/bhoriuchi/go-bunyan/bunyan"
//    )
//
//    func main() {
//        staticFields := make(map[string]interface{})
//        staticFields["foo"] = "bar"
//
//        config := bunyan.Config{
//            Name: "app",
//            Streams: []bunyan.Stream{
//                {
//                    Name: "app-info",
//                    Level: bunyan.LogLevelInfo,
//                    Stream: os.Stdout,
//                },
//                {
//                    Name: "app-errors",
//                    Level: bunyan.LogLevelError,
//                    Path: "/path/to/logs/app-errors.log"
//                },
//            },
//            StaticFields: staticFields,
//        }
//
//        if log, err := bunyan.CreateLogger(config); err == nil {
//            log.Info("Hello %s!", "World")
//            log.Error(errors.New("Foo Failed"), "Foo %s!", "Failed")
//        }
//    }
package bunyan
