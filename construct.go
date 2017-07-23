package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
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

func (targetInterpolation *interpolation) inheritFromInterp(interp interpolation) {
	if targetInterpolation.BeginsWith == "" {
		targetInterpolation.BeginsWith = interp.BeginsWith
	}
}

func (targetShard *shard) grabShardOverrideURL() (*goquery.Document, error) {
	var doc *goquery.Document
	if !validateURL(targetShard.URL) {
		err := errors.New(`Shard URL override "` + targetShard.URL +
			`" is not well formed (i.e. beginning with "http[s]://www.")`)
		return doc, err
	}
	return goquery.NewDocument(targetShard.URL)
}

func (targetShard *shard) populateShard(doc *goquery.Document) string {
	if targetShard.Override != "" {
		return runModifications(targetShard.Override, targetShard.Modifications)
	}

	var err error
	if targetShard.URL != "" {
		if doc, err = targetShard.grabShardOverrideURL(); err != nil {
			log.Error("Shard", targetShard.Name,
				"- could not grab shard override URL!\n", err)
			return ""
		}
	}

	var found string
	if targetShard.Attr != "" {
		var exists bool
		found, exists = doc.Find(targetShard.Selector).Attr(targetShard.Attr)
		if !exists {
			log.Error("Shard", targetShard.Name,
				"- could not find selector attribute", targetShard.Attr)
		}
	} else {
		found = doc.Find(targetShard.Selector).Text()
	}

	if found == "" {
		log.Error("Shard", targetShard.Name,
			"- could not find anything to populate shard with!")
	}

	return runModifications(found, targetShard.Modifications)
}

func constructFromURL(url string, config tomlConfig, tempShards map[string]shard, interpContains string) string {
	var allInterpolations string
	for interpName, interp := range config.Interpolations {
		for _, urlSubstring := range interp.URLContains {
			if strings.Contains(url, urlSubstring) {
				if strings.Contains(interpName, interpContains) {
					allInterpolations += constructInterpolation(url, interp, config, tempShards)
				}
			}
		}
	}
	if allInterpolations == "" {
		log.Error("Unable to find an interpolation marked for URL:", url)
		return ""
	}
	return allInterpolations
}

func stripTokenDelimiters(token []byte) string { return string(token[1 : len(token)-1]) }

func validateTokenReplacement(originalToken []byte, replacement string) []byte {
	if replacement == "" {
		log.Error("Empty replacement on token:", string(originalToken), "\n Not replacing.")
		return originalToken
	}
	return []byte(replacement)
}

func findShard(potentialShard []byte, interp interpolation, shardMaps []map[string]shard, siteBody *goquery.Document) []byte {
	for _, shardMap := range shardMaps {
		shardName := stripTokenDelimiters(potentialShard)
		resultantShard, ok := shardMap[interp.BeginsWith+shardName]
		if !ok {
			resultantShard, ok = shardMap[shardName]
		}
		if ok {
			replacement := resultantShard.populateShard(siteBody)
			return validateTokenReplacement(potentialShard, replacement)
		}
	}
	return potentialShard
}

func findInterpolation(interp interpolation, potentialInterp []byte, config tomlConfig, url string, tempShards map[string]shard) []byte {
	targetInterp, ok := config.Interpolations[stripTokenDelimiters(potentialInterp)]
	if !ok {
		return potentialInterp
	}
	targetInterp.inheritFromInterp(interp)
	replacement := constructInterpolation(url, targetInterp, config, tempShards)
	return validateTokenReplacement(potentialInterp, replacement)
}

func constructInterpolation(url string, interp interpolation, config tomlConfig, tempShards map[string]shard) string {
	siteBody, err := goquery.NewDocument(url)
	if err != nil {
		log.Error(err)
		return ""
	}

	interpolationBody := interp.Interpolation
	shardMaps := []map[string]shard{config.Shards, tempShards}

	replaceShards := func(potentialShard []byte) []byte {
		return findShard(potentialShard, interp, shardMaps, siteBody)
	}

	replaceInterp := func(potentialInterp []byte) []byte {
		return findInterpolation(interp, potentialInterp, config, url, tempShards)
	}

	interpolationBody = string(
		config.ShardRegex.ReplaceAllFunc([]byte(interpolationBody), replaceShards))
	interpolationBody = string(
		config.InterpRegex.ReplaceAllFunc([]byte(interpolationBody), replaceInterp))

	return runModifications(interpolationBody, interp.Modifications)
}
