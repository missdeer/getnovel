package config

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/missdeer/getnovel/ebook/bs"
	"github.com/missdeer/golib/fsutil"
)

// ReadLocalBookSource reads book sources from the default booksource directory
func ReadLocalBookSource() {
	// Load from command line specified sources first
	LoadBookSourcesFromConfig()

	// Then load from default booksource directory if exists
	if b, _ := fsutil.FileExists("booksource"); b {
		if err := bs.LoadBookSourcesFromDirectory("booksource"); err != nil {
			log.Printf("Failed to load book sources from booksource directory: %v", err)
		}
	}
}

// LoadBookSourcesFromConfig loads book sources based on command line options
func LoadBookSourcesFromConfig() {
	// Load from URL
	if Opts.BookSourceURL != "" {
		urls := strings.Split(Opts.BookSourceURL, ",")
		bs.LoadBookSourcesFromURLs(urls)
	}

	// Load from directory
	if Opts.BookSourceDir != "" {
		if err := bs.LoadBookSourcesFromDirectory(Opts.BookSourceDir); err != nil {
			log.Printf("Failed to load book sources from directory %s: %v", Opts.BookSourceDir, err)
		}
	}

	// Load from single file
	if Opts.BookSourceFile != "" {
		// Try legado format first
		sources := bs.ReadLegadoSourceFromLocalFileSystem(Opts.BookSourceFile)
		if len(sources) > 0 {
			log.Printf("Loaded %d legado sources from %s", len(sources), Opts.BookSourceFile)
		} else {
			// Fall back to V2/V3 format
			bss2 := bs.ReadBookSourceFromLocalFileSystem(Opts.BookSourceFile)
			if len(bss2) > 0 {
				log.Printf("Loaded %d V2/V3 sources from %s", len(bss2), Opts.BookSourceFile)
			}
		}
	}
}

func ParseConfigurations(content []byte, opts *Options) bool {
	var options map[string]interface{}
	if err := json.Unmarshal(content, &options); err != nil {
		log.Println("unmarshal configurations failed", err)
		return false
	}

	oe := reflect.ValueOf(opts).Elem()
	for i := 0; i < oe.NumField(); i++ {
		fieldName := oe.Type().Field(i).Name
		key := strings.ToLower(fieldName[:1]) + fieldName[1:]
		if f, ok := options[key]; ok {
			of := oe.Field(i)
			switch of.Kind() {
			case reflect.String:
				if v := f.(string); len(v) > 0 {
					of.SetString(v)
				}
			case reflect.Float64:
				if v := f.(float64); v > 0 {
					of.SetFloat(v)
				}
			case reflect.Int, reflect.Int64:
				if v := f.(float64); v > 0 {
					of.SetInt(int64(v))
				}
			}
		}
	}
	return true
}

func ReadRemotePreset(opts *Options) bool {
	u := "https://raw.githubusercontent.com/missdeer/getnovel/master/pdfpresets/" + opts.ConfigFile
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse preset request:", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send request:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("response not 200:", resp.StatusCode, resp.Status)
		return false
	}

	c, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading content failed")
		return false
	}

	return ParseConfigurations(c, opts)
}

func ReadLocalConfigFile(opts *Options) bool {
	configFile := opts.ConfigFile
	if b, e := fsutil.FileExists(configFile); e != nil || !b {
		configFile = filepath.Join("pdfpresets", opts.ConfigFile)
		if b, e = fsutil.FileExists(configFile); e != nil || !b {
			log.Println("cannot find configuration file", opts.ConfigFile, "on local file system")
			return false
		}
	}

	contentFd, err := os.OpenFile(configFile, os.O_RDONLY, 0644)
	if err != nil {
		log.Println("opening config file", configFile, "for reading failed", err)
		return false
	}

	contentC, err := io.ReadAll(contentFd)
	contentFd.Close()
	if err != nil {
		log.Println("reading config file", configFile, "failed", err)
		return false
	}

	return ParseConfigurations(contentC, opts)
}
