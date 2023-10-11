package manifestprocessor

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"text/template"

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

func ParseWithTemplate(manifest []byte, templateData interface{}) ([]unstructured.Unstructured, error) {
	man, err := parseTemplateWithData(string(manifest), templateData)
	if err != nil {
		return nil, err
	}

	var manifests []unstructured.Unstructured
	decoder := yaml.NewDecoder(bytes.NewBufferString(man))
	for {
		var d map[string]interface{}
		if err := decoder.Decode(&d); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Document decode failed: %w", err)
		}
		manifests = append(manifests, unstructured.Unstructured{Object: d})
	}

	return manifests, nil
}

func ParseYamlFromFile(fileName string, directory string) ([]unstructured.Unstructured, error) {
	rawData, err := os.ReadFile(path.Join(directory, fileName))
	if err != nil {
		return nil, err
	}

	var manifests []unstructured.Unstructured
	decoder := yaml.NewDecoder(bytes.NewBufferString(string(rawData)))
	for {
		var d map[string]interface{}
		if err := decoder.Decode(&d); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Document decode failed: %w", err)
		}
		manifests = append(manifests, unstructured.Unstructured{Object: d})
	}
	return manifests, nil
}
