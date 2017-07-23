package main

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

func (targetInterpolation *interpolation) inheritFromInterp(interp interpolation) {
	if targetInterpolation.BeginsWith == "" {
		targetInterpolation.BeginsWith = interp.BeginsWith
	}
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
