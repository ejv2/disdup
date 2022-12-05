package clconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	config "github.com/ethanv2/disdup/conf"
	"github.com/ethanv2/disdup/output"
)

// Output parsing or processing errors.
var (
	ErrWrongType      = errors.New("unexpected type")
	ErrUnknownCollate = errors.New("unknown collation mode")
)

// An Output is a json-encodable representation of a disdup output.
type Output struct {
	Type      string                 `json:"type"`
	Arguments map[string]interface{} `json:"args"`
}

func parseCollation(conf map[string]interface{}) (int, error) {
	if rcoll, ok := conf["collate"]; ok {
		// Unwrap option into map of options
		coll, ok := rcoll.(string)
		if !ok {
			return 0, fmt.Errorf("key collate: %w", ErrWrongType)
		}

		switch coll {
		case "channel":
			return output.WriterCollateChannel, nil
		case "user":
			return output.WriterCollateUser, nil
		default:
			return 0, fmt.Errorf("%s: %w", coll, ErrUnknownCollate)
		}
	}

	return 0, nil
}

func parseWriter(dest io.WriteCloser, conf map[string]interface{}) (*output.Writer, error) {
	coll, err := parseCollation(conf)
	if err != nil {
		return nil, err
	}

	rprefix, ok := conf["prefix"]
	prefix := ""
	if ok {
		if prefix, ok = rprefix.(string); !ok {
			return nil, fmt.Errorf("key prefix: %w", ErrWrongType)
		}
		prefix += " " // Append a space to properly space output in log format
	}

	w := &output.Writer{
		Output:  dest,
		Prefix:  prefix,
		Collate: coll,
	}
	return w, nil
}

func parseMailer(conf map[string]interface{}) (*output.Mailer, error) {
	ret := &output.Mailer{}

	// Specific keys mapped to non-string values
	// Need to be deleted after use to prevent next loop from using them
	rreply, ok := conf["reply_mode"]
	if ok {
		reply, ok := rreply.(uint)
		if !ok {
			return nil, fmt.Errorf("key reply_mode: %w: expected string", ErrWrongType)
		}

		ret.ReplyMode = reply
		delete(conf, "reply_mode")
	}
	orsrv, ok := conf["server"]
	if ok {
		rsrv, ok := orsrv.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key server: %w: expected object", ErrWrongType)
		}

		for key, rval := range rsrv {
			val, ok := rval.(string)
			if !ok {
				return nil, fmt.Errorf("key server: %w: expected all string values", ErrWrongType)
			}

			switch key {
			case "address":
				ret.Server.Address = val
			case "username":
				ret.Server.Username = val
			case "password":
				ret.Server.Password = val
			}
		}
		delete(conf, "server")
	}

	// Generic keys mapped to string values
	for key, rval := range conf {
		val, ok := rval.(string)
		if !ok {
			return nil, fmt.Errorf("key %s: %w: expected string", key, ErrWrongType)
		}

		switch key {
		case "to":
			ret.To = val
		case "from":
			ret.From = val
		case "preamble":
			ret.Preamble = val
		case "footer":
			ret.Footer = val
		}
	}

	return ret, nil
}

// convertOutput converts a temporary representation of an output to the format
// which can be read by disdup.
func convertOutput(name string, tmpl Output, cfg *config.Config) error {
	var err error
	var out output.Output

	switch tmpl.Type {
	case "stdout":
		out, err = parseWriter(os.Stdout, tmpl.Arguments)
	case "mail":
		out, err = parseMailer(tmpl.Arguments)
	default:
		err = ErrOutput
	}

	if err != nil {
		return err
	}

	cfg.Outputs = append(cfg.Outputs, config.OutputConfig{Name: name, Output: out})
	return nil
}

func loadOutputs(cfg *config.Config, paths []string) error {
	found := false
	var outputs map[string]Output
	for _, dir := range paths {
		path := dir + string(os.PathSeparator) + OutputConfigName
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()

			found = true
			buf, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("outputs config: %w", ErrIO)
			}

			buf = processConfig(buf)
			err = json.Unmarshal(buf, &outputs)
			if err != nil {
				return fmt.Errorf("outputs config: bad syntax: %w", err)
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("outputs config: %w", ErrNotFound)
	}

	for name, output := range outputs {
		if err := convertOutput(name, output, cfg); err != nil {
			return fmt.Errorf("outputs config: output %s (%s): %w", name, output.Type, err)
		}
	}

	return nil
}
