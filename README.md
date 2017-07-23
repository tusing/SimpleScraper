# SimpleScraper


[![Go Report Card](https://goreportcard.com/badge/github.com/tusing/SimpleScraper)](https://goreportcard.com/report/github.com/tusing/SimpleScraper)


SimpleScraper is a tool designed to allow easy scraping of the web with a only a minimal config file. Users can grab snippets ("shards") from website(s) and apply various changes to the resultant string. Users can then combine these shards into what is referred as an "interpolation", and interpolations can be nested as desired.

This is my first application written in Golang.

### Building
Run `go build -o simplescraper` from the root directory.

### Running

**Quick Start:** `./simplescraper -u <URL> -c <config_file_path>`

```
  -l, --log-level:  Log level: <info|notice|warning|error|critical> (default:
                   error)
  -u, --url:        URL to scrape, provides shard REQUESTED_URL
  -c, --config:     Location of config file (default: config.toml)
  -i, --contains:   If multiple interpolations are compatible with your URL,
                   filter by the interpolations containig the given string.
```

## Configuring SimpleScraper

SimpleScraper features two central ideas: *shards* and *interpolations*.

* **Shards** are individual snippets. Shards can pull information from the internet and apply changes to their own content (prepend/append/regex/replace). Shard fields are as follows:
    - `DelimitShard` defines which (single) characters can be used to mark the beginning and end of a shard.
        + Defining `DelimitShard = ["(", ")"]` and using an interpolation defined as `foo(SOME_SHARD)baz` will interpolate the result of `SOME_SHARD` into the string (if `SOME_SHARD` exists).
    - `[shards.<shard_name>]` defines the shard in the config file.
        + `[shards.SomeShard]` would create a shard with a name `SomeShard`.
    - `Modifications` performs modifications on the shard's result.
        + `Modifications[0]` prepends an element to the shard result.
        + `Modifications[1]` appends an element to the shard result.
        + `Modifications[2]` performs a regex selection on the shard result.
        + `Modifications[3]` replaces what was selected in `Modifications[2]`.
        + All modifications repeat modulo 4. This array can be as long as desired. Modifications are applied in order.
            * `Modifications = ["bar", "baz", "g", "f"]` on some shard `goo` would give us `barfoobaz`.
    - `URL` defines a target URL for the shard; it will override any command-line URL.
        + Running `./simplescraper -u some_url` but defining a shard with `URL: "other_url"` would mean that the shard would source on the `other_url` insted.
    - `Selector` grabs the defined classpath from the given URL.
        + This uses Goquery's `Find(...)` function.
    - `Attr` grabs the HTML attribute associated with the selected object.
        +  `Attr: "href"` will grab the hyperlink for the selected object
    - `Override` bypasses the process of grabbing a value from the web with a hardcoded string. Modifications are still applied.
        + `Override: "some_value"` will ignore the selector and URL on a shard and simply return `some_value` with modifications applied (if any).

* **Interpolations** can be thought of as scaffolding from shards. Once we have nicely formatted values from shards, we can define an interpolation which can present these shards in various manners. Interpolations are quite useful for structuring shards in various markup languages.
    - `DelimitInterp` defines which (single) characters can be used to mark the beginning and end of a interpolation.
        + This is useful to nest interpolations: defining `SOME_INTERPOLATION` as `"foo{ANOTHER_INTERPOLATION}(SOME_SHARD)baz"` will look for and insert the values of both `ANOTHER_INTERPOLATION` and `SOME_SHARD` into `SOME_INTERPOLATION`.
    - `[interpolations.<interpolation_name>]` defines and names an interpolation, similarly to shards.
    - `Modifications` functions as it does for shards.
    - `URLContains` defines substrings of URLs for which the interpolation will be used.
        + Defining `some_interpolation` with `URLContains: ["news"]` and passing `-u https://www.news.google.com` means that `some_interpolation` will be used to scrape `news.google.com`.
    - `BeginsWith` will initially restrict the scope of shards considered to shards that begin with the given string.
        + If we have 2 shards, `foo_shard` and `bar_shard`, setting `BeginsWith="foo_"` will mean that we will look for `foo_shard` first. If `foo_shard` doesn't exist, we try again without restricting our search via the `BeginsWith` string.

---

## Example Config File

The goal of this config file is to grab metadata for stories on `fanfiction.net` and present it in Markdown (for the purposes of a bot designed to run on Reddit.)


#### Config File


```
delimitShard=["(", ")"]
delimitInterp=["{", "}"]


[shards.ffn_title]
selector="#profile_top > b"

[shards.ffn_author]
selector="#profile_top > a:nth-child(5)"

[shards.ffn_author_link]
selector="#profile_top > a:nth-child(5)"
attr="href"
modifications=["https://www.fanfiction.net", "", 'm\.', ""]

[shards.ffn_download]
selector="document.URL"

[shards.ffn_summary]
selector="#profile_top > div"

[shards.ffn_extra]
selector="#profile_top > span.xgray.xcontrast_txt"
modifications=["", "", "[(]", "[",
               "", "", "[)]", "]"]

[interpolations.ffn_markdown]
urlContains=["fanfiction.net"]
beginsWith="ffn_"
interpolation="{generic_markdown}"


[interpolations.generic_markdown]
interpolation="""
[**(title)**]((REQUESTED_URL)) by [(author)]((author_link))

> (summary)

^((extra))
"""
```


#### Explanation

* Shards are getting little snippets of text and attributes from throughout the website.

* Here, interpolations let us easily create different scaffolds to present our shards in. It would be quite easy to define additional interpolations for BBCode, HTML, or even plaintext.

* Nested interpolations are useful when expanding to multiple websites. For example, if I wanted to add an interpolation for a different story site (let's call it `ao3`), then I would simply have to do
```
[shards.ao3_<shard_name>] # do this for all shards used by the generic_markdown interpolation

[interpolations.ao3_markdown]
urlContains=["archiveofourown.org"]
beginsWith="ao3_"
interpolation="{generic_markdown}"
```

* In the above example, notice how the `BeginsWith` attribute saved us the pain of duplicating much. We can define both shards `ffn_author` and `ao3_author`, but the interpolation will intelligently pick out which one to use even if we simply pass it `author`.


#### Behavior

```
$ go run *.go -u https://www.fanfiction.net/s/3964606/1/Alexandra-Quick-and-the-Thorn-Circle

[**Alexandra Quick and the Thorn Circle**](https://www.fanfiction.net/s/3964606/1/Alexandra-Quick-and-the-Thorn-Circle) by [Inverarity](https://www.fanfiction.net/u/1374917/Inverarity)

> The war against Voldemort never reached America, but all is not well there. When 11-year-old Alexandra Quick learns she is a witch, she is plunged into a world of prejudices, intrigue, and danger. Who wants Alexandra dead, and why?

^(Rated: Fiction  K+ - English - Fantasy/Adventure -  OC - Chapters: 29   - Words: 165,657 - Reviews: 570 - Favs: 794 - Follows: 296 - Updated: 12/24/2007 - Published: 12/23/2007 - Status: Complete - id: 3964606 )
```
