// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"log"
	"maps"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	version       = "v0.0.5"
	GoVersion     = runtime.Version()
	GoOS          = runtime.GOOS
	GoArch        = runtime.GOARCH
	envMap        = make(map[string]any)
	listenAddress = kingpin.Flag(
		"listen-address",
		"Address on which to expose metrics and web interface.",
	).Envar("HI_LISTEN_PORT").Default(":8080").String()

	onlyFields = kingpin.Flag(
		"filed",
		"The regular expression to check the field, case insensitive. example: --filed=\"(URI|Proto|Proto|PWD|User-Agent)\"",
	).Envar("HI_ECHO_FIELD").String()

	enableIndent = kingpin.Flag(
		"indent",
		"enable Json Indent").Default("true").
		Envar("HI_ENABLE_INDENT").Bool()
)

func appVersion(appName string) string {
	return fmt.Sprintf("%s version %s\n  go version:\t%s\n  platform:\t%s/%s\n", appName, version, GoVersion, GoOS, GoArch)
}

func SayHi(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	respCode := http.StatusOK

	defer func() {
		log.Printf("%v | %v | %v | %v %v",
			respCode,
			time.Duration(time.Now().UnixNano()-start.UnixNano())*time.Microsecond,
			r.RemoteAddr, r.Method, r.RequestURI)
	}()

	// 构造HTTP 字段 TODO 用反射实现
	var httpMap = make(map[string]any)
	httpMap["Method"] = r.Method
	httpMap["URI"] = r.RequestURI
	httpMap["Host"] = r.Host
	httpMap["RemoteAddr"] = r.RemoteAddr
	httpMap["Proto"] = r.Proto
	httpMap["ContentLength"] = r.ContentLength

	var fieldsMap = make(map[string]any)

	// 存在冲突: 后应用的生效。
	if len(*onlyFields) > 0 {
		fieldsReg := regexp.MustCompile("(?i)" + *onlyFields)
		// 匹配环境变量
		maps.Copy(fieldsMap, envMap)

		// 匹配Http Header
		for k, v := range r.Header {
			if fieldsReg.MatchString(k) {
				fieldsMap[k] = v
			}
		}

		// 匹配Http Field
		for k, v := range httpMap {
			if fieldsReg.MatchString(k) {
				fieldsMap[k] = v
			}
		}

	} else {
		for k, v := range envMap {
			httpMap[k] = v
		}
		maps.Copy(fieldsMap, httpMap)
		fieldsMap["Header"] = r.Header
	}

	res, err := json.Marshal(fieldsMap)
	if err != nil {
		log.Println("JSON Marshal error: ", err)
		respCode = http.StatusInternalServerError
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if *enableIndent {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, res, "", "  ")

		if err != nil {
			log.Println("JSON Indent error: ", err)
			respCode = http.StatusInternalServerError
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			_, err := w.Write(prettyJSON.Bytes())
			if err != nil {
				log.Println("JSON write error: ", err)
			}
		}
	} else {
		_, err := w.Write(res)
		if err != nil {
			log.Println("JSON write error: ", err)
		}
	}
}

func main() {
	log.SetFlags(23)

	kingpin.Version(appVersion("hi"))
	kingpin.Parse()

	log.Printf("listen->'%s'  fileds->'%s'  indent->'%v'\n", *listenAddress, *onlyFields, *enableIndent)

	if len(*onlyFields) > 0 {
		fieldsReg := regexp.MustCompile("(?i)" + *onlyFields)
		// 匹配环境变量
		for _, e := range os.Environ() {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 {
				continue
			} else {
				if fieldsReg.MatchString(parts[0]) {
					envMap[parts[0]] = parts[1]
					log.Printf("Field from ENV: %s=%s\n", parts[0], parts[1])
				}
			}
		}
	} else {
		if hostName := os.Getenv("HOSTNAME"); hostName != "" {
			envMap["HOSTNAME"] = hostName
		}

		if podName := os.Getenv("POD_NAME"); podName != "" {
			envMap["POD_NAME"] = podName
		}
	}

	log.Fatal(http.ListenAndServe(*listenAddress, http.HandlerFunc(SayHi)))
}
