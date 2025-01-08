package manifestprocessor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"text/template"

	"gopkg.in/yaml.v3"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

// ParseFromFileWithTemplate parse manifests with goTemplate support
func ParseFromFileWithTemplate(fileName string, directory string, templateData interface{}) ([]unstructured.Unstructured, error) {
	rawData, err := os.ReadFile(path.Join(directory, fileName))
	if err != nil {
		return nil, err
	}

	return ParseWithTemplate(rawData, templateData)
}

func ParseSingleEntryFromFileWithTemplate(fileName string, directory string, templateData interface{}) (unstructured.Unstructured, error) {
	result, err := ParseFromFileWithTemplate(fileName, directory, templateData)
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	if len(result) > 1 {
		return unstructured.Unstructured{}, fmt.Errorf("Template in file %s contains more than one entry", fileName)
	}
	return result[0], nil
}

func ParseWithTemplate(manifest []byte, templateData interface{}) ([]unstructured.Unstructured, error) {
	man, err := parseTemplateWithData(string(manifest), templateData)
	if err != nil {
		return nil, err
	}

	return parse(bytes.NewBufferString(man))
}

func ParseYamlFromFile(fileName string, directory string) ([]unstructured.Unstructured, error) {
	rawData, err := os.ReadFile(path.Join(directory, fileName))
	if err != nil {
		return nil, err
	}

	return ParseYaml(rawData)
}

func ParseYaml(rawYaml []byte) ([]unstructured.Unstructured, error) {
	return parse(bytes.NewBufferString(string(rawYaml)))
}

func parse(buffer *bytes.Buffer) ([]unstructured.Unstructured, error) {
	var manifests []unstructured.Unstructured
	decoder := yaml.NewDecoder(buffer)
	for {
		var d map[string]interface{}
		if err := decoder.Decode(&d); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("document decode failed: %w", err)
		}
		manifests = append(manifests, unstructured.Unstructured{Object: d})
	}
	return manifests, nil
}
