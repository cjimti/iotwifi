package bunyan

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

const LOG_VERSION = 0 // states the current bunyan log specification version

// Config is used to construct a bunyanLogger with one or more logging streams.
type Config struct {
	Name         string                                         // default name to use for streams; required
	Level        string                                         // default log level to use for streams
	Stream       io.Writer                                      // a stream location that implements the io.Writer interface
	Streams      []Stream                                       // an array of Stream configurations
	Serializers  map[string]func(value interface{}) interface{} // a mapping of field names to serialization functions used for those fields
	StaticFields map[string]interface{}                         // a predefined set of fields that will be added to all logs
}

// CreateLogger creates a new bunyanLogger.
// Either a Config or string can be passed as the only argument.
// It returns a new Logger. If no errors were encountered, error will be nil.
func CreateLogger(args ...interface{}) (Logger, error) {
	config := Config{}
	logger := Logger{}
	r := regexp.MustCompile(`bunyan.Config$`)

	if len(args) == 0 {
		return logger, fmt.Errorf("Create logger requires either a bunyan.Config or String argument")
	}

	// get hostname
	if hostname, err := os.Hostname(); err != nil {
		return logger, err
	} else {
		logger.hostname = hostname
	}

	arg := args[0]
	argType := typeName(arg)

	if argType == "string" {
		config.Name = arg.(string)
	} else if r.MatchString(argType) {
		config = arg.(Config)
		if config.Name == "" {
			return logger, fmt.Errorf("Bunyan Config requires a name, none specified")
		}
	} else {
		return logger, fmt.Errorf("Create logger requires either a bunyan.Config or String argument")
	}

	// add the config to the logger
	logger.config = config
	logger.staticFields = config.StaticFields
	logger.serializers = config.Serializers

	// add the streams
	if len(config.Streams) != 0 {
		for _, stream := range config.Streams {
			logger.AddStream(stream)
		}
	} else if config.Stream != nil {
		simpleStream := Stream{Stream: config.Stream, Name: config.Name}
		logger.AddStream(simpleStream)
	}

	return logger, nil
}
