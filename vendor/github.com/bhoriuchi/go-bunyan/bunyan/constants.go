package bunyan

const LogTypeStream = "stream"              // Writes logs to an io.Writer interface
const LogTypeFile = "file"                  // Writes logs to a location on the filesystem
const LogTypeRotatingFile = "rotating-file" // Writes logs to a location on the filesystem and rotates them
const LogTypeRaw = "raw"                    // Writes logs to a custom writer implementing the io.Writer interface

const LogLevelFatal = "fatal"
const LogLevelError = "error"
const LogLevelWarn = "warn"
const LogLevelInfo = "info"
const LogLevelDebug = "debug"
const LogLevelTrace = "trace"
