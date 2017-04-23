/*
    Copyright (C) Jens Ramhorst
	  This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"fmt"
	"github.com/secsy/goftp"
	"io/ioutil"
	"log"
	"os"
	"smartpi"
	"strings"
	"time"
)

func main() {

	config := smartpi.NewConfig()

	if config.FTPupload == 0 {
		os.Exit(0)
	}

	startDate := time.Now()

	location := time.Now().Location()

	lastdate, err := ioutil.ReadFile("/var/smartpi/csvftp")
	if err == nil {
		startDate, err = time.ParseInLocation("2006-01-02 15:04:05", string(lastdate), location)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		startDate = time.Now().AddDate(0, 0, -1)
	}

	// startDate = startDate.UTC()
	endDate := time.Now()

	csvfile := bytes.NewBufferString(smartpi.CreateCSV(startDate, endDate))

	ftpconfig := goftp.Config{
		User:               config.FTPuser,
		Password:           config.FTPpass,
		ConnectionsPerHost: 10,
		Logger:             os.Stderr,
	}

	client, err := goftp.DialConfig(ftpconfig, config.FTPserver)
	if err != nil {
		panic(err)
	}

	ftp_path := config.FTPpath + config.Serial
	pathlist := strings.Split(ftp_path, "/")
	for i := 0; i < len(pathlist); i++ {
		if len(pathlist[i]) == 0 {
			pathlist = append(pathlist[:i], pathlist[i+1:]...)
		}
	}

	workingpath := "/"
	createpath := ""

	for j := 0; j < len(pathlist); j++ {
		if j > 0 {
			workingpath = workingpath + pathlist[j-1] + "/"
		}
		createpath = createpath + "/" + pathlist[j]

		files, err := client.ReadDir(workingpath)
		smartpi.Checklog(err)
		fileexist := 0
		for _, file := range files {
			if file.IsDir() && file.Name() == pathlist[j] {
				fileexist = 1
			}

		}

		if fileexist == 0 {
			dir, _ := client.Mkdir(createpath)
			fmt.Println("\n\n\n" + dir)
		}

	}

	filename := time.Now().Format("20060102150405") + "_" + config.Serial + ".csv"
	err = client.Store(createpath+"/"+filename, csvfile)
	if err != nil {
		panic(err)
	} else {
		err = ioutil.WriteFile("/var/smartpi/csvftp", []byte(endDate.Local().Format("2006-01-02 15:04:05")), 0644)
		if err != nil {
			panic(err)
		}
	}

}
