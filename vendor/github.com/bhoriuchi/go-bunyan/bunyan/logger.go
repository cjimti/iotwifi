package bunyan

import "strings"

// Logger handles writing logs to the appropriate streams.
type Logger struct {
	config       Config
	hostname     string
	streams      []Stream
	staticFields map[string]interface{}
	serializers  map[string]func(value interface{}) interface{}
}

// AddStream dynamically adds a stream to the current logger.
func (l *Logger) AddStream(stream Stream) error {
	if err := stream.init(l.config); err != nil {
		return err
	}
	l.streams = append(l.streams, stream)
	return nil
}

// AddSerializers dynamically adds serializers to the current logger.
func (l *Logger) AddSerializers(serializers map[string]func(value interface{}) interface{}) {
	for key, value := range serializers {
		l.serializers[string(key)] = value
	}
}

// Level dynamically queries the current logging streams. See https://github.com/trentm/node-bunyan#levels
func (l *Logger) Level(args ...interface{}) interface{} {
	if len(args) == 0 {
		// implements log.Level() -> LOG_LEVEL
		lowestLevel := LogLevelInfo
		for _, stream := range l.streams {
			if toLogLevelInt(stream.Level) < toLogLevelInt(lowestLevel) {
				lowestLevel = strings.ToLower(stream.Level)
			}
		}
		return lowestLevel
	} else if typeName(args[0]) == "string" && toLogLevelInt(args[0].(string)) > 0 {
		// implements log.Level(LOG_LEVEL)
		for _, stream := range l.streams {
			stream.Level = strings.ToLower(args[0].(string))
		}
	}
	return nil
}

// Levels dynamically sets and queries the current logging streams. See https://github.com/trentm/node-bunyan#levels
func (l *Logger) Levels(args ...interface{}) interface{} {
	argl := len(args)

	switch argl {
	case 0:
		// implements log.Levels() -> [STREAM_LEVEL_0, ...]
		logLevels := make([]string, 0)
		for _, stream := range l.streams {
			logLevels = append(logLevels, stream.Level)
		}
		return logLevels
	case 1:
		switch typeName(args[0]) {
		case "int":
			// implements log.Levels(STREAM_INDEX)
			if len(l.streams) >= (args[0].(int) + 1) {
				return l.streams[args[0].(int)].Level
			}
			return ""
		case "string":
			// implements log.Levels(STREAM_NAME)
			for _, stream := range l.streams {
				if stream.Name == args[0].(string) {
					return stream.Level
				}
			}
		}
	case 2:
		if typeName(args[1]) == "string" && toLogLevelInt(args[1].(string)) > 0 {
			switch typeName(args[0]) {
			case "int":
				// implements log.Levels(STREAM_INDEX, LOG_LEVEL)
				if len(l.streams) >= (args[0].(int) + 1) {
					l.streams[args[0].(int)].Level = strings.ToLower(args[1].(string))
				}
			case "string":
				// implements log.Levels(STREAM_NAME, LOG_LEVEL)
				for _, stream := range l.streams {
					if stream.Name == args[0].(string) {
						stream.Level = strings.ToLower(args[1].(string))
					}
				}
			}
		}
	}

	return nil
}

// creates a new child logger with extra static fields
func (l *Logger) Child(staticFields map[string]interface{}) Logger {
	newStaticFields := make(map[string]interface{})

	// merge the static fields into the new logger
	for key, field := range l.staticFields {
		newStaticFields[string(key)] = field
	}
	for key, field := range staticFields {
		newStaticFields[string(key)] = field
	}

	logger := Logger{
		config:       l.config,
		hostname:     l.hostname,
		streams:      l.streams,
		staticFields: newStaticFields,
	}
	return logger
}

func (l *Logger) Fatal(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelFatal) {
			_log.write(stream, toLogLevelInt(LogLevelFatal))
		}
	}
}

func (l *Logger) Error(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelError) {
			_log.write(stream, toLogLevelInt(LogLevelError))
		}
	}
}

func (l *Logger) Warn(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelWarn) {
			_log.write(stream, toLogLevelInt(LogLevelWarn))
		}
	}
}

func (l *Logger) Info(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelInfo) {
			_log.write(stream, toLogLevelInt(LogLevelInfo))
		}
	}
}

func (l *Logger) Debug(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelDebug) {
			_log.write(stream, toLogLevelInt(LogLevelDebug))
		}
	}
}

func (l *Logger) Trace(args ...interface{}) {
	_log := bunyanLog{args: args, logger: *l}

	for _, stream := range l.streams {
		if toLogLevelInt(stream.Level) <= toLogLevelInt(LogLevelTrace) {
			_log.write(stream, toLogLevelInt(LogLevelTrace))
		}
	}
}
