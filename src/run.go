package js

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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

// Start is a function that analyzes the source code directory and generates a software bill of materials (SBOM) output.
// It returns an sbomTypes.Output struct containing the analysis results.
func Start(sourceCodeDir string, analysisId uuid.UUID, codeclarityDB *bun.DB) types.Output {
	analysis := codeclarity.Analysis{
		Id: analysisId,
	}
	err := codeclarityDB.NewSelect().Model(&analysis).WherePK().Scan(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch analysis by id: %s", err.Error()))
	}

	python_config, ok := analysis.Config["python"].(map[string]interface{})
	if !ok {
		panic("Failed to fetch analysis config")
	}

	projectId := python_config["project"].(string)

	var chat types.Chat
	err = codeclarityDB.NewSelect().Model(&chat).Where("? = ?", bun.Ident("projectId"), projectId).Scan(context.Background())
	if err == nil {
		chat.Messages[0].Result = analysisId.String()
		_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Failed to add image to chat history: %s", err.Error()))
		}
	}
	out := ExecuteScript(sourceCodeDir, analysisId)
	chat.Messages[0].Image = out.Result.Image
	chat.Messages[0].Text = out.Result.Text
	chat.Messages[0].Data = out.Result.Data
	_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to add result content to chat history: %s", err.Error()))
	}
	return out
}

func ExecuteScript(sourceCodeDir string, analysisId uuid.UUID) types.Output {
	start := time.Now()

	scriptPath := path.Join(sourceCodeDir, "python", "script.py")
	files, err := filepath.Glob(scriptPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		return generate_output(start, "", nil, "", codeclarity.FAILURE, []exceptionManager.Error{})
	}
	scanpyOutputPath := path.Join(sourceCodeDir, "scanpy")
	outputPath := path.Join(sourceCodeDir, "python")
	dataPath := path.Join(sourceCodeDir, "data")
	os.MkdirAll(outputPath, os.ModePerm)
	os.MkdirAll(dataPath, os.ModePerm)

	args := []string{scriptPath, scanpyOutputPath, outputPath}

	// Run Rscript in sourceCodeDir
	cmd := exec.Command("python3", args...)
	message, err := cmd.CombinedOutput()
	if err != nil {
		// panic(fmt.Sprintf("Failed to run Rscript: %s", err.Error()))
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
		return generate_output(start, "", nil, "", codeclarity.FAILURE, []exceptionManager.Error{codeclarity_error})
	}

	// Move generated image inside outputPath to dataPath
	// Rename it to $analysisId$.png
	files, err = filepath.Glob(outputPath + "/*")
	if err != nil {
		log.Fatal(err)
	}
	image := ""
	for _, f := range files {
		if strings.HasSuffix(f, ".png") {
			newName := filepath.Join(dataPath, analysisId.String()+".png")
			os.Rename(f, newName)
			image = analysisId.String()
		}
	}

	// Move generated image inside outputPath to dataPath
	// Rename it to $analysisId$.txt
	files, err = filepath.Glob(outputPath + "/*")
	if err != nil {
		log.Fatal(err)
	}
	text := ""
	for _, f := range files {
		if strings.HasSuffix(f, ".txt") {
			newName := filepath.Join(dataPath, analysisId.String()+".txt")
			os.Rename(f, newName)
			// Open the renamed text file and put its content in the 'text' variable
			txtFile, err := os.Open(newName)
			if err != nil {
				panic(fmt.Sprintf("Failed to open text file: %s", err.Error()))
			}
			defer txtFile.Close()

			var buffer bytes.Buffer
			scanner := bufio.NewScanner(txtFile)
			for scanner.Scan() {
				buffer.WriteString(scanner.Text() + "\n")
			}
			text = buffer.String()
		}
	}

	// Move generated image inside outputPath to dataPath
	// Rename it to $analysisId$.txt
	files, err = filepath.Glob(outputPath + "/*")
	if err != nil {
		log.Fatal(err)
	}
	var data map[string]interface{}
	for _, f := range files {
		if strings.HasSuffix(f, ".json") {
			newName := filepath.Join(dataPath, analysisId.String()+".json")
			os.Rename(f, newName)
			jsonFile, err := os.Open(newName)
			if err != nil {
				panic(fmt.Sprintf("Failed to open JSON file: %s", err.Error()))
			}
			defer jsonFile.Close()

			decoder := json.NewDecoder(jsonFile)
			err = decoder.Decode(&data)
			if err != nil {
				panic(fmt.Sprintf("Failed to decode JSON data: %s", err.Error()))
			}
		}
	}

	return generate_output(start, image, data, text, codeclarity.SUCCESS, []exceptionManager.Error{})
}

func generate_output(start time.Time, imageName string, data any, text string, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(start)

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
	return output
}
