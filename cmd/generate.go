package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pablor21/gonnotation/parser"
	"github.com/pablor21/protoschemagen/plugin"
	"gopkg.in/yaml.v3"
)

func Generate() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	var cfg *parser.Config

	if *configFile != "" {
		// Read configuration from file
		fmt.Printf("Loading configuration from: %s\n", *configFile)

		// Read the config file
		data, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to read config file %s: %v", *configFile, err)
		}

		// Parse the config
		cfg = &parser.Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.Fatalf("Failed to parse config file %s: %v", *configFile, err)
		}

		// Set the config directory for relative path resolution
		cfg.ConfigDir = filepath.Dir(*configFile)

		// Enable debug logging
		cfg.LogLevel = parser.Ptr(parser.LogLevelDebug)
	} else {
		// Use default configuration
		cfg = parser.NewConfigWithDefaults()

		// Enable debug logging
		cfg.LogLevel = parser.Ptr(parser.LogLevelDebug)

		// Override defaults for protobuf generation
		cfg.Generate = []string{"protobuf"}
		cfg.Packages = []string{"./*.go"}

		// Set auto-generation to include all annotated structs
		cfg.AutoGenerate = &parser.AutoGenerateConfig{
			Enabled:  true,
			Strategy: parser.AutoGenAll,
		}

		// Add explicit config for protobuf plugin
		cfg.Plugins["protobuf"] = map[string]interface{}{
			"enabled": true,
			"syntax":  "proto3",
			"package": "example.v1",
			"output":  "schema.proto",
			"options": map[string]string{
				"go_package": "github.com/example/proto/v1",
			},
		}
	}

	fmt.Printf("Config: %+v\n", cfg)
	fmt.Printf("Packages: %v\n", cfg.Packages)

	// Create multi-format generator
	gen := parser.NewMultiFormatGenerator(cfg)

	// Create protobuf plugin and register it
	protoPlugin := plugin.NewPlugin(nil) // Use default config which will be overridden by cfg.Plugins

	gen.RegisterPlugin(protoPlugin)

	// Generate schema
	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
		os.Exit(1)
	}

	// After schema generation is complete, run protoc if stub generation is enabled
	if protoPlugin.GetConfig().GenerateStubs != nil && protoPlugin.GetConfig().GenerateStubs.Enabled {
		fmt.Println("Generating protobuf Go stubs...")
		if err := plugin.GenerateProtobufGoFilesStandaloneWithConfigDir(protoPlugin.GetConfig().GenerateStubs, protoPlugin.GetConfig(), cfg.ConfigDir); err != nil {
			log.Fatalf("Failed to generate protobuf Go files: %v", err)
		}
		fmt.Println("Protobuf Go stubs generation completed successfully!")
	}

	fmt.Println("Schema generation completed successfully!")
}

func GenerateStubs() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	if *configFile == "" {
		log.Fatalf("Config file is required for generate-stubs command")
	}

	// Read configuration from file
	fmt.Printf("Loading configuration from: %s\n", *configFile)

	// Read the config file
	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", *configFile, err)
	}

	// Parse the config
	cfg := &parser.Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("Failed to parse config file %s: %v", *configFile, err)
	}

	// Set the config directory for relative path resolution
	cfg.ConfigDir = filepath.Dir(*configFile)

	// Find protobuf plugin configuration in the map
	var protoPluginConfig *plugin.Config
	if pluginData, exists := cfg.Plugins["protobuf"]; exists {
		if pluginMap, ok := pluginData.(map[string]interface{}); ok {
			// Convert the map back to plugin.Config
			pluginBytes, err := yaml.Marshal(pluginMap)
			if err != nil {
				log.Fatalf("Failed to marshal plugin config: %v", err)
			}
			protoPluginConfig = &plugin.Config{}
			if err := yaml.Unmarshal(pluginBytes, protoPluginConfig); err != nil {
				log.Fatalf("Failed to unmarshal plugin config: %v", err)
			}
		}
	}

	if protoPluginConfig == nil {
		log.Fatalf("No protobuf plugin configuration found in config file")
	}

	if protoPluginConfig.GenerateStubs == nil || !protoPluginConfig.GenerateStubs.Enabled {
		fmt.Println("Stub generation is not enabled, nothing to do")
		return
	}

	// Run protoc on existing proto files
	fmt.Println("Generating protobuf Go stubs...")
	if err := plugin.GenerateProtobufGoFilesStandalone(protoPluginConfig.GenerateStubs, protoPluginConfig); err != nil {
		log.Fatalf("Failed to generate protobuf Go files: %v", err)
	}

	fmt.Println("Protobuf Go stubs generation completed successfully!")
}
