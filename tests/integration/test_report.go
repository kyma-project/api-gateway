package api_gateway

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"
	"gitlab.com/rodrigoodhin/gocure/report/html"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func generateReport(ts testcontext.Testsuite) {
	htmlOutputDir := "reports/"

	h := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    "cucumber-report.json",
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
	err := h.Generate()
	if err != nil {
		log.Fatal(err.Error())
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
		log.Fatal(err.Error())
	}

	if artifactsDir, ok := os.LookupEnv("ARTIFACTS"); ok {
		err = filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
			if path == "reports" {
				return nil
			}

			_, err1 := copyReport(path, fmt.Sprintf("%s/report-%s.html", artifactsDir, ts.Name()))
			if err1 != nil {
				return err1
			}
			return nil
		})

		if err != nil {
			log.Fatal(err.Error())
		}

		_, err = copyReport("./junit-report.xml", fmt.Sprintf("%s/junit-report-%s.xml", artifactsDir, ts.Name()))
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func copyReport(src, dst string) (int64, error) {
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
	defer func() {
		if tempErr := source.Close(); tempErr != nil {
			err = tempErr
		}
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer func() {
		if tempErr := destination.Close(); tempErr != nil {
			err = tempErr
		}
	}()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
