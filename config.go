package main

import (
	"github.com/BurntSushi/toml"
	"regexp"
)

func mustConstructConfig(configFilepath string) tomlConfig {
	var config tomlConfig
	if _, err := toml.DecodeFile(configFilepath, &config); err != nil {
		log.Fatal("Could not parse config file!\n", err)
	}
	return config
}

type tomlConfig struct {
	Shards         map[string]shard
	Interpolations map[string]interpolation
	DelimitShard   []string
	DelimitInterp  []string
	ShardRegex     *regexp.Regexp
	InterpRegex    *regexp.Regexp
}

type shard struct {
	Name          string
	Modifications []string
	Override      string
	URL           string
	Selector      string
	Attr          string
}

type interpolation struct {
	Modifications []string
	URLContains   []string
	Interpolation string
	BeginsWith    string
}

func (config *tomlConfig) initialize() {
	for shardName, targetShard := range config.Shards {
		tmpShard := targetShard
		tmpShard.Name = shardName
		config.Shards[shardName] = tmpShard
	}
	config.ShardRegex = constructTokenRegex(config.DelimitShard[0], config.DelimitShard[1])
	config.InterpRegex = constructTokenRegex(config.DelimitInterp[0], config.DelimitInterp[1])
}

func constructTokenRegex(tokenBegin string, tokenEnd string) *regexp.Regexp {
	re, err := regexp.Compile("[" + tokenBegin + "][^" + tokenBegin + tokenEnd + "]+[" + tokenEnd + "]")
	if err != nil {
		log.Error("Could not compile regex for tokens: " + tokenBegin + " and " + tokenEnd)
		log.Fatal(err)
	}
	return re
}

func runModifications(original string, mods []string) string {
	modified := []byte(original)
	padding := make([]string, 4-len(mods)%4)
	mods = append(mods, padding...)
	for i := 0; i < len(mods); i += 4 {
		modified = append([]byte(mods[i]), modified...)   // 1. Prepend
		modified = append(modified, []byte(mods[i+1])...) // 2. Append
		regex, err := regexp.Compile(mods[i+2])           // 3. Select & 4. Replace
		if err != nil {
			log.Error("Failed to compile regex:", err)
		}
		replacement := mods[i+3]
		modified = regex.ReplaceAll(modified, []byte(replacement))
	}

	return string(modified)
}

func stripTokenDelimiters(token []byte) string { return string(token[1 : len(token)-1]) }

func validateTokenReplacement(originalToken []byte, replacement string) []byte {
	if replacement == "" {
		log.Error("Empty replacement on token:", string(originalToken), "\n Not replacing.")
		return originalToken
	}
	return []byte(replacement)
}
