package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

type PluginConfig struct {
	ID         string   `json:"id"`
	Language   string   `json:"language"`
	Command    string   `json:"command"`
	Extensions []string `json:"extensions"`
}

var PluginConfigsPath = "./internal/plugins/config.json"

type PluginManager struct {
	Plugins  []PluginConfig    `json:"plugins"`
	Defaults map[string]string `json:"defaults"`
}

func NewPluginManager() (*PluginManager, error) {
	/* TODO: implement functionality
	1.) create an empty PluginManger
	2.) return it or error
	*/
	return &PluginManager{
		[]PluginConfig{},
		make(map[string]string),
	}, nil
}

func (pm *PluginManager) LoadPlugins() error {
	/* TODO: implement functionality
	1.) Read in from plugins.json
	2.) return pluginconfigs data
	*/
	data, err := os.ReadFile(PluginConfigsPath)
	if err != nil {
		return fmt.Errorf("could not read plugins config file: %v", err)
	}
	if err := json.Unmarshal(data, &pm); err != nil {
		return fmt.Errorf("coud not unmarshal plugins config data: %v", err)
	}

	return nil
}

func (pm *PluginManager) SavePlugins() error {
	/* TODO: Implement functionality
	1.) write plugins on memory to plugins.json
	*/
	data, err := json.MarshalIndent(pm, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling plugins data %v", err)

	}
	dir := filepath.Dir(PluginConfigsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fialed to create config directory %v", err)
	}
	return os.WriteFile(PluginConfigsPath, data, 0644)
}

func (pm *PluginManager) AddPlugin(newPlugin PluginConfig, setAsDefault bool) error {
	/* TODO: implement functionality
	1.) pull plugin list from plugins.json
	2.) check if plugin already exist
	3.) if not, add plugin to list
	4.) check if setAsDefault is true
	5.) if true, swap default with newPlugin's ID
	6.) save plugins list to plugins.json
	*/
	if newPlugin.Command == "" {
		return fmt.Errorf("no command provided")
	}

	if len(newPlugin.Extensions) < 1 {
		return fmt.Errorf("extensions list cannot be empty")
	}

	if newPlugin.Language == "" {
		return fmt.Errorf("language must be provided")
	}

	if newPlugin.ID == "" {
		return fmt.Errorf("plugin id must be provided")
	}

	for _, existingPlugin := range pm.Plugins {
		if existingPlugin.ID == newPlugin.ID {
			fmt.Println("plugin:", newPlugin.ID, "already exists")
			return fmt.Errorf("plugin: %s, already exists", newPlugin.ID)
		}
	}

	inDefault := false
	// check if it already exists in defaults
	for _, pluginId := range pm.Defaults {
		if pluginId == newPlugin.ID {
			inDefault = true
			return fmt.Errorf("plugin: %s, is already in defaults", newPlugin.ID)
		}
	}

	// if it doesnt exist in plugins, added it
	pm.Plugins = append(pm.Plugins, newPlugin)

	if setAsDefault && !inDefault {
		pm.Defaults[newPlugin.Language] = newPlugin.ID
	}
	return nil

}

func (pm *PluginManager) RemovePlugin(plugin PluginConfig) error {
	/* TODO: implement functionality
	1.) find plugin to delete
	2.) if is default find first compatible plugin and replace the old one
	3.)
	*/
	if plugin.ID == "" {
		return fmt.Errorf("a plugin ID must be provided")
	}

	langForDefaultPlugin := ""
	// remove plugin from defaults list
	for lang, defaultPlugin := range pm.Defaults {
		if defaultPlugin == plugin.ID {
			langForDefaultPlugin = lang
			delete(pm.Defaults, lang)
		}
	}

	isSetNewDefaultPlugin := false
	for i, savedPlugin := range pm.Plugins {
		if savedPlugin.ID == plugin.ID {
			pm.Plugins = slices.Delete(pm.Plugins, i, i+1)
			continue
		}
		if savedPlugin.Language == langForDefaultPlugin {
			pm.Defaults[langForDefaultPlugin] = savedPlugin.ID
			isSetNewDefaultPlugin = true
			break
		}
	}

	if !isSetNewDefaultPlugin {
		fmt.Println("Warning: no other plugin for", langForDefaultPlugin, "was found. Files for this language will not be tracked.")
	}

	return nil
}

func (pm PluginManager) ValidateParserCommand(parser, testFilePath string) error {
	/* TODO: implement functionality
	1.) run the command against the testFile
	2.) if no error detected -> Sweet!
	*/
	cmd := exec.Command(parser, testFilePath)

	// try running command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("parser validation failed: %v (output: %s)", err, string(output))
	}
	return nil
}

func (pm *PluginManager) SetAsDefault(id string) error {
	if id == "" {
		return fmt.Errorf("plugin ID must be provided")
	}

	exsists := false
	var defaultLang string
	var exts []string
	for _, plugin := range pm.Plugins {
		if plugin.ID == id {
			exsists = true
			defaultLang = plugin.Language
			exts = plugin.Extensions
			break
		}
	}

	if !exsists {
		return fmt.Errorf("cannot find plugin with id: %s", id)
	}

	if prevDefaultPlugin, ok := pm.Defaults[defaultLang]; ok {
		fmt.Printf("Plugin %s replacing %s for parsing %s:%v files\n", id, prevDefaultPlugin, defaultLang, exts)
	} else {
		fmt.Printf("Plugin %s set for parsing %s:%v files\n", id, defaultLang, exts)
	}
	pm.Defaults[defaultLang] = id

	return nil

}
