package python

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/parithera/plugin-python/src/types"
	"github.com/parithera/plugin-python/src/utils/output_generator"
)

// Start initiates the Python analysis process.
// It fetches analysis details, executes the Python script, and updates the chat history with the results.
func Start(sourceCodeDir string, analysisId uuid.UUID, codeclarityDB *bun.DB) types.Output {
	// Fetch analysis details from the database.
	analysis := codeclarity.Analysis{
		Id: analysisId,
	}
	err := codeclarityDB.NewSelect().Model(&analysis).WherePK().Scan(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch analysis by id: %s", err.Error()))
	}

	// Retrieve Python configuration from the analysis details.
	python_config, ok := analysis.Config["python"].(map[string]interface{})
	if !ok {
		panic("Failed to fetch analysis config")
	}

	// Extract the project ID from the Python configuration.
	projectId := python_config["project"].(string)

	// Fetch chat history associated with the project ID.
	var chat types.Chat
	err = codeclarityDB.NewSelect().Model(&chat).Where("? = ?", bun.Ident("projectId"), projectId).Scan(context.Background())
	if err == nil {
		// Update the chat message with the analysis ID.
		chat.Messages[0].Image = analysisId.String()
		// Execute the update operation in the database.
		_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Failed to add image to chat history: %s", err.Error()))
		}
	}

	// Execute the Python script and obtain the output.
	out := ExecuteScript(sourceCodeDir, analysisId)

	// Update the chat message with the script's results.
	chat.Messages[0].Text = out.Result.Text
	chat.Messages[0].Image = out.Result.Image
	chat.Messages[0].JSON = out.Result.Data

	// Execute the update operation in the database.
	_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to add result content to chat history: %s", err.Error()))
	}

	// Return the output from the script execution.
	return out
}

// ExecuteScript executes the Python script and processes the results.
func ExecuteScript(sourceCodeDir string, analysisId uuid.UUID) types.Output {
	// Record the start time for performance analysis.
	start := time.Now()

	// Construct the path to the Python script.
	scriptPath := path.Join(sourceCodeDir, "python", "script.py")

	// Find all files matching the script path.
	files, err := filepath.Glob(scriptPath)
	if err != nil {
		log.Fatal(err)
	}

	// If the script is not found, return a failure output.
	if len(files) == 0 {
		return generate_output(start, "", nil, "", codeclarity.FAILURE, []exceptionManager.Error{})
	}

	// Construct the output and data paths.
	outputPath := path.Join(sourceCodeDir, "python")
	dataPath := path.Join(sourceCodeDir, "data")

	// Create the output and data directories if they don't exist.
	os.MkdirAll(outputPath, os.ModePerm)
	os.MkdirAll(dataPath, os.ModePerm)

	// Define the arguments for the Python script execution.
	args := []string{scriptPath, outputPath}

	// Execute the Python script using the 'python3' command.
	cmd := exec.Command("python3", args...)
	message, err := cmd.CombinedOutput()
	if err != nil {
		// Create an error object with the error message and type.
		codeclarity_error := exceptionManager.Error{
			Private: exceptionManager.ErrorContent{
				Description: string(message),
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "The script failed to execute",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}
		// Return a failure output with the error object.
		return generate_output(start, "", nil, "", codeclarity.FAILURE, []exceptionManager.Error{codeclarity_error})
	}

	// Find all files in the output path.
	files, err = filepath.Glob(outputPath + "/*")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize variables to store the image name, text content, and data.
	image := ""
	text := ""
	var data map[string]interface{}

	// Iterate over the files in the output path.
	for _, f := range files {
		// Check if the file is a PNG image.
		if strings.HasSuffix(f, ".png") {
			// Rename the image file to include the analysis ID.
			newName := filepath.Join(dataPath, analysisId.String()+".png")
			os.Rename(f, newName)
			image = analysisId.String()
		}

		// Check if the file is a TXT file.
		if strings.HasSuffix(f, ".txt") {
			// Rename the text file to include the analysis ID.
			newName := filepath.Join(dataPath, analysisId.String()+".txt")
			os.Rename(f, newName)

			// Open the renamed text file.
			txtFile, err := os.Open(newName)
			if err != nil {
				panic(fmt.Sprintf("Failed to open text file: %s", err.Error()))
			}
			defer txtFile.Close()

			// Read the content of the text file into a buffer.
			var buffer bytes.Buffer
			scanner := bufio.NewScanner(txtFile)
			for scanner.Scan() {
				buffer.WriteString(scanner.Text() + "\n")
			}
			text = buffer.String()
		}
	}

	// Find all files in the output path.
	files, err = filepath.Glob(outputPath + "/*")
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over the files in the output path.
	for _, f := range files {
		// Skip the groups.json file.
		if strings.Contains(f, "groups.json") {
			continue
		}

		// Check if the file is a JSON file.
		if strings.HasSuffix(f, ".json") {
			// Rename the JSON file to include the analysis ID.
			newName := filepath.Join(dataPath, analysisId.String()+".json")
			os.Rename(f, newName)
		}
	}

	// Generate the output with the image name, data, text content, status, and errors.
	return generate_output(start, image, data, text, codeclarity.SUCCESS, []exceptionManager.Error{})
}

// generate_output generates the output object with the analysis results.
func generate_output(start time.Time, imageName string, data any, text string, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	// Calculate the analysis timing.
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(start)

	// Create the output object.
	output := types.Output{
		Result: types.Result{
			Image: imageName,
			Data:  data,
			Text:  text,
		},
		AnalysisInfo: types.AnalysisInfo{
			Errors: errors,
			Time: types.Time{
				AnalysisStartTime: formattedStart,
				AnalysisEndTime:   formattedEnd,
				AnalysisDeltaTime: delta,
			},
			Status: status,
		},
	}

	// Return the output object.
	return output
}
