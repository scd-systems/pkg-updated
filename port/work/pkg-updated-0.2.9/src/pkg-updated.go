// 1727 "pkg-updated.nw"
/*
Copyright (c) SCD-SYSTEMS.NET

The Regents of the University of California.
All rights reserved.

Redistribution and use in source and binary forms, with
or without modification, are permitted provided that the
following conditions are met:

1. Redistribution's of source code must retain the above
copyright notice, this list of conditions and
the following disclaimer.

2. Redistribution's in binary form must reproduce the above
copyright notice, this list of conditions and the following
disclaimer in the documentation and/or other materials
provided with the distribution.

3. All advertising materials mentioning features or use
of this software must display the following
acknowledgement: "This product includes software developed
by the University of California, Berkeley and
its contributors."

4. Neither the name of the University nor the names
of its contributors may be used to endorse or promote
products derived from this software without specific prior
written permission.

THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING,
BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
MERCHANT-ABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS
BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// 1676 "pkg-updated.nw"
package main

import (

	// 203 "pkg-updated.nw"
	"encoding/json"
	// 280 "pkg-updated.nw"
	"fmt"
	"os"
	"strconv"
	// 293 "pkg-updated.nw"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	// 763 "pkg-updated.nw"
	"bytes"
	"os/exec"
	// 1122 "pkg-updated.nw"
	"regexp"
	"time"
	// 1271 "pkg-updated.nw"
	"io/ioutil"
	// 1387 "pkg-updated.nw"
	"log"
	// 1508 "pkg-updated.nw"
	"os/user"
	// 1550 "pkg-updated.nw"
	"flag"
	// 1680 "pkg-updated.nw"
)

// 8 "pkg-updated.nw"
var (
	MAJOR_VERSION = 0
	MINOR_VERSION = 2
	PATCH_VERSION = 7
)

// 174 "pkg-updated.nw"
const config_file = "./pkg-updated.conf"

// 180 "pkg-updated.nw"
var config struct {
	RecurTime                       string   `json:"schedule"`
	StrictRecurTime                 bool     `json:"schedule-in-time"`
	ExcludePackages                 []string `json:"exclude-packages"`
	CreateReport                    bool     `json:"create-report"`
	ClearSyncDatabaseEnabled        bool     `json:"fresh-db-sync-on-start"`
	DoFreebsdUpdate                 bool     `json:"do-freebsd-update"`
	RestartDaemons                  bool     `json:"restart-daemons"`
	DowngradePackageOnFailedRestart bool     `json:"downgrade-package-on-failed-restart"`
	UseSudo                         bool     `json:"use-sudo"`
	ArchiveEnable                   bool     `json:"pkg-archive-enable"`
	ArchiveFile                     string   `json:"pkg-archive-directory"`
	PkgDatabaseFile                 string   `json:"pkg-database-file"`
	DatabaseFile                    string   `json:"database-file"`
	ReportDatabaseFile              string   `json:"report-database-file"`
	Param_DebugMode                 *bool
	FileExistsNoLog                 bool
}

// 1393 "pkg-updated.nw"
const LOG_FATAL = "FATAL"
const LOG_FATAL2 = "FATAL2"
const LOG_DEBUG = "DEBUG"
const LOG_INFO = "INFO"
const LOG_ERROR = "ERROR"
const LOG_EVENT = "EVENT"
const LOG_STDOUT = "CONSOLE_STDOUT"
const LOG_STDERR = "CONSOLE_STDERR"

// 1514 "pkg-updated.nw"
func Check() {
	var ret int

	account, err := user.Current()
	if err != nil {
		logging(LOG_FATAL, "check", fmt.Sprintf("Could not detect user id: %s", err))
	}

	if account.Uid != "0" {
		if config.UseSudo == false {
			logging(LOG_EVENT, "check", "Warning: Program started as user without sudo usage, maybe it will not work !!!")
		}
	}

	if config.UseSudo == true {
		ret = FileExists("/usr/local/bin/sudo")
		if ret != 0 {
			logging(LOG_FATAL, "check", "Error: No sudo binary (/usr/local/bin/sudo ) found, please install sudo")
		}
	}
}

// 207 "pkg-updated.nw"
func ReadConfig() {
	configfile, err := os.Open(config_file)
	if err != nil {
		logging(LOG_FATAL, "readconfig", fmt.Sprintf("Cannot open config file: %s", config_file))
	}
	defer configfile.Close()

	jsonParser := json.NewDecoder(configfile)

	if err = jsonParser.Decode(&config); err != nil {
		logging(LOG_FATAL, "readconfig", fmt.Sprint("Failed to read/parse config: %s", err))
	}
	logging(LOG_DEBUG, "readconfig-parsed", fmt.Sprint(config))
}

// 302 "pkg-updated.nw"
func FileExists(filename string) int {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if config.FileExistsNoLog == false {
			logging(LOG_ERROR, "fileexists-stat", fmt.Sprint(err))
		} else {
			config.FileExistsNoLog = false
		}
		return -1
	}
	return 0
}

func OpenDB(filename string) *sql.DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		logging(LOG_ERROR, "opendb", fmt.Sprint(err))
		log.Fatal(err)
	}
	return db
}

func CreateDatabase(db *sql.DB, id int) int {

	// 573 "pkg-updated.nw"
	var DBSchema []string
	DBSchema = make([]string, 2)
	DBSchema[0] = "CREATE TABLE packages (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL UNIQUE, origin TEXT, version TEXT NOT NULL, status TEXT NOT NULL, archivepath TEXT, lockstatus TEXT); CREATE INDEX package_name ON packages(name COLLATE NOCASE);CREATE TABLE services (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL, svccmd TEXT, enabled TEXT);"
	// 585 "pkg-updated.nw"
	DBSchema[1] = "CREATE TABLE report (id INTEGER NOT NULL PRIMARY KEY, timestamp INTEGER, eventtype TEXT, facility TEXT, message TEXT NOT NULL); CREATE INDEX report_type ON report (eventtype COLLATE NOCASE);"

	// 328 "pkg-updated.nw"
	_, err := db.Exec(DBSchema[id])
	if err != nil {
		logging(LOG_ERROR, "createdb", fmt.Sprintf("Schema: %s | error: %s", DBSchema[id], err))
		return -1
	}
	return 0
}

// 350 "pkg-updated.nw"
func CountRows(db *sql.DB, table string, column string, search string) int {
	var result int

	err := db.QueryRow("SELECT count(*) FROM " + table + " WHERE " + column + " = '" + search + "';").Scan(&result)
	if err != nil {
		logging(LOG_ERROR, "countrows", fmt.Sprint(err))
		return -1
	}
	logging(LOG_DEBUG, "countrows", fmt.Sprintf("sql query: SELECT count(*) FROM %s WHERE %s = %s | count(*): %d", table, column, search, result))
	return result
}

// 366 "pkg-updated.nw"
func GetPackageInfo(db *sql.DB, field string, pkgname string, opts ...string) (string, error) {
	var (
		result string
		res    *string
		err    error
		stmt   *sql.Stmt
		count  int
	)

	if len(opts) > 0 {
		if len(opts) < 2 {
			count = CountRows(db, opts[0], "name", pkgname)
		} else {
			count = CountRows(db, opts[0], opts[1], pkgname)
		}
	} else {
		count = CountRows(db, "packages", "name", pkgname)
	}

	if count < 0 {
		result = "ENOSQLOUT"
		return result, nil
	}

	if count == 0 {
		result = "ENOEXIST"
		return result, nil
	}

	if len(opts) > 0 {
		if len(opts) < 2 {
			stmt, err = db.Prepare("SELECT " + field + " FROM " + opts[0] + " WHERE name = ?")
		} else {
			stmt, err = db.Prepare("SELECT " + field + " FROM " + opts[0] + " WHERE " + opts[1] + " = ?")
		}
	} else {
		stmt, err = db.Prepare("SELECT " + field + " FROM packages WHERE name = ?")
	}
	if err != nil {
		log.Fatal(err)
		return result, fmt.Errorf("%s", err)
	}

	err = stmt.QueryRow(pkgname).Scan(&res)

	if res != nil {
		result = *res
	} else {
		result = "NULL"
		return result, nil
	}
	if err != nil {
		log.Fatal(err)
		return result, fmt.Errorf("%s", err)
	}

	return result, nil
}

// 429 "pkg-updated.nw"
func AddPackage(db *sql.DB, name string, origin string, version string, status string) int {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into packages(name, origin, version, status) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, origin, version, status)

	if err != nil {
		log.Fatal(err)
		return -1
	}
	tx.Commit()

	return 0
}

// 455 "pkg-updated.nw"
func UpdatePackage(db *sql.DB, set_field string, set_value string, where_field string, where_value string, opts ...string) int {
	var query string

	tx, err := db.Begin()
	if err != nil {
		logging(LOG_ERROR, "updatepackage", fmt.Sprint(err))
	}

	if len(opts) > 0 {
		query = "UPDATE " + opts[0] + " SET " + set_field + " = ? WHERE " + where_field + " = ?"
	} else {
		query = "UPDATE packages SET " + set_field + " = ? WHERE " + where_field + " = ?"
	}

	logging(LOG_DEBUG, "updatepackage", fmt.Sprintf("sql query: %s | Params: %s;%s", query, set_value, where_value))

	stmt, err := tx.Prepare(query)
	if err != nil {
		logging(LOG_ERROR, "updatepackage", fmt.Sprint(err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(set_value, where_value)
	if err != nil {
		logging(LOG_ERROR, "updatepackage", fmt.Sprint(err))
		return -1
	}
	tx.Commit()
	return 0
}

// 490 "pkg-updated.nw"
func GetAllPackages(db *sql.DB) ([]string, error) {
	var (
		result []string
		name   string
	)

	rows, err := db.Query("SELECT name FROM packages")
	if err != nil {
		logging(LOG_ERROR, "getallpackages", fmt.Sprint(err))
	}
	defer rows.Close()

	result = make([]string, 1)
	tmp_result := make([]string, 1)
	i := 0
	for rows.Next() {
		rows.Scan(&name)
		tmp_result[i] = name
		copy(result, tmp_result)
		tmp_result = make([]string, len(result)+1)
		copy(tmp_result, result)
		result = make([]string, len(result)+1)
		i++

	}
	result = make([]string, len(tmp_result)-1)
	copy(result, tmp_result)
	return result, err
}

// 526 "pkg-updated.nw"
func AddService(db *sql.DB, name string, svccmd string, enabled int) int {
	tx, err := db.Begin()
	if err != nil {
		logging(LOG_ERROR, "addservice", fmt.Sprint(err))
	}
	stmt, err := tx.Prepare("insert into services (name, svccmd, enabled) values (?, ?, ?)")
	if err != nil {
		logging(LOG_ERROR, "addservice", fmt.Sprint(err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, svccmd, enabled)

	if err != nil {
		logging(LOG_ERROR, "addservice", fmt.Sprint(err))
		return -1
	}
	tx.Commit()

	return 0
}

// 1456 "pkg-updated.nw"
func AddLogToDB(recordtime time.Time, logtype string, facility string, msg string) int {
	reportdb := OpenDB(config.ReportDatabaseFile)
	defer reportdb.Close()

	tx, err := reportdb.Begin()
	if err != nil {
		logging(LOG_FATAL2, "addtologdb", fmt.Sprintf("Error to open reportdb: %s", err))
		return -1
	}
	stmt, err := tx.Prepare("insert into report(timestamp, eventtype, facility, message) values(?, ?, ?, ?)")
	if err != nil {
		logging(LOG_FATAL2, "addtologdb", fmt.Sprintf("Can not insert into reportdb: %s", err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(recordtime.Unix(), logtype, facility, msg)

	if err != nil {
		logging(LOG_FATAL2, "addtologdb", fmt.Sprintf("Can not insert into reportdb: %s", err))
		return -1
	}
	tx.Commit()
	return 0
}

// 1409 "pkg-updated.nw"
func logging(logtype string, facility string, msg string) {
	recordtime := time.Now()
	die := 0

	// Normal: events,
	// Debug: ALL

	if *config.Param_DebugMode == true {
		fmt.Printf("[%s][%s] at %s: %s\n", logtype, facility, recordtime, msg)
	} else {
		if logtype == LOG_EVENT {
			fmt.Printf("%s\n", msg)
		}
	}

	if logtype == "FATAL" {
		die = 1
	}
	if logtype == "FATAL2" {
		fmt.Printf("%s\n", msg)
		os.Exit(2)
	}

	config.FileExistsNoLog = true
	ret := FileExists(config.ReportDatabaseFile)
	if ret == 0 {
		ret := AddLogToDB(recordtime, logtype, facility, msg)
		if ret != 0 {
			if *config.Param_DebugMode == true {
				fmt.Printf("[ERROR][logging] at %s: Report DB not working\n", recordtime)
			} else {
				fmt.Printf("Report DB not working\n")
			}
		}
	}

	/* Need to die if fatal error happend */
	if die >= 1 {
		os.Exit(die)
	}
}

// 1565 "pkg-updated.nw"
func RunCmd(cmd string, opts ...string) (string, error) {
	var (
		cmdName string
		cmdArgs []string
		cmdOut  []byte
		err     error
		cmdArg1 []string
	)
	cmdName = "pkg"

	switch cmd {
	case "install":
		cmdArg1 = []string{"install", "-y", "-f"}
	case "update":
		cmdArg1 = []string{"update", "-y", "-f"}
	case "lock":
		cmdArg1 = []string{"lock", "-y"}
	case "unlock":
		cmdArg1 = []string{"unlock", "-y"}
	case "upgrade":
		cmdArg1 = []string{"upgrade", "-y"}
	case "create":
		cmdArg1 = []string{"create", "-o"}
	case "version":
		cmdArg1 = []string{"version", "-Rov"}
	case "which":
		cmdArg1 = []string{"which", "-qo"}
	case "sleep":
		_, err = exec.Command("sleep", opts[0]).CombinedOutput()
		return "wakeup", err
	case "service":
		cmdName = "service"
		cmdArg1 = []string{"-e"}
	}

	if config.UseSudo == true {
		cmdArgs = make([]string, len(cmdArg1)+1)
		cmdArgs[0] = cmdName
		cmdName = "sudo"
		copy(cmdArgs[1:], cmdArg1)
	} else {
		cmdArgs = make([]string, len(cmdArg1))
		copy(cmdArgs, cmdArg1)
	}

	cmdArgs = append(cmdArgs, opts...)

	var cmdline bytes.Buffer
	cmdline.WriteString(cmdName)
	for _, value := range cmdArgs {
		cmdline.WriteString(" ")
		cmdline.WriteString(value)
	}
	logging(LOG_DEBUG, "runcmd", fmt.Sprintf("cmd: % ", cmdline))

	cmdOut, err = exec.Command(cmdName, cmdArgs...).CombinedOutput()

	return string(cmdOut), err
}

// 1629 "pkg-updated.nw"
func chop(s string) string {
	return s[0 : len(s)-1]
}

// 668 "pkg-updated.nw"
func SyncPkgDatabases(db *sql.DB) error {
	dbpkg := OpenDB(config.PkgDatabaseFile)
	defer dbpkg.Close()

	var (
		lockstatus string
		pkglist    []string
		err        error
		request    string
	)

	pkglist, err = GetAllPackages(db)

	if len(pkglist) > 1 {
		for _, name := range pkglist {
			if lockstatus, err = GetPackageInfo(db, "lockstatus", name); err != nil {
				logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Current Lockstatus Error: %s", err))
			}
			if lockstatus == "1" {
				logging(LOG_INFO, "syncpkgdatabase", fmt.Sprintf("Unlock excluded packages before sync: %s", name))
				LockPackage(db, 0, name)
			}
		}
	}

	if config.ClearSyncDatabaseEnabled == true {
		logging(LOG_INFO, "syncpkgdatabase", "Fresh pkg databases syncronize")
		_, err = db.Exec("DELETE FROM packages")
		if err != nil {
			logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Clear pkg-updated database: failed | Error: %s", err))
			return err
		}
		logging(LOG_INFO, "syncpkgdatabase", "Clear pkg-updated database: done")
	}

	rows, err := dbpkg.Query("SELECT name, version, origin, locked FROM packages")
	if err != nil {
		logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Sync pkg database: failed | Error: %s", err))
	}
	defer rows.Close()

	var (
		name    string
		version string
		origin  string
		locked  int
	)

	// 729 "pkg-updated.nw"
	for rows.Next() {
		rows.Scan(&name, &version, &origin, &locked)
		request, err = GetPackageInfo(db, "name", name)
		if request != "ENOEXIST" {
			request, err = GetPackageInfo(db, "version", name)
			if request != version {
				UpdatePackage(db, "version", version, "name", name)
			}
			request, err = GetPackageInfo(db, "origin", name)
			if request != origin {
				UpdatePackage(db, "origin", origin, "name", name)
			}
			request, err = GetPackageInfo(db, "status", name)
			if request == "update-available" {
				UpdatePackage(db, "status", "up-to-date", "name", name)
			}
		} else {
			AddPackage(db, name, origin, version, "up-to-date")
		}

		if locked != 0 {
			UpdatePackage(db, "lockstatus", "2", "name", name)
		}
	}
	logging(LOG_INFO, "syncpkgdatabase", "Sync pkg database: done")
	return nil
}

// 772 "pkg-updated.nw"
func CheckUpdates(db *sql.DB) {
	var (
		cmdOut  string
		err     error
		pkgname bytes.Buffer
	)

	cmdOut, err = RunCmd("version")

	if err != nil {
		logging(LOG_ERROR, "checkupdates", fmt.Sprintf("There was an error running pkg command: %s", err))
	}

	// 60 == '<'
	// 10 == SPACE
	// 32 == C.R.
	for n := 0; n < len(cmdOut); n++ {
		if cmdOut[n] == 60 {
			UpdatePackage(db, "status", "update-available", "origin", pkgname.String())
			pkgname.Reset()
			continue
		}
		if cmdOut[n] == 10 {
			pkgname.Reset()
			continue
		}
		if cmdOut[n] != 32 {
			pkgname.WriteString(string(cmdOut[n]))
		}
	}
}

// 999 "pkg-updated.nw"
func LockPackage(db *sql.DB, lock int, name string) string {
	var (
		lockstatus string
		err        error
	)

	if lockstatus, err = GetPackageInfo(db, "lockstatus", name); err != nil {
		logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Error: %v", err))
	}
	if lockstatus == "ENOEXIST" {
		lockstatus = "Package not exists"
		return lockstatus
	}

	if lockstatus == "2" {
		lockstatus = "systemlocked"
		return lockstatus
	}

	// 1024 "pkg-updated.nw"
	switch lock {
	case 0:
		if _, err = RunCmd("unlock", name); err != nil {
			logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Unlock pkg error: %v", err))
		}

	case 1:
		if _, err = RunCmd("lock", name); err != nil {
			logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Lock pkg error: %v", err))
		}
	default:
		lockstatus = "Not supported lock mode"
		return lockstatus
	}

	setlock := strconv.Itoa(lock)

	UpdatePackage(db, "lockstatus", setlock, "name", name)

	lockstatus, err = GetPackageInfo(db, "lockstatus", name)

	switch lockstatus {
	case "NULL":
		lockstatus = "Not locked"
	case "ENOEXIST":
		lockstatus = "Packages does not exists or unlocked"
	case "1":
		lockstatus = "Locked"
	case "2":
		lockstatus = "Systemlocked"
	default:
		lockstatus = "Unknown lock status"
	}
	return lockstatus
}

// 1064 "pkg-updated.nw"
func LockExclude(db *sql.DB, lock int) {
	var mode string
	switch lock {
	case 0:
		mode = "Unlock"
	case 1:
		mode = "Lock"
	default:
		logging(LOG_ERROR, "lockexclude", fmt.Sprintf("The lock id is not supported"))
		return
	}
	for _, name := range config.ExcludePackages {
		logging(LOG_INFO, "lockexclude", fmt.Sprintf("%s exclude package %s: %s", mode, name, LockPackage(db, lock, name)))
	}
}

// 923 "pkg-updated.nw"
func RollbackPackage(db *sql.DB, name string) {
	var (
		path   string
		cmdOut string
		err    error
	)

	if path, err = GetPackageInfo(db, "archivepath", name); err != nil {
		logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Error to get archivepath: %v", err))
	}

	if path == "NULL" {
		logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Rollback Error: No rollback pkg available for package %s", name))
		return
	}
	logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Rollback package: %s , pkg: %s", name, path))

	if cmdOut, err = RunCmd("install", path); err != nil {
		logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Install error: %s", err))
	}

	/* Set status for success rollback
	 else {
		UpdatePackage(db, "status", "update-available", "name", name);
	}
	*/
	logging(LOG_DEBUG, "rollbackpackage", fmt.Sprintf("Output: %s", string(cmdOut)))
	return
}

// 1278 "pkg-updated.nw"
func ScanEnabledServices(db *sql.DB) {
	if config.RestartDaemons == true {
		var (
			pkgorigin string
			pkgname   string
			ret       int
			err       error
			enabled   int
			cmdout    string
		)
		enabled = 0

		_, err = db.Exec("DELETE FROM services")
		if err != nil {
			logging(LOG_ERROR, "scanenabledservices", fmt.Sprintf("Truncated service db: failed | Error: %s", err))
			return
		}
		logging(LOG_INFO, "scanenabledservices", "Truncated service db: done")

		absolute := "/usr/local/etc/rc.d/"
		files, _ := ioutil.ReadDir(absolute)

		for _, f := range files {
			if f.IsDir() == false {
				svc := absolute + f.Name()
				pkgorigin, err = RunCmd("which", svc)
				if err == nil {

					pkgname, err = GetPackageInfo(db, "name", chop(pkgorigin), "packages", "origin")

					logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Pkgname for origin %s: %s", chop(pkgorigin), pkgname))
					logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Add %s: %s", pkgname, svc))

					ret, cmdout, err = ScanScript(svc, cmdout)
					if ret != 0 {
						logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Service: %s is enabled", svc))
						enabled = 1
					} else {
						logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Service: %s is not enabled", svc))
						enabled = 0
					}
					AddService(db, pkgname, svc, enabled)
				}
			}
		}
	}
}

// 1334 "pkg-updated.nw"
func ScanScript(path string, preout string) (int, string, error) {
	var (
		err error
	)
	var cmdOut string
	if preout != "" {
		cmdOut = preout
	} else {
		cmdOut, err = RunCmd("service", "")
	}

	if err != nil {
		return -1, cmdOut, err
	}

	var buffer bytes.Buffer
	for _, val := range cmdOut {
		if val == 10 {
			if buffer.String() == path {
				return 1, cmdOut, nil
			}
			buffer.Reset()
			continue
		}
		buffer.WriteString(string(val))
	}
	return 0, cmdOut, nil
}

// 1367 "pkg-updated.nw"
func RestartService(svc string) (int, error) {
	var (
		ret int
	)
	cmdOut, err := RunCmd("restart", svc)
	if err != nil {
		logging(LOG_ERROR, "restartservice", fmt.Sprintf("Service: Could not restart: %s | cmdOut: %v", svc, cmdOut))
		ret = 1
	}
	return ret, err
}

// 813 "pkg-updated.nw"
func GetUpdateList(db *sql.DB) ([]string, error) {
	var list []string
	var name string
	list = make([]string, 0)
	var nlist []string

	rows, err := db.Query("SELECT name FROM packages WHERE status = $1", "update-available")
	if err != nil {
		logging(LOG_ERROR, "getupdatelist", fmt.Sprint(err))
		return list, fmt.Errorf("%s", err)
	}
	defer rows.Close()

	for rows.Next() {
		nlist = make([]string, len(list)+1)
		copy(nlist, list)
		list = nlist
		err = rows.Scan(&name)
		if err != nil {
			logging(LOG_ERROR, "getupdatelist", fmt.Sprint(err))
			return list, fmt.Errorf("%s", err)
		}
		list[len(list)-1] = name
	}
	return list, nil
}

// 846 "pkg-updated.nw"
func SavePackages(db *sql.DB) {
	var (
		version string
		origin  string
		path    string
		index   int
	)

	updatelist, err := GetUpdateList(db)
	if err != nil {
		logging(LOG_ERROR, "savepackages", fmt.Sprintf("GetUpdateList(): %v", err))
	}

	// 865 "pkg-updated.nw"
	for _, pkg := range updatelist {
		version, err = GetPackageInfo(db, "version", pkg)
		path = config.ArchiveFile + "/" + pkg + "-" + version + ".txz"
		if _, err := os.Stat(path); err != nil {
			RunCmd("create", config.ArchiveFile, pkg)
			index++
		}
	}

	for _, pkg := range updatelist {
		version, err = GetPackageInfo(db, "version", pkg)
		path = config.ArchiveFile + "/" + pkg + "-" + version + ".txz"
		if _, err := os.Stat(path); err == nil {
			origin, err = GetPackageInfo(db, "origin", pkg)
			// 884 "pkg-updated.nw"
			//			if err != nil {
			UpdatePackage(db, "archivepath", path, "origin", origin)
			//				UpdatePackage(db, origin, "archive", path)
			//			}
		} else {
			logging(LOG_ERROR, "savepackages", fmt.Sprintf("Could not found rollback package file: %s", path))
		}
	}
}

// 900 "pkg-updated.nw"
func Upgrade() {
	var (
		cmdOut string
		err    error
	)
	cmdOut, err = RunCmd("upgrade")
	logging(LOG_STDOUT, "upgrade", fmt.Sprintf("Output Upgrade(): %s", cmdOut))
	if err != nil {
		logging(LOG_ERROR, "upgrade", fmt.Sprintf("Error: ", string(cmdOut), err))
	}
}

// 1127 "pkg-updated.nw"
func Scheduler(db *sql.DB) {
	var (
		err            error
		recur          time.Time
		AlreadyRun     bool
		Ticker         int
		TimeNow        time.Time
		LastCheck      time.Time
		romandaytime   bool
		sleeptimer     int
		tolerance      int
		Run            bool
		diff           int64
		diff2          int
		RecurUnixStamp int
	)

	sleeptimer = 60
	tolerance = 10

	romandaytime = false
	re := regexp.MustCompile("[0-9][0-9][a|A|p|P][m|M]$")
	ret := re.FindString(config.RecurTime)

	if len(ret) > 0 {
		romandaytime = true
	}

	if romandaytime {
		recur, err = time.Parse(time.Kitchen, config.RecurTime)
	} else {
		var buffer bytes.Buffer
		buffer.WriteString("24 Dec 00 ")
		buffer.WriteString(config.RecurTime)
		buffer.WriteString(" UTC")
		value := buffer.String()
		recur, err = time.Parse(time.RFC822, value)
	}
	if err != nil {
		logging(LOG_ERROR, "scheduler", fmt.Sprintf("Cannot parse schedule value: %s", err))
	}

	logging(LOG_DEBUG, "scheduler", fmt.Sprintf("Recurring time: %v:%v", recur.Hour(), recur.Minute()))

	LastCheck = time.Now()
	RecurUnixStamp = (recur.Hour() * 3600) + (recur.Minute() * 60)

	for {
		TimeNow = time.Now()
		diff = LastCheck.Unix() - TimeNow.Unix()
		Ticker = (TimeNow.Hour() * 3600) + (TimeNow.Minute() * 60) + (TimeNow.Second())

		LastCheck = TimeNow
		if (diff <= (0 - int64(sleeptimer) - int64(tolerance))) || (diff > (int64(sleeptimer) + int64(tolerance))) {
			logging(LOG_INFO, "scheduler", "Time jump detected")
			logging(LOG_INFO, "scheduler", "Set AlreadyRun to false")
			AlreadyRun = false
		}

		// Only run if time has been arrived

		if (AlreadyRun == false) && ((TimeNow.Hour() > recur.Hour()) || ((TimeNow.Hour() == recur.Hour()) && (TimeNow.Minute() >= recur.Minute()))) {

			// If strict time is configured, check the 5 minute tolerance

			if config.StrictRecurTime == true {
				Run = false
				diff2 = Ticker - RecurUnixStamp
				if (diff2 > 0) && (diff2 < 300) {
					Run = true
				}
			} else {
				Run = true
			}

			if Run == true {
				logging(LOG_EVENT, "scheduler", fmt.Sprintf("Scheduled Time reached, start job"))

				SyncPkgDatabases(db)
				CheckUpdates(db)
				if config.ArchiveEnable {
					SavePackages(db)
				}
				LockExclude(db, 1)
				ScanEnabledServices(db)

				Upgrade()

				/* Second RUN
					Need to verify how it behaves

				if config.ClearSyncDatabaseEnabled == true {
					config.ClearSyncDatabaseEnabled = false
				}
				SyncPkgDatabases(db)
				*/

				AlreadyRun = true
				Run = false
			}
		}

		if (Ticker >= (86400 - sleeptimer - tolerance)) && (AlreadyRun == true) {
			logging(LOG_INFO, "scheduler", fmt.Sprintf("New day detect, set AlreadyRunn to false"))
			AlreadyRun = false
		}

		logging(LOG_DEBUG, "scheduler", fmt.Sprintf("Recur: %v:%v | Now: %v:%v | AlreadyRun: %v", recur.Hour(), recur.Minute(), TimeNow.Hour(), TimeNow.Minute(), AlreadyRun))

		RunCmd("sleep", strconv.Itoa(sleeptimer))
	}
}

// 595 "pkg-updated.nw"
func main() {

	// 1554 "pkg-updated.nw"
	config.Param_DebugMode = flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()

	// 607 "pkg-updated.nw"
	ReadConfig()
	Check()

	logging(LOG_EVENT, "main", "Started pkg-updated")

	// Need to find a better code for open file only if exists and not before test file.

	var ret int
	db := OpenDB(config.DatabaseFile)
	defer db.Close()
	ret = FileExists(config.DatabaseFile)
	logging(LOG_INFO, "main", fmt.Sprintf("Check packages Database %s: %d", config.DatabaseFile, ret))

	if ret != 0 {
		ret = CreateDatabase(db, 0)
		if ret != 0 {
			logging(LOG_ERROR, "main", "Create packages database: failed")
			os.Exit(2)
		} else {
			logging(LOG_INFO, "main", "Create packages database: done")
		}
	}

	ret = FileExists(config.ReportDatabaseFile)
	logging(LOG_INFO, "main", fmt.Sprintf("Check report Database %s: %d", config.ReportDatabaseFile, ret))

	if ret != 0 {
		reportdb := OpenDB(config.ReportDatabaseFile)
		ret = CreateDatabase(reportdb, 1)
		if ret != 0 {
			logging(LOG_ERROR, "main", "Create report database: failed")
			reportdb.Close()
			os.Exit(2)
		} else {
			logging(LOG_INFO, "main", "Create report database: done")
		}
		reportdb.Close()
	}

	// 652 "pkg-updated.nw"
	go Scheduler(db)
	for {
		RunCmd("sleep", "300")
	}
	// 1090 "pkg-updated.nw"
}
