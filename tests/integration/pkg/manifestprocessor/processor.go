package manifestprocessor

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

func getManifestsFromFile(fileName string, directory string, separator string) ([]string, error) {
	data, err := os.ReadFile(path.Join(directory, fileName))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), separator), nil
}

func convert(inputYAML []string) ([]string, error) {
	var result []string
	for _, input := range inputYAML {
		json, err := yaml.YAMLToJSON([]byte(input))
		if err != nil {
			return nil, err
		}
		if string(json) != "null" {
			result = append(result, string(json))
		}
	}
	return result, nil
}

func parseTemplateWithData(templateRaw string, data interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Parse(templateRaw)
	if err != nil {
		return "", err
	}
	var resource bytes.Buffer
	err = tmpl.Execute(&resource, data)
	if err != nil {
		return "", err
	}
	return resource.String(), nil
}

// ParseFromFile parse simple yaml manifests
func ParseFromFile(fileName string, directory string, separator string) ([]unstructured.Unstructured, error) {
	manifestArray, err := getManifestsFromFile(fileName, directory, separator)
	if err != nil {
		return nil, err
	}
	manifests, err := convert(manifestArray)
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, man := range manifests {
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}

// ParseFromFileWithTemplate parse manifests with goTemplate support
func ParseFromFileWithTemplate(fileName string, directory string, separator string, templateData interface{}) ([]unstructured.Unstructured, error) {
	rawData, err := os.ReadFile(path.Join(directory, fileName))
	if err != nil {
		return nil, err
	}

	man, err := parseTemplateWithData(string(rawData), templateData)
	if err != nil {
		return nil, err
	}

	manifestArray := strings.Split(man, separator)
	manifests, err := convert(manifestArray)
	if err != nil {
		return nil, err
	}

	var resources []unstructured.Unstructured
	for _, man := range manifests {
		res, err := parseManifest([]byte(man))
		if err != nil {
			return nil, err
		}
		resources = append(resources, *res)
	}
	return resources, nil
}
