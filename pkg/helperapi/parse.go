package helperapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Unmarshals a file into the provided struct pointer.
// Returns an error if unmarshalling fails.
// Returns an error if the file type is not recognized.
func (api *Api) UnmarshalFile(file string, data any) error {
	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return fmt.Errorf("data must be pointer")
	}
	api.Logger.Info("unmarshal file", "path", file)
	if strings.HasSuffix(file, ".json") {
		return api.unmarshalJsonFile(file, data)
	} else if strings.HasSuffix(file, ".xml") {
		return api.unmarshalXmlFile(file, data)
	} else {
		return fmt.Errorf("unrecognized file type %s", file)
	}
}

// Marshals data into the given file
// Returns an error if marshalling fails.
// Returns an error if the file type is not recognized.
func (api *Api) MarshalFile(data any, file string) error {
	api.Logger.Info("marshal file", "path", file)
	if strings.HasSuffix(file, ".json") {
		return api.marshalJsonFile(data, file)
	} else if strings.HasSuffix(file, ".xml") {
		return api.marshalXmlFile(data, file)
	} else {
		return fmt.Errorf("unrecognized file type %s", file)
	}
}

// Unmarshals a JSON file into the provided struct pointer.
// Returns an error if the file is unreadable.
// Returns an error if the file's contents is not JSON encoded.
func (api *Api) unmarshalJsonFile(file string, data any) error {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileBytes, data)
	if err != nil {
		return err
	}

	return nil
}

// Marshals data into JSON and writes it to the given file.
// Returns an error if the data could not be JSON encoded.
// Returns an error if the file is not writeable
func (api *Api) marshalJsonFile(data any, file string) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, dataBytes, 0755)
	if err != nil {
		return err
	}

	return nil
}

// Unmarshals an XML file into the provided struct pointer.
// Returns an error if the file is unreadable.
// Returns an error if the file's contents is not XML encoded.
func (api *Api) unmarshalXmlFile(file string, data any) error {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(fileBytes, data)
	if err != nil {
		return err
	}

	return nil
}

// Marshals data into XML and writes it to the given file.
// Returns an error if the data could not be XML encoded.
// Returns an error if the file is not writeable
func (api *Api) marshalXmlFile(data any, file string) error {
	dataBytes, err := xml.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, dataBytes, 0755)
	if err != nil {
		return err
	}

	return nil
}
