/*
GOTAServer a Simple http server for OTA upgrades of ESP8266.

Copyright (c) 2017 Fredrik Steen <fredrik@ppo2.se>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Defaults for configuration/commandline options
const defaultFirmwareDir = "firmwares"
const defaultFirmwareSuffix = "*.bin"
const defaultVersionReg = `([a-zA-Z]+)_([0-9]+\.[0-9]+)\.`
const defaultServerHostPort = "127.0.0.1:8000"

// Command line flags
var flagFirmwareDir = flag.String("d", defaultFirmwareDir, "Directory to serve firmware-files from")
var flagFirmwareSuffix = flag.String("s", defaultFirmwareSuffix, "Suffix of files to use as firmware")
var flagConfigurationFile = flag.String("c", "", "Configuration file")
var flagServerHostPort = flag.String("p", defaultServerHostPort, "<host>:<port> to listen to")

// Configuration holds configuration settings, either defaults or from config file
type Configuration struct {
	FirmwareDir    string
	FirmwareSuffix string
	ServerHostPort string
}

var cfg Configuration

func loadConfig(path string) Configuration {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File Missing. ", err)
	}

	var config Configuration
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Config Parse Error: ", err)
	}

	// flags overload
	if *flagFirmwareDir != defaultFirmwareDir {
		config.FirmwareDir = *flagFirmwareDir
	}

	if *flagFirmwareSuffix != defaultFirmwareSuffix {
		config.FirmwareSuffix = *flagFirmwareSuffix
	}

	if *flagServerHostPort != defaultServerHostPort {
		config.ServerHostPort = *flagServerHostPort
	}

	return config
}

// getLatestVersion will check cfg.flagFirmwareDir for firmwares
// Returns error when no new firmware is found for version.
func getLatestVersion(project, version string) (filename string, err error) {
	re1 := regexp.MustCompile(defaultVersionReg)
	fwglob := filepath.Join(cfg.FirmwareDir, project, cfg.FirmwareSuffix)
	md := map[float64]string{}

	if files, err := filepath.Glob(fwglob); err == nil {
		for f := range files {
			bf := filepath.Base(files[f])
			res := re1.FindStringSubmatch(bf)[2]
			if ver, err := strconv.ParseFloat(res, 64); err == nil {
				md[ver] = files[f]
			}

		}
		// Get latest version
		var keys []float64
		for k := range md {
			keys = append(keys, k)
		}
		// Sort and reverse to get latest version
		sort.Sort(sort.Reverse(sort.Float64Slice(keys)))
		latestVersion := keys[0]
		fver, err := strconv.ParseFloat(version, 64)
		if err != nil {
			return "", errors.New("Unable to Parse version string")
		}
		if latestVersion <= fver {
			return "", errors.New("Already got the latest version")
		}
		return md[keys[0]], nil
	}
	return "", errors.New("No new firmware")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GOTAserver Firware Serving Server")
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["prj"]
	version := vars["version"]
	latestVersion, err := getLatestVersion(project, version)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("%s", err)
		return
	}
	log.Printf("PRJ: %s REQ-VERSION: %s FOUND: %s", project, version, filepath.Base(latestVersion))
	w.Header().Set("Content-Disposition", "attachment; filename=firmware.bin")
	http.ServeFile(w, r, latestVersion)
}

func logM(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		h.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	})
}

func main() {
	flag.Parse()
	if *flagConfigurationFile != "" {
		cfg = loadConfig(*flagConfigurationFile)
	} else {
		cfg.FirmwareDir = *flagFirmwareDir
		cfg.FirmwareSuffix = *flagFirmwareSuffix
		cfg.ServerHostPort = *flagServerHostPort
	}
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/{prj}/{version}/", versionHandler)

	srv := &http.Server{
		Handler:      logM(r),
		Addr:         cfg.ServerHostPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("============[ GOTAserver ]==================")
	fmt.Printf("Firmware Directory  : %s\n", cfg.FirmwareDir)
	fmt.Printf("Firmware Suffix     : %s\n", cfg.FirmwareSuffix)
	fmt.Printf("Server listening on : %s\n", cfg.ServerHostPort)
	fmt.Println("===========================================")
	log.Fatal(srv.ListenAndServe())

}
