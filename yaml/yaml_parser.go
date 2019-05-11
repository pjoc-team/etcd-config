package yaml

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

import (
	"gopkg.in/yaml.v2"
)

// Check if the file or directory is existed
func IsNotExist(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return true, err
		}
	}

	return false, nil
}

// Unmarshal decodes the first document found within the in byte slice
// and assigns decoded values into the out value.
func Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

// Marshal serializes the value provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
// Maps and pointers (to struct, string, int, etc) are accepted as the in value.
func Marshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

// Unmarshal decodes the first document found within the in byte slice
// and assigns decoded values into the out value.
func UnmarshalFromFile(name string, out interface{}) error {
	if out == nil {
		return errors.New("out is nil")
	}

	// Check if the file is existed
	if not_existed, err := IsNotExist(name); not_existed {
		return err
	}

	// Read the whole file
	stream, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	// Unmarshal the yaml stream
	return Unmarshal(stream, out)
}

// Marshal serializes the value provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
// Maps and pointers (to struct, string, int, etc) are accepted as the in value.
func MarshalToFile(in interface{}, name string) error {
	if in == nil {
		return errors.New("in is nil")
	}

	// Check is the directory of file is existed
	dir := filepath.Dir(name)
	if not_existed, err := IsNotExist(dir); not_existed {
		return err
	}

	// Marshal the object
	stream, err := Marshal(in)
	if err != nil {
		return err
	}

	// Write the marshalled bytes to file
	return ioutil.WriteFile(name, stream, 0644)
}
