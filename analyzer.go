package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

type FileResult struct {
	language  string
	filename  string
	Lines     int
	CodeLines int
	Comments  int
	Keywords  int
}

type LanguageResult struct {
	TotalFiles   int
	TotalResults FileResult
	Files        SingleFileResult
}

type AnalysisResults map[string]*LanguageResult
type SingleFileResult map[string]*FileResult

func main() {
	log.Println("Read configuration")

	// parse the input arguments for the configuration file
	configFile, outputFile := parseArgs()
	// read the JSON configuration file
	configuration := readConfiguration(configFile)
	// start analyzing the code directory
	analysisResults := analyzeCodeDirectory(configuration)

	outputResults(analysisResults, outputFile)
}

// start a parallel process to handle incoming results
func startResultProc(analysisResults *AnalysisResults) (chan FileResult, *sync.WaitGroup) {
	resultProcWaitGroup := &sync.WaitGroup{}
	resultProcChannel := make(chan FileResult)

	resultProcWaitGroup.Add(1)

	go resultProc(analysisResults, resultProcChannel, resultProcWaitGroup)

	return resultProcChannel, resultProcWaitGroup
}

// process incoming results for single files and store them in a combined result object
func resultProc(analysisResults *AnalysisResults,
	resultProcChannel chan FileResult,
	resultProcWaitGroup *sync.WaitGroup) {

	for result := range resultProcChannel {
		if (*analysisResults)[result.language] == nil {
			(*analysisResults)[result.language] = &LanguageResult{}
			(*analysisResults)[result.language].TotalResults.language = result.language
			(*analysisResults)[result.language].Files = make(SingleFileResult)
		}

		langResult := (*analysisResults)[result.language]
		langResult.TotalFiles += 1

		// update the total result values
		langResult.TotalResults.Lines += result.Lines
		langResult.TotalResults.Comments += result.Comments
		langResult.TotalResults.CodeLines += result.CodeLines
		langResult.TotalResults.Keywords += result.Keywords

		// store the result on a file basis
		langResult.Files[result.filename] = &FileResult{
			Lines:     result.Lines,
			Comments:  result.Comments,
			CodeLines: result.CodeLines,
			Keywords:  result.Keywords}
	}

	resultProcWaitGroup.Done()
}

// Start analyzing the code directory specified in the configuration file
// This is done in a parallel way per language per file
func analyzeCodeDirectory(configuration map[string]interface{}) *AnalysisResults {
	log.Println("Start analyzing...")

	codeRepoDir, succ1 := configuration["source_dir"].(string)

	if !check_file_exists(codeRepoDir) {
		log.Printf("Specified source directory doesn't exist '%s'\n", codeRepoDir)
		os.Exit(0)
	}

	languages, succ2 := configuration["languages"].(map[string]interface{})

	if !succ1 || !succ2 {
		panic("Error parsing configuration")
	}

	// channelLang, waitGroupLang := createChannelLang(len(languages))
	analysisResults := &AnalysisResults{}
	resultProcChannel, resultProcWaitGroup := startResultProc(analysisResults)

	langWaitGroup := &sync.WaitGroup{}
	langWaitGroup.Add(len(languages))

	for lang, conf := range languages {
		if conf, success := conf.(map[string]interface{}); success {
			log.Printf("Processing language: %s\n", lang)

			file_extension := conf["file_extension"].(string)

			// Start parallel processing of languages
			go func(lang string, conf map[string]interface{}, codeRepoDir string) {
				files := filesWithExtension(file_extension, codeRepoDir)

				processFiles(resultProcChannel, langWaitGroup, lang, files, conf)
			}(lang, conf, codeRepoDir)
		} else {
			panic("Error reading configuration")
		}
	}

	langWaitGroup.Wait()
	close(resultProcChannel)

	resultProcWaitGroup.Wait()

	log.Println("Done analyzing")

	return analysisResults
}

// process the files for a specific language in parallel
func processFiles(resultProcChannel chan FileResult,
	langWaitGroup *sync.WaitGroup,
	lang string,
	files []string,
	conf map[string]interface{}) {

	var waitGroupFiles sync.WaitGroup
	semaphor := make(chan struct{}, 12)

	waitGroupFiles.Add(len(files))

	// evaluate the regex expression for the comments
	commentRegex := regexp.MustCompile(conf["comments"].(string))

	// evaluate the regex expression for the keywords
	keywords := convertArray(conf["keywords"].([]interface{}))
	keywordsStr := "\\b(" + strings.Join(keywords, "|") + ")\\b"
	keywordsRegex := regexp.MustCompile(keywordsStr)

	// analyze each file in parallel with a max number of parallel processes though (12)
	// since it could lead to too many open file pointers per program
	for _, fileName := range files {
		go processFile(semaphor, &waitGroupFiles, fileName, resultProcChannel, lang, commentRegex, keywordsRegex)
	}

	// wait until all files have been processed
	// then close the file channel and the curent
	// language channel
	go func() {
		waitGroupFiles.Wait()
		langWaitGroup.Done()
		close(semaphor)
	}()
}

func processFile(semaphor chan struct{},
	waitGroupFiles *sync.WaitGroup,
	fileName string,
	resultProcChannel chan FileResult,
	lang string,
	commentRegex *regexp.Regexp,
	keywordsRegex *regexp.Regexp) {
	// create a lock while processing files; this is necessary
	// so that not all files are read at the same time which
	// will exceed the system max open file pointers
	semaphor <- struct{}{} // Lock

	defer func() {
		<-semaphor // Unlock
	}()
	defer waitGroupFiles.Done()

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()

	resultFile := analyzeFile(lines, commentRegex, keywordsRegex)
	resultFile.language = lang
	resultFile.filename = fileName

	resultProcChannel <- resultFile
}

// analyze the content of a single file
func analyzeFile(lines []string,
	commentRegex *regexp.Regexp,
	keywordsRegex *regexp.Regexp) FileResult {

	// store the results for this file
	result := FileResult{}

	// determine number of lines and code lines
	result.Lines += len(lines)

	numberofCodeLines := 0
	for _, line := range lines {
		line = strings.Trim(line, "\t\n ")
		if len(line) > 0 {
			numberofCodeLines += 1

			// find all keywords in line
			foundKeywords := keywordsRegex.FindAllStringIndex(line, -1)
			result.Keywords += len(foundKeywords)

			// find all comments in line
			foundComments := commentRegex.FindAllStringIndex(line, -1)
			result.Comments += len(foundComments)
		}
	}

	result.CodeLines += numberofCodeLines

	return result
}
