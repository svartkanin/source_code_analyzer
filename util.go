package main

import (
	enc_json "encoding/json"
	"flag"
	"fmt"
	"github.com/gookit/config"
	"github.com/gookit/config/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// parse the command line parameters
func parseArgs() (string, string) {
	config := flag.String("config", "", "Path to the configuration file")
	outputFile := flag.String("output", "", "Result output file")
	flag.Parse()

	if *config == "" {
		fmt.Println("No configuration specifiede")
		os.Exit(0)
	} else if !check_file_exists(*config) {
		fmt.Printf("Configuration doesn't exist: %s\n", *config)
		os.Exit(0)
	}

	return *config, *outputFile
}

// check if a given file path exists
func check_file_exists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// read the configuration file
func readConfiguration(config_file string) map[string]interface{} {
	config.AddDriver(json.Driver)

	if err := config.LoadFiles(config_file); err != nil {
		panic(err)
	}

	return config.Data()
}

// Get all files from the directory with a specific file extension
func filesWithExtension(ext string, codeRepoDir string) []string {
	var files []string

	filepath.Walk(codeRepoDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ext {
				files = append(files, path)
			}
		}
		return nil
	})
	return files
}

// convert an interface array to a string array
func convertArray(arr []interface{}) []string {
	s := make([]string, len(arr))
	for i, v := range arr {
		s[i] = fmt.Sprint(v)
	}
	return s
}

// convert AnalysesResults object to JSON and print it
func outputResults(analysesResults *AnalysisResults, outputFile string) {
	jsonified, err := enc_json.MarshalIndent(analysesResults, "", "  ")

	if err == nil {
		if outputFile == "" {
			log.Println(string(jsonified))
		} else {
			err := ioutil.WriteFile(outputFile, jsonified, 0644)

			if err != nil {
				panic(err)
			}
			log.Printf("Output written to file: %s", outputFile)
		}
	}

}
