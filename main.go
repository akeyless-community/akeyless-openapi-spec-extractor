package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	fetchCmd.Flags().BoolP("validate", "v", true, "Enable validation of OpenAPI spec (optional, default is 'false')")
	fetchCmd.MarkFlagRequired("url")
	fetchCmd.MarkFlagRequired("pattern")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func fetch(cmd *cobra.Command, args []string) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		logrus.Fatal("Error retrieving URL:", err)
	}
	logrus.Debug("URL        : ", url)
	pattern, err := cmd.Flags().GetString("pattern")
	if err != nil {
		logrus.Fatal("Error retrieving pattern:", err)
	}
	logrus.Debug("Pattern    : ", pattern)
	outputType, err := cmd.Flags().GetString("output")
	if err != nil {
		logrus.Fatal("Error retrieving output type:", err)
	}
	logrus.Debug("Output Type: ", outputType)
	loglevel, err := cmd.Flags().GetString("loglevel")
	if err != nil {
		logrus.Fatal("Error retrieving log level:", err)
	}
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		logrus.Fatal("Invalid log level:", err)
	}
	logrus.SetLevel(level)
	validate, err := cmd.Flags().GetBool("validate")
	if err != nil {
		logrus.Fatal("Error retrieving validate flag:", err)
	}
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

	fullApiSpec := data

	result, err := jmespath.Search(pattern, data)
	if err != nil {
		logrus.Fatal(err)
	}

	// Output the result
	printResult(result, specType, outputType, fullApiSpec)
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

// printResult takes a result of type interface{}, a specType indicating the original format of the result,
// and an outputType indicating the desired output format. It attempts to marshal and possibly convert the result
// into the desired format. Errors and debug information are logged, while the final output is printed to stdout.
func printResult(result interface{}, specType string, outputType string, fullApiSpec interface{}) {
	logrus.Debug("Attempting to process result with specType: ", specType, " and outputType: ", outputType)

	var output []byte
	var err error

	processReferences(result, fullApiSpec)

	// Marshal result into its original format (JSON or YAML)
	switch specType {
	case "json":
		output, err = json.Marshal(result)
		if err != nil {
			logrus.Error("Error marshaling result to JSON: ", err)
			return
		}
	case "yaml":
		output, err = yaml.Marshal(result)
		if err != nil {
			logrus.Error("Error marshaling result to YAML: ", err)
			return
		}
	default:
		logrus.Error("Unknown spec type: ", specType)
		return
	}

	// Convert marshaled data to the desired output format if necessary
	if specType != outputType {
		output, err = convertFormat(output, specType, outputType)
		if err != nil {
			logrus.Error("Error converting data format: ", err)
			return
		}
	}

	outputJson := string(output)

	if outputJson != "" {
		fmt.Println(outputJson) // Direct final output to stdout
	} else {
		logrus.Error("No output generated")
	}
}

func processReferences(data interface{}, fullApiSpec interface{}) {
	// Walk through the map and replace "$ref" values
	replaceRefValues(data, fullApiSpec)
}

// replaceRefValues recursively processes the given data structure (v) to replace "$ref" references with their resolved values from the fullApiSpec.
// It handles both map and slice types, diving into nested structures as needed.
func replaceRefValues(v interface{}, fullApiSpec interface{}) {
	// Determine the type of the provided data structure.
	switch data := v.(type) {
	case map[string]interface{}:
		// If the data is a map, iterate over each key-value pair.
		for k, v := range data {
			// Check if the current key is a "$ref" that needs to be resolved.
			if k == "$ref" {
				refValue := data[k].(string)
				logrus.Debugf("Found $ref to resolve: %s", refValue)
				newPattern := transformString(refValue)
				logrus.Debugf("Transformed $ref to JMESPath query: %s", newPattern)

				// Perform a JMESPath search on the fullApiSpec with the transformed query.
				result, err := jmespath.Search(newPattern, fullApiSpec)
				if err != nil {
					logrus.Errorf("Failed to resolve $ref: %s with error: %v", refValue, err)
					continue // Skip this $ref if we can't resolve it.
				}

				logrus.Debugf("Successfully resolved $ref: %s", refValue)

				// Replace "$ref" key with the resolved object.
				delete(data, "$ref") // Remove the "$ref" key.
				if resolvedRef, ok := result.(map[string]interface{}); ok {
					// Update or add each key-value pair from the resolved reference into the original map.
					for key, value := range resolvedRef {
						data[key] = value
					}
				} else {
					logrus.Error("Resolved reference is not of type map[string]interface{}")
				}
			} else if k == "items" && v != nil {
				// Special handling for arrays defined by $ref in their items.
				logrus.Debug("Processing 'items' with $ref")
				itemsMap, ok := v.(map[string]interface{})
				if ok && itemsMap["$ref"] != nil {
					replaceRefValues(itemsMap, fullApiSpec) // Resolve the $ref within the items definition.
				}
			} else if k == "responses" {
				// Explicitly handle response objects.
				logrus.Debug("Processing 'responses' objects")
				if responses, ok := v.(map[string]interface{}); ok {
					for _, response := range responses {
						replaceRefValues(response, fullApiSpec) // Recursively process response objects.
					}
				}
			} else {
				// Recursively process nested maps.
				replaceRefValues(v, fullApiSpec)
			}
		}
	case []interface{}:
		// If the data is a slice, iterate over each element.
		logrus.Debug("Processing slice for $ref values")
		for i := range data {
			replaceRefValues(data[i], fullApiSpec) // Process each element in the slice.
		}
	}
}

// convertFormat takes a byte slice of data, its current format (currentFormat), and the desired format (targetFormat),
// and attempts to convert the data to the target format. It returns the converted data or an error if the conversion fails.
func convertFormat(data []byte, currentFormat string, targetFormat string) ([]byte, error) {
	logrus.Debug("Converting data from ", currentFormat, " to ", targetFormat)
	switch {
	case currentFormat == "json" && targetFormat == "yaml":
		return yaml.JSONToYAML(data)
	case currentFormat == "yaml" && targetFormat == "json":
		return yaml.YAMLToJSON(data)
	default:
		logrus.Error("Unsupported conversion: ", currentFormat, " to ", targetFormat)
		return nil, fmt.Errorf("unsupported conversion from %s to %s", currentFormat, targetFormat)
	}
}

// transformString takes a string formatted like "#/components/responses/errorResponse",
// removes the leading "#/", and then replaces "/" with "."
func transformString(s string) string {
	// Remove the leading "#/"
	s = strings.TrimPrefix(s, "#/")

	// Replace all occurrences of "/" with "."
	s = strings.Replace(s, "/", ".", -1)

	return s
}
