package helper

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Parses environment variables into the provided struct pointer.
// Returns an error if parsing the environment variables fail.
// See: [env.Parse]
func ParseEnv(ctx context.Context, cfg any) error {
	Logger(ctx).Info("parse env", "type", fmt.Sprintf("%T", cfg))
	return env.Parse(cfg)
}

// Marshals data into the given file
// Returns an error if marshalling fails.
// Returns an error if the file type is not recognized.
func MarshalFile(ctx context.Context, data any, file string) error {
	Logger(ctx).Info("marshal file", "path", file)
	if strings.HasSuffix(file, ".json") {
		return marshalJsonFile(ctx, data, file)
	} else if strings.HasSuffix(file, ".xml") {
		return marshalXmlFile(ctx, data, file)
	} else {
		return fmt.Errorf("unrecognized file type %s", file)
	}
}

// Marshals data into JSON and writes it to the given file.
// Returns an error if the data could not be JSON encoded.
// Returns an error if the file is not writeable
func marshalJsonFile(ctx context.Context, data any, file string) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(file, dataBytes, 0755)
}

// Marshals data into XML and writes it to the given file.
// Returns an error if the data could not be XML encoded.
// Returns an error if the file is not writeable
func marshalXmlFile(ctx context.Context, data any, file string) error {
	dataBytes, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(file, dataBytes, 0755)
}

// Unmarshals a file into the provided struct pointer.
// Returns an error if unmarshalling fails.
// Returns an error if the file type is not recognized.
func UnmarshalFile(ctx context.Context, file string, data any) error {
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return fmt.Errorf("data must be pointer")
	}
	Logger(ctx).Info("unmarshal file", "path", file)
	if strings.HasSuffix(file, ".json") {
		return unmarshalJsonFile(ctx, file, data)
	} else if strings.HasSuffix(file, ".xml") {
		return unmarshalXmlFile(ctx, file, data)
	} else {
		return fmt.Errorf("unrecognized file type %s", file)
	}
}

// Unmarshals a JSON file into the provided struct pointer.
// Returns an error if the file is unreadable.
// Returns an error if the file's contents is not JSON encoded.
func unmarshalJsonFile(ctx context.Context, file string, data any) error {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(fileBytes, data)
}

// Unmarshals an XML file into the provided struct pointer.
// Returns an error if the file is unreadable.
// Returns an error if the file's contents is not XML encoded.
func unmarshalXmlFile(ctx context.Context, file string, data any) error {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return xml.Unmarshal(fileBytes, data)
}
