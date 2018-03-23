package bunyan

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type bunyanLog struct {
	args   []interface{}
	logger Logger
}

// serializes a log field
func (l *bunyanLog) serialize(key string, value interface{}) interface{} {
	if fn, ok := l.logger.serializers[key]; ok {
		return fn(value)
	} else if isError(value) {
		return fmt.Sprintf("%v", value)
	} else {
		return value
	}
}

// prints a formatted string using the arguments provided
func (l *bunyanLog) sprintf(args []interface{}) string {
	return fmt.Sprintf(args[0].(string), args[1:]...)
}

// composes a json formatted log string and writes it to the appropriate stream
func (l *bunyanLog) write(stream Stream, level int) error {
	data := make([]byte, 0)
	argl := len(l.args)

	if argl == 0 {
		return nil
	}

	d := make(map[string]interface{})
	d["v"] = 0
	d["level"] = level
	d["name"] = stream.Name
	d["hostname"] = l.logger.hostname
	d["pid"] = os.Getppid()
	d["time"] = nowTimestamp()

	// add static fields first
	for key, value := range l.logger.staticFields {
		if canSetField(key) {
			d[key] = l.serialize(key, value)
		}
	}

	// add passed fields/data last
	if argl == 1 && typeName(l.args[0]) == "string" {
		// if 1 argument that is a string, the string is the msg
		d["msg"] = l.serialize("msg", l.args[0])
	} else if isError(l.args[0]) {
		// if the first argument is an error, set error field with string value of error
		d["error"] = l.serialize("error", l.args[0])
		if argl == 2 {
			d["msg"] = l.serialize("msg", l.args[1])
		} else if argl > 2 {
			d["msg"] = l.serialize("msg", l.sprintf(l.args[1:]))
		}
	} else if isHashMap(l.args[0]) {
		// if the first argument is a hashmap, process its values
		for key, value := range l.args[0].(map[string]interface{}) {
			if canSetField(key) {
				d[key] = l.serialize(key, value)
			}
		}
		if argl == 2 {
			d["msg"] = l.serialize("msg", l.args[1])
		} else if argl > 2 {
			d["msg"] = l.serialize("msg", l.sprintf(l.args[1:]))
		}
	} else {
		d["msg"] = l.serialize("msg", l.sprintf(l.args))
	}

	// marshal the json
	if jsonData, err := json.Marshal(d); err != nil {
		return err
	} else {
		data = []byte(fmt.Sprintf("%s\n", string(jsonData)))
	}

	switch stream.Type {
	case LogTypeStream:
		return l.writeStream(stream, data)
	case LogTypeFile:
		return l.writeFile(stream, data)
	case LogTypeRotatingFile:
		return l.writeRotatingFile(stream, data)
	case LogTypeRaw:
		return l.writeStream(stream, data)
	}
	return nil
}

// writes the data to a stream that implements io.Writer
func (l *bunyanLog) writeStream(stream Stream, data []byte) error {
	stream.Stream.Write(data)
	return nil
}

// writes the data to a log file
func (l *bunyanLog) writeFile(stream Stream, data []byte) error {
	if f, err := os.OpenFile(stream.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Printf("[bunyan] error: %v", err)
	} else if _, err := f.Write(data); err != nil {
		log.Printf("[bunyan] error: %v", err)
	} else if err := f.Close(); err != nil {
		log.Printf("[bunyan] error: %v", err)
	}
	return nil
}

// TODO: implement this
func (l *bunyanLog) writeRotatingFile(stream Stream, data []byte) error {
	return nil
}
