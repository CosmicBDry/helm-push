package tools

import (
	"gopkg.in/yaml.v3"
)

func YamlUnmarshalMap(in_stream []byte, outMap map[string]interface{}) error {

	err := yaml.Unmarshal(in_stream, &outMap)

	return err
}

func YamlMarshalMap(outMap map[string]any) ([]byte, error) {
	out_stream, err := yaml.Marshal(outMap)
	if err != nil {
		return nil, err
	}
	return out_stream, nil

}
