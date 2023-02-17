package api_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tidwall/pretty"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"
	"gitlab.com/rodrigoodhin/gocure/report/html"
	"io"
	"io/fs"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

func generateReport() {
	htmlOutputDir := "reports/"

	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: htmlOutputDir,
			Title:            "Kyma API-Gateway component tests",
			Metadata: models.Metadata{
				Platform:        runtime.GOOS,
				TestEnvironment: "Gardener GCP",
				Parallel:        "Scenarios",
				Executed:        "Remote",
				AppVersion:      "main",
				Browser:         "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
		if path == "reports" {
			return nil
		}

		data, err1 := os.ReadFile(path)
		if err1 != nil {
			return err
		}

		//Format all patterns like "&lt" to not be replaced later
		find := regexp.MustCompile(`&\w\w`)
		formatted := find.ReplaceAllFunc(data, func(b []byte) []byte {
			return []byte{b[0], ' ', b[1], b[2]}
		})

		err = os.WriteFile(path, formatted, fs.FileMode(02))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf(err.Error())
	}

	if artifactsDir, ok := os.LookupEnv("ARTIFACTS"); ok {
		err = filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
			if path == "reports" {
				return nil
			}

			_, err1 := copy(path, fmt.Sprintf("%s/report.html", artifactsDir))
			if err1 != nil {
				return err1
			}
			return nil
		})

		if err != nil {
			log.Fatalf(err.Error())
		}

		_, err = copy("./junit-report.xml", fmt.Sprintf("%s/junit-report.xml", artifactsDir))
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func getPodListReport() string {
	type returnedPodList struct {
		PodList []struct {
			Metadata struct {
				Name              string `json:"name"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	list, _ := k8sClient.Resource(res).Namespace("").List(context.Background(), v1.ListOptions{})

	p := returnedPodList{}
	toMarshal, _ := json.Marshal(list)
	err := json.Unmarshal(toMarshal, &p)
	if err != nil {
		log.Fatalf(err.Error())
	}
	toPrint, _ := json.Marshal(p)
	return string(pretty.Pretty(toPrint))
}
