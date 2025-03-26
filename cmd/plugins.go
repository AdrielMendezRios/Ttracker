/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	plugin "Ttracker/internal/plugins"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// addParserCmd represents the addParser command
var pluginsCmd = &cobra.Command{
	Use:   "plugins [-a] [-r] [-d]",
	Short: "display, add, remove or set a plugin as the default for a given language",
	Long: `Manage Ttracker parser plugins for different programming languages.

The plugins command allows you to list, add, remove, and configure 
parser plugins that Ttracker uses to find TODOs in various languages.

OPERATIONS:
  --list, -l       List all available plugins (default if no operation specified)
  --add, -a        Add a new parser plugin
  --remove, -r     Remove an existing parser plugin
  --default, -d    Set a plugin as the default for its language

PLUGIN DETAILS:
  --id, -i         Plugin identifier (required for add/remove/default)
  --lang           Programming language (required for add)
  --cmd            Command path to the parser (required for add)
  --ext            File extensions as comma-separated list (required for add)

ADDITIONAL OPTIONS:
  --force, -f      Skip confirmation prompts
  --verbose, -v    Show detailed information when listing
  --format         Output format (text, json) // TODO: look into implementing this
  --no-validate    Skip command validation when adding

EXAMPLES:
  # List all plugins
  tt plugins

  # List plugins with detailed information
  tt plugins --list --verbose

  # Add a new JavaScript parser
  tt plugins --add --id "js-standard" --lang "javascript" \
    --cmd "./parsers/js_parser.js" --ext ".js,.jsx"

  # Remove a plugin
  tt plugins --remove --id "js-standard"

  # Set a plugin as default for its language
  tt plugins --default --id "js-standard"

Each plugin must have a unique ID. To overwrite an existing plugin,
use the --force flag with --add.`,
	Run: pluginsRun,
}

func init() {
	rootCmd.AddCommand(pluginsCmd)

	// Operation flags (mutually exclusive)
	pluginsCmd.Flags().BoolP("list", "l", false, "List all available plugins")
	pluginsCmd.Flags().BoolP("add", "a", false, "Add a new parser plugin")
	pluginsCmd.Flags().BoolP("remove", "r", false, "Remove an existing parser plugin")
	pluginsCmd.Flags().BoolP("default", "d", false, "Set a plugin as the default for its language")

	// plugin details flags
	pluginsCmd.Flags().StringP("id", "i", "", "Plugin Identifier")
	pluginsCmd.Flags().String("lang", "", "The language the parser is meant to parse")
	pluginsCmd.Flags().String("cmd", "", "Command path to the parser")
	pluginsCmd.Flags().String("ext", "", "File extensions as a comma-separated list")
	pluginsCmd.Flags().String("testFile", "", "Path to test file to validate new plugin")

	// additional options
	pluginsCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompts")
	pluginsCmd.Flags().BoolP("verbose", "v", false, "Show detailed information when listing")
	// pluginsCmd.Flags().String("format", "text", "Output format (text, json)")
	pluginsCmd.Flags().Bool("no-validate", false, "Skip command validation when adding")
}

func pluginsRun(cmd *cobra.Command, args []string) {
	isAdd := cmd.Flags().Lookup("add").Changed
	isRemove := cmd.Flags().Lookup("remove").Changed
	isDefault := cmd.Flags().Lookup("default").Changed
	isList := cmd.Flags().Lookup("list").Changed
	ID, _ := cmd.Flags().GetString("id")

	// Default to list if no operation specified
	if !isAdd && !isRemove && !isDefault && !isList {
		isList = true
	}

	// Error is multiple operation specified
	if (isAdd && isRemove) || (isRemove && isDefault) {
		fmt.Println("Error: Please specify only one operation")
		return
	}
	pluginMngr, err := plugin.NewPluginManager()

	if err != nil {
		fmt.Println("Error instantiating Plugin Manager:", err)
		return
	}
	pluginMngr.LoadPlugins()

	if isDefault && ID != "" && !isAdd {
		err := pluginMngr.SetAsDefault(ID)
		if err != nil {
			fmt.Printf("Error setting plugin with id %s as a default plugin for a language", ID)
		}
	}

	if isList {
		for _, plugin := range pluginMngr.Plugins {
			fmt.Printf("%s:\n\tCommand: %s\n\tLanguage: %s\n\tExtensions: %s\n", plugin.ID, plugin.Command, plugin.Language, plugin.Extensions)
		}
		return
	}

	if isAdd {
		addPlugin(cmd, pluginMngr, isDefault)
		return
	}

	if isRemove {
		removePlugin(cmd, pluginMngr)
		return
	}

}

func addPlugin(cmd *cobra.Command, pluginMngr *plugin.PluginManager, isDefault bool) {
	id, _ := cmd.Flags().GetString("id")
	lang, _ := cmd.Flags().GetString("lang")
	command, _ := cmd.Flags().GetString("cmd")
	ext, _ := cmd.Flags().GetString("ext")
	testFile, _ := cmd.Flags().GetString("testFile")
	isValidate, _ := cmd.Flags().GetBool("no-validate")

	if id == "" || lang == "" || command == "" || ext == "" {
		fmt.Println("Error: --id, --lang, --cmd, and --ext are required for add operation")
		return
	}
	extRaw := strings.Split(ext, ",")
	exts := make([]string, len(extRaw))
	for i, e := range extRaw {
		exts[i] = strings.TrimSpace(e)
	}

	newPlugin := plugin.PluginConfig{
		ID:         id,
		Language:   lang,
		Command:    command,
		Extensions: exts,
	}

	if err := pluginMngr.AddPlugin(newPlugin, isDefault); err != nil {
		fmt.Println("Error adding:", id, "plugin. Error:", err)
		return
	}

	if isValidate {
		if err := pluginMngr.ValidateParserCommand(command, testFile); err != nil {
			fmt.Println(err)
			return
		}
	}

	if err := pluginMngr.SavePlugins(); err != nil {
		fmt.Println("Error saving plugin:", id, "Error:", err)
		return
	}
}

func removePlugin(cmd *cobra.Command, pluginMngr *plugin.PluginManager) {
	id, _ := cmd.Flags().GetString("id")
	var deadPlugin plugin.PluginConfig
	for _, plugin := range pluginMngr.Plugins {
		if plugin.ID == id {
			deadPlugin = plugin
			break
		}
	}

	if err := pluginMngr.RemovePlugin(deadPlugin); err != nil {
		fmt.Println("Error removing plugin: ", id, "Error:", err)
		return
	}

	if err := pluginMngr.SavePlugins(); err != nil {
		fmt.Println("error saving plugin data, error:", err)
		return
	}
}
