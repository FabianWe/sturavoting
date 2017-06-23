// The MIT License (MIT)

// Copyright (c) 2017 Fabian Wenzelmann

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/FabianWe/sturavoting"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func main() {
	configDirPtr := flag.String("config", "./config", "Directory to store the configuration files.")
	flag.Parse()
	configDir, configDirParseErr := filepath.Abs(*configDirPtr)
	if configDirParseErr != nil {
		log.WithError(configDirParseErr).Fatal("Can't parse config dir path: ", configDir)
	}
	appContext, configErr := sturavoting.ParseConfig(configDir)
	if configErr != nil {
		log.WithError(configErr).Fatal("Can't parse config file(s)")
	}
	// categories, catErr := sturavoting.ListCategories(appContext)
	// if catErr != nil {
	// 	log.Fatal(catErr)
	// }
	// for _, c := range categories {
	// 	fmt.Println(c)
	// }
	// revs, revsErr := sturavoting.ListVotersRevision(appContext, 1)
	// if revsErr != nil {
	// 	log.Fatal(revsErr)
	// }
	// for _, r := range revs {
	// 	fmt.Println(r)
	// }
	// f, openErr := os.Open("examples/voters.txt")
	// if openErr != nil {
	// 	log.Fatal(openErr)
	// }
	// defer f.Close()
	// voters, votersParseErr := sturavoting.ParseVoters(f)
	// if votersParseErr != nil {
	// 	log.Fatal(votersParseErr)
	// }
	// votersInsertErr := sturavoting.InsertVoters(appContext, 1, voters)
	// if votersInsertErr != nil {
	// 	log.Fatal(votersInsertErr)
	// }
	votersList, votersGetErr := sturavoting.ListVoters(appContext, 1)
	if votersGetErr != nil {
		log.Fatal(votersGetErr)
	}
	for _, v := range votersList {
		fmt.Println(v)
	}
}
