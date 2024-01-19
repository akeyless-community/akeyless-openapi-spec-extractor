package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/jmespath/go-jmespath"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "apispec",
		Short: "API Spec Extractor",
	}
	fetchCmd = &cobra.Command{
		Use:   "fetch --url [url] --pattern [jmespath_pattern] --output [output_type]",
		Short: "Fetch and process OpenAPI spec",
		Run:   fetch,
	}
)

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().StringP("url", "u", "", "URL of the OpenAPI spec (required)")
	fetchCmd.Flags().StringP("pattern", "p", "", "JMESPath pattern (required) : https://jmespath.org/tutorial.html")
	fetchCmd.Flags().StringP("output", "o", "json", "Output type (optional, default is 'json'): 'json' or 'yaml'")
	fetchCmd.Flags().StringP("loglevel", "l", "debug", "Logging level (optional, default is 'debug'): 'panic', 'fatal', 'error', 'warn', 'info', 'debug', 'trace'")
	fetchCmd.Flags().BoolP("validate", "v", false, "Enable validation of OpenAPI spec (optional, default is 'false')")
	fetchCmd.MarkFlagRequired("url")
	fetchCmd.MarkFlagRequired("pattern")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func fetch(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	logrus.Debug("URL        : ", url)
	pattern, _ := cmd.Flags().GetString("pattern")
	logrus.Debug("Pattern    : ", pattern)
	outputType, _ := cmd.Flags().GetString("output")
	logrus.Debug("Output Type: ", outputType)
	loglevel, _ := cmd.Flags().GetString("loglevel")
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		logrus.Fatal("Invalid log level:", err)
	}
	logrus.SetLevel(level)
	validate, _ := cmd.Flags().GetBool("validate")
	logrus.Debug("Validation : ", validate)

	// Fetch OpenAPI spec
	specBytes, specType, err := fetchOpenAPISpec(url)
	if err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Debug("SpecType :", specType)
		size := len(specBytes)
		if size < 1024 {
			logrus.Debug("Size of OpenAPI spec: ", size, " Bytes")
		} else if size < 1048576 {
			logrus.Debug("Size of OpenAPI spec: ", size/1024, " KB")
		} else {
			logrus.Debug("Size of OpenAPI spec: ", size/1048576, " MB")
		}
	}

	// Validate OpenAPI spec
	if validate {
		if err = validateOpenAPISpec(specBytes); err != nil {
			logrus.Fatal("Invalid OpenAPI spec:", err)
		}
	}

	// Process with JMESPath
	var data interface{}
	// Unmarshal based on specType
	switch specType {
	case "json":
		if err := json.Unmarshal(specBytes, &data); err != nil {
			logrus.Fatal("Failed to unmarshal JSON OpenAPI Spec:", err)
		}
	case "yaml":
		jsonBytes, err := yaml.YAMLToJSON(specBytes)
		if err != nil {
			logrus.Fatal("Failed to convert YAML to JSON:", err)
		}
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			logrus.Fatal("Failed to unmarshal YAML OpenAPI Spec:", err)
		}
	default:
		logrus.Fatal("Unsupported spec type: ", specType)
	}

	result, err := jmespath.Search(pattern, data)
	if err != nil {
		logrus.Fatal(err)
	}

	// Output the result
	printResult(result, specType, outputType)
}

func fetchOpenAPISpec(url string) ([]byte, string, error) {
	// Create a custom client to avoid SSL checks and follow redirects
	// Set the timeout to 4 minutes
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Minute * 4,
	}

	// Make the request
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Parse the Content-Type header
	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		logrus.Warn("Failed to parse Content-Type:", err)
		// Optionally handle the error or set a default mediaType
	}

	logrus.Debug("Content-Type: ", contentType)

	// Determine the format based on the media type
	var format string
	if strings.HasPrefix(mediaType, "application/json") || strings.HasPrefix(mediaType, "text/json") {
		format = "json"
	} else if strings.HasPrefix(mediaType, "application/yaml") || strings.HasPrefix(mediaType, "text/yaml") {
		format = "yaml"
	} else {
		format = "unknown"
	}

	if format == "unknown" {
		// Attempt to figure out the format by the file prefix of the url
		if strings.HasSuffix(url, ".json") {
			format = "json"
		} else if strings.HasSuffix(url, ".yaml") || strings.HasSuffix(url, ".yml") {
			format = "yaml"
		}
	}

	return body, format, nil
}

func validateOpenAPISpec(specBytes []byte) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specBytes)
	if err != nil {
		return err
	}

	// Validate the OpenAPI spec
	if err = doc.Validate(loader.Context); err != nil {
		return err
	}
	return nil
}

func printResult(result interface{}, specType string, outputType string) {
	var output []byte
	var err error

	switch specType {
	case "json":
		if outputType == "yaml" {
			output, err = yaml.JSONToYAML(result.([]byte))
			if err != nil {
				logrus.Error("Error converting JSON to YAML: ", err)
				return // or handle the error as required
			}
		} else {
			output = result.([]byte)
		}
	case "yaml":
		if outputType == "json" {
			output, err = yaml.YAMLToJSON(result.([]byte))
			if err != nil {
				logrus.Error("Error converting YAML to JSON: ", err)
				return // or handle the error as required
			}
		} else {
			output = result.([]byte)
		}
	default:
		logrus.Error("Unknown spec type: ", specType)
		return
	}

	if output != nil {
		logrus.Info(string(output))
	} else {
		logrus.Error("No output generated")
	}
}
