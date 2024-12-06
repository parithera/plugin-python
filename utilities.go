package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp_helper "github.com/CodeClarityCE/utility-amqp-helper"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin_db "github.com/CodeClarityCE/utility-types/plugin_db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// callback is a function that processes a message received from a plugin dispatcher.
// It takes the following parameters:
// - args: any, the arguments passed to the callback function.
// - config: types_plugin.Plugin, the configuration of the plugin.
// - message: []byte, the message received from the plugin dispatcher.
//
// The callback function performs the following steps:
// 1. Extracts the arguments from the args parameter.
// 2. Opens a database connection.
// 3. Reads the message and unmarshals it into a dispatcherMessage struct.
// 4. Starts a timer to measure the execution time.
// 5. Retrieves the analysis document from the database.
// 6. Starts the analysis using the startAnalysis function.
// 7. Prints the elapsed time.
// 8. Updates the analysis with the results and status.
// 9. Commits the transaction.
// 10. Sends the results to the plugins_dispatcher.
//
// If any error occurs during the execution of the callback function, it will be logged and the transaction will be aborted.
func callback(args any, config plugin_db.Plugin, message []byte) {
	// Get arguments
	s, ok := args.(Arguments)
	if !ok {
		log.Printf("not ok")
		return
	}

	// Read message
	var dispatcherMessage types_amqp.DispatcherPluginMessage
	err := json.Unmarshal([]byte(message), &dispatcherMessage)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Start timer
	start := time.Now()

	// Open DB
	host := os.Getenv("PG_DB_HOST")
	if host == "" {
		log.Printf("PG_DB_HOST is not set")
		return
	}
	port := os.Getenv("PG_DB_PORT")
	if port == "" {
		log.Printf("PG_DB_PORT is not set")
		return
	}
	user := os.Getenv("PG_DB_USER")
	if user == "" {
		log.Printf("PG_DB_USER is not set")
		return
	}
	password := os.Getenv("PG_DB_PASSWORD")
	if password == "" {
		log.Printf("PG_DB_PASSWORD is not set")
		return
	}

	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbhelper.Config.Database.Results + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	ctx := context.Background()
	analysis_document := codeclarity.Analysis{
		Id: dispatcherMessage.AnalysisId,
	}
	err = db.NewSelect().Model(&analysis_document).WherePK().Scan(ctx)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Start analysis
	result, status, err := startAnalysis(s, dispatcherMessage, config, analysis_document)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Print time elapsed
	t := time.Now()
	elapsed := t.Sub(start)
	log.Println(elapsed)

	// Send results
	analysis_document, err = updateAnalysis(result, status, analysis_document, config, start, t, db)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	// Send results
	sbom_message := types_amqp.PluginDispatcherMessage{
		AnalysisId: dispatcherMessage.AnalysisId,
		Plugin:     config.Name,
	}
	data, _ := json.Marshal(sbom_message)
	amqp_helper.Send("plugins_dispatcher", data)
}

// readConfig reads the configuration file and returns a Plugin object and an error.
// The configuration file is expected to be named "config.json" and should be located in the same directory as the source file.
// If the file cannot be opened or if there is an error decoding the file, an error is returned.
// The returned Plugin object contains the parsed configuration values, with the Key field set as the concatenation of the Name and Version fields.
// If there is an error registering the plugin, an error is returned.
func readConfig() (plugin_db.Plugin, error) {
	// Read config file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Printf("%v", err)
		return plugin_db.Plugin{}, err
	}
	defer configFile.Close()

	// Decode config file
	var config plugin_db.Plugin
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		log.Printf("%v", err)
		return plugin_db.Plugin{}, err
	}
	// config.Key = config.Name + ":" + config.Version

	err = register(config)
	if err != nil {
		log.Printf("%v", err)
		return plugin_db.Plugin{}, err
	}

	return config, nil
}

// register is a function that registers a plugin configuration in the database.
// It takes a config parameter of type types_plugin.Plugin, which represents the plugin configuration to be registered.
// The function returns an error if there was an issue with the registration process.
func register(config plugin_db.Plugin) error {
	host := os.Getenv("PG_DB_HOST")
	if host == "" {
		log.Printf("PG_DB_HOST is not set")
		return fmt.Errorf("PG_DB_HOST is not set")
	}
	port := os.Getenv("PG_DB_PORT")
	if port == "" {
		log.Printf("PG_DB_PORT is not set")
		return fmt.Errorf("PG_DB_PORT is not set")
	}
	user := os.Getenv("PG_DB_USER")
	if user == "" {
		log.Printf("PG_DB_USER is not set")
		return fmt.Errorf("PG_DB_USER is not set")
	}
	password := os.Getenv("PG_DB_PASSWORD")
	if password == "" {
		log.Printf("PG_DB_PASSWORD is not set")
		return fmt.Errorf("PG_DB_PASSWORD is not set")
	}

	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbhelper.Config.Database.Plugins + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	ctx := context.Background()
	exists, err := db.NewSelect().Model((*plugin_db.Plugin)(nil)).Where("name = ?", config.Name).Exists(ctx)

	if err != nil {
		log.Printf("%v", err)
		return err
	}
	if !exists {
		_, err = db.NewInsert().Model(&config).Exec(ctx)
		if err != nil {
			log.Printf("%v", err)
			return err
		}
	}
	return nil
}

// updateAnalysis updates the analysis document in the database with the provided result and status.
// It searches for the step with the same name as the plugin's name in the specified stage of the analysis document.
// If the step is found, its status is updated to "success" or "failure" based on the provided status.
// The result is stored in the step's result field.
// Finally, the updated analysis document is saved back to the database.
// If the step is not found, an error is returned.
func updateAnalysis(result map[string]any, status codeclarity.AnalysisStatus, analysis_document codeclarity.Analysis, config plugin_db.Plugin, start time.Time, end time.Time, db *bun.DB) (codeclarity.Analysis, error) {
	for step_id, step := range analysis_document.Steps[analysis_document.Stage] {
		// Update step status
		if step.Name == config.Name {
			err := db.RunInTx(context.Background(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
				// Retrieve analysis document
				err := tx.NewSelect().Model(&analysis_document).WherePK().Scan(ctx)
				if err != nil {
					return err
				}

				// Update step information
				analysis_document.Steps[analysis_document.Stage][step_id].Status = status
				analysis_document.Steps[analysis_document.Stage][step_id].Result = result
				analysis_document.Steps[analysis_document.Stage][step_id].Started_on = start.Format(time.RFC3339Nano)
				analysis_document.Steps[analysis_document.Stage][step_id].Ended_on = end.Format(time.RFC3339Nano)

				// Update analysis document
				_, err = tx.NewUpdate().Model(&analysis_document).WherePK().Exec(ctx)
				return err
			})
			if err != nil {
				log.Printf("%v", err)
				return codeclarity.Analysis{}, fmt.Errorf("error updating analysis")
			}
			return analysis_document, nil
		}
	}
	return codeclarity.Analysis{}, fmt.Errorf("step not found")
}
