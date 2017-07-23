package main

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
)

func (targetShard *shard) grabShardOverrideURL() (*goquery.Document, error) {
	var doc *goquery.Document
	if !validateURL(targetShard.URL) {
		err := errors.New(`Shard URL override "` + targetShard.URL +
			`" is not well formed (i.e. beginning with "http[s]://www.")`)
		return doc, err
	}
	return goquery.NewDocument(targetShard.URL)
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
