package SimpleScraper

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (targetInterpolation *interpolation) inheritFromInterp(interp interpolation) {
	if targetInterpolation.BeginsWith == "" {
		targetInterpolation.BeginsWith = interp.BeginsWith
	}
}

func validateURL(url string) bool {
	validURL := regexp.MustCompile(`^https{0,1}:\/\/www\.`)
	return validURL.Match([]byte(url))
}

func findInterpolation(interp interpolation, potentialInterp []byte, config tomlConfig, url string, tempShards map[string]Shard) []byte {
	targetInterp, ok := config.Interpolations[interp.BeginsWith+stripTokenDelimiters(potentialInterp)]
	if !ok {
		targetInterp, ok = config.Interpolations[stripTokenDelimiters(potentialInterp)]
	}

	if !ok {
		return potentialInterp
	}

	targetInterp.inheritFromInterp(interp)
	replacement := constructInterpolation(url, targetInterp, config, tempShards)
	return validateTokenReplacement(potentialInterp, replacement)
}

// ConstructFromURL - Construct output from the given URL
func ConstructFromURL(url string, tempShards map[string]Shard, interpContains string) string {
	if !validateURL(url) {
		log.Fatalf(`Provided URL is not well formed (i.e. beginning with "http[s]://www.")`)
	}

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

func constructInterpolation(url string, interp interpolation, config tomlConfig, tempShards map[string]Shard) string {
	siteBody, err := goquery.NewDocument(url)
	if err != nil {
		log.Error(err)
		return ""
	}

	interpolationBody := interp.Interpolation
	shardMaps := []map[string]Shard{config.Shards, tempShards}

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
