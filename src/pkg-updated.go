// 2135 "pkg-updated.nw"
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

// 2081 "pkg-updated.nw"
package main

import (

	// 230 "pkg-updated.nw"
	"encoding/json"
	// 315 "pkg-updated.nw"
	"fmt"
	"os"
	"strconv"
	// 328 "pkg-updated.nw"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	// 810 "pkg-updated.nw"
	"bytes"
	"os/exec"
	// 1247 "pkg-updated.nw"
	"regexp"
	"time"
	// 1391 "pkg-updated.nw"
	"io/ioutil"
	// 1567 "pkg-updated.nw"
	"errors"
	"log"
	// 1703 "pkg-updated.nw"
	"os/user"
	// 1804 "pkg-updated.nw"
	"flag"
	// 2085 "pkg-updated.nw"
)

// 9 "pkg-updated.nw"
var (
	MAJOR_VERSION = 0
	MINOR_VERSION = 3
	PATCH_VERSION = 3
)

// 183 "pkg-updated.nw"
const config_file = "/usr/local/etc/pkg-updated/pkg-updated.conf"

// 189 "pkg-updated.nw"
var config struct {
	RecurTime                       string   `json:"schedule"`
	StrictRecurTime                 bool     `json:"schedule-in-time"`
	ExcludePackages                 []string `json:"exclude-packages"`
	RecurDays                       []int    `json:"schedule-days"`
	CreateReport                    bool     `json:"create-report"`
	ClearSyncDatabaseEnabled        bool     `json:"fresh-db-sync-on-start"`
	DoFreebsdUpdate                 bool     `json:"do-freebsd-update"`
	RestartDaemons                  bool     `json:"restart-daemons"`
	DowngradePackageOnFailedRestart bool     `json:"downgrade-package-on-failed-restart"`
	UseSudo                         bool     `json:"use-sudo"`
	ArchiveEnable                   bool     `json:"pkg-archive-enable"`
	ArchivePath                     string   `json:"pkg-archive-directory"`
	PkgDatabaseFile                 string   `json:"pkg-database-file"`
	DatabaseFile                    string   `json:"database-file"`
	ReportDatabaseFile              string   `json:"report-database-file"`
	Param_DebugMode                 *bool
	Param_RunOnce                   *bool
	Param_ConfigFile                *string
	Param_Help                      *bool
	Param_Version                   *bool
	/*
		Param_CreateReport *bool;
		Param_ClearSyncDatabaseEnabled *bool;
		Param_DoFreebsdUpdate *bool;
		Param_RestartDaemons *bool;
		Param_DowngradePackageOnFailedRestart *bool;
		Param_UseSudo *bool;
		Param_ArchiveEnable *bool;
	*/
	Param_ArchivePath        *string
	Param_PkgDatabaseFile    *string
	Param_DatabaseFile       *string
	Param_ReportDatabaseFile *string
	FileExistsNoLog          bool
}

// 1575 "pkg-updated.nw"
const LOG_FATAL = "FATAL"
const LOG_FATAL2 = "FATAL2"
const LOG_DEBUG = "DEBUG"
const LOG_INFO = "INFO"
const LOG_ERROR = "ERROR"
const LOG_EVENT = "EVENT"
const LOG_STDOUT = "CONSOLE_STDOUT"
const LOG_STDERR = "CONSOLE_STDERR"

// 1966 "pkg-updated.nw"
func HelpPage() {
	fmt.Printf("pkg-updated help:\n\n")
	fmt.Printf("Usage: pkg-updated [-option] [-option] ... [-option <FILENAME>] ... \n\n")
	fmt.Printf("Options:\n--------\n")
	fmt.Printf("-help\t\t\t\tShow this page\n")
	fmt.Printf("-config <FILENAME>\t\t\t\tPath to config file to use\n")
	fmt.Printf("-debug\t\t\t\tRuns in debug mode and prints all LOG types\n")
	fmt.Printf("-runone\t\t\t\tDisable scheduler and just run once update procedure\n")
	fmt.Printf("-version\t\t\tShow version and exit\n")
	/*
		fmt.Printf("-reporting\tCreate and use a report db for all events");
		fmt.Printf("-cleardbonstart\tIf pkg-updated db is already exists, truncate all informations before sync\n");
		fmt.Printf("-enableosupdate\tEnable update of OS too\n");
		fmt.Printf("-restartdaemons\tRestart enable services if an update affecte\n");
		fmt.Printf("-enablerollback\tRollback package if service restart failed, require -enablearchive true\n");
		fmt.Printf("-sudo\tUse sudo for all commands\n");
		fmt.Printf("-enablearchive\tCreate a backup package before upgrade, required for rollback\n");
		fmt.Printf("-archivepath <filename>\t\tIn which directory should the pkg backups stored\n");
		fmt.Printf("-pkgdbfile <filename>\t\tThe local pkg database file\n");
		fmt.Printf("-dbfile <filename>\t\tThe pkg-updated database file\n");
		fmt.Printf("-reportdbfile <filename>\tThe report database file\n");
	*/
	os.Exit(0)
}

func Version() {
	fmt.Printf("pkg-update version: %d.%d.%d\n", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)
	os.Exit(0)
}

// 1838 "pkg-updated.nw"
func FileExists(filename string) int {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if config.FileExistsNoLog == false {
			Logging(LOG_ERROR, "fileexists", fmt.Sprint(err))
		} else {
			config.FileExistsNoLog = false
		}
		return -1
	}
	return 0
}

// 1715 "pkg-updated.nw"
func Check() {
	var ret int

	if *config.Param_Help == true {
		HelpPage()
	}

	if *config.Param_Version == true {
		Version()
	}

	if len(*config.Param_ConfigFile) < 1 {
		*config.Param_ConfigFile = config_file
	}

	ret = FileExists(*config.Param_ConfigFile)
	if ret != 0 {
		Logging(LOG_FATAL2, "check", fmt.Sprintf("Could not read config file: %s", *config.Param_ConfigFile))
	}

	ReadConfig()

	/* does not work, because config struct needs to create new and map the new struct
	if (len(*config.Param_ArchivePath) > 1) {
		config.ArchivePath = *config.Param_ArchivePath;
		Logging(LOG_DEBUG, "check", fmt.Sprintf("Set ArchivePath to: %s", config.ArchivePath));
	}
	if (len(*config.Param_PkgDatabaseFile) > 1) {
		config.PkgDatabaseFile = *config.Param_PkgDatabaseFile;
		Logging(LOG_DEBUG, "check", fmt.Sprintf("Set PkgDatabaseFile to: %s", config.PkgDatabaseFile));
	}
	if (len(*config.Param_DatabaseFile) > 1) {
		config.DatabaseFile = *config.Param_DatabaseFile;
		Logging(LOG_DEBUG, "check", fmt.Sprintf("Set DatabaseFile to: %s", config.DatabaseFile));
	}
	if (len(*config.Param_ReportDatabaseFile) > 1) {
		config.ReportDatabaseFile = *config.Param_ReportDatabaseFile;
		Logging(LOG_DEBUG, "check", fmt.Sprintf("Set ReportDatabaseFile to: %s", config.ReportDatabaseFile));
	}
	*/

	account, err := user.Current()
	if err != nil {
		Logging(LOG_FATAL, "check", fmt.Sprintf("Could not detect user id: %s", err))
	}

	if account.Uid != "0" {
		if config.UseSudo == false {
			Logging(LOG_EVENT, "check", "Warning: Program started as user without sudo usage, maybe it will not work !!!")
		}
	}

	if config.UseSudo == true {
		ret = FileExists("/usr/local/bin/sudo")
		if ret != 0 {
			Logging(LOG_FATAL2, "check", "Error: No sudo binary (/usr/local/bin/sudo ) found, please install sudo")
		}
	}
	ret = FileExists(config.PkgDatabaseFile)
	if ret != 0 {
		Logging(LOG_FATAL2, "check", "Error: No local pkg database exists")
	}

	if len(config.RecurDays) > 7 {
		Logging(LOG_FATAL2, "check", "Config-Error: To much entries in schedule-days")
	}
	for _, value := range config.RecurDays {
		if (value < 0) || (value > 7) {
			Logging(LOG_FATAL2, "check", "Config-Error: Wrong day number in schedule-days, allowed only 0 to 7")
		}
	}

}

// 238 "pkg-updated.nw"
func ReadConfig() int {
	Logging(LOG_DEBUG, "readconfig", fmt.Sprintf("Read config file: %s", *config.Param_ConfigFile))

	configfile, err := os.Open(*config.Param_ConfigFile)
	defer configfile.Close()

	if err != nil {
		Logging(LOG_FATAL2, "readconfig", fmt.Sprintf("Cannot open config file: %s", *config.Param_ConfigFile))
	}
	jsonParser := json.NewDecoder(configfile)

	if err = jsonParser.Decode(&config); err != nil {
		Logging(LOG_FATAL2, "readconfig", fmt.Sprintf("Failed to read/parse config: %s", err))
	}
	Logging(LOG_DEBUG, "readconfig-parsed", fmt.Sprint(config))
	return 0
}

// 334 "pkg-updated.nw"
func OpenDB(filename string) *sql.DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		Logging(LOG_ERROR, "opendb", fmt.Sprint(err))
		log.Fatal(err)
	}
	return db
}

func CreateDatabase(db *sql.DB, id int) int {

	// 614 "pkg-updated.nw"
	var DBSchema []string
	DBSchema = make([]string, 2)
	DBSchema[0] = "CREATE TABLE packages (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL UNIQUE, origin TEXT, version TEXT NOT NULL, status TEXT NOT NULL, archivepath TEXT, lockstatus TEXT); CREATE INDEX package_name ON packages(name COLLATE NOCASE);CREATE TABLE services (id INTEGER NOT NULL PRIMARY KEY, name TEXT NOT NULL, svccmd TEXT, enabled TEXT);"
	// 626 "pkg-updated.nw"
	DBSchema[1] = "CREATE TABLE report (id INTEGER NOT NULL PRIMARY KEY, timestamp INTEGER, eventtype TEXT, facility TEXT, message TEXT NOT NULL); CREATE INDEX report_type ON report (eventtype COLLATE NOCASE);"

	// 347 "pkg-updated.nw"
	_, err := db.Exec(DBSchema[id])
	if err != nil {
		Logging(LOG_ERROR, "createdb", fmt.Sprintf("Schema: %s | error: %s", DBSchema[id], err))
		return -1
	}
	return 0
}

// 368 "pkg-updated.nw"
func CountRows(db *sql.DB, table string, column string, search string) int {
	var result int

	err := db.QueryRow("SELECT count(*) FROM " + table + " WHERE " + column + " = '" + search + "';").Scan(&result)
	if err != nil {
		Logging(LOG_ERROR, "countrows", fmt.Sprint(err))
		return -1
	}
	Logging(LOG_DEBUG, "countrows", fmt.Sprintf("sql query: SELECT count(*) FROM %s WHERE %s = %s | count(*): %d", table, column, search, result))
	return result
}

// 384 "pkg-updated.nw"
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
			Logging(LOG_DEBUG, "getpackageinfo", fmt.Sprintf("sql query: SELECT %s FROM %s WHERE name = %s", field, opts[0], pkgname))
			stmt, err = db.Prepare("SELECT " + field + " FROM " + opts[0] + " WHERE name = ?")

		} else {
			Logging(LOG_DEBUG, "getpackageinfo", fmt.Sprintf("sql query: SELECT %s FROM %s WHERE %s = %s", field, opts[0], opts[1], pkgname))
			stmt, err = db.Prepare("SELECT " + field + " FROM " + opts[0] + " WHERE " + opts[1] + " = ?")
		}
	} else {
		Logging(LOG_DEBUG, "getpackageinfo", fmt.Sprintf("sql query: SELECT %s FROM packages WHERE name = %s", field, pkgname))
		stmt, err = db.Prepare("SELECT " + field + " FROM packages WHERE name = ?")
	}

	if err != nil {
		Logging(LOG_FATAL, "getpackageinfo", fmt.Sprintf("Error: %v", err))
		return result, fmt.Errorf("%s", err)
	}

	err = stmt.QueryRow(pkgname).Scan(&res)

	if res != nil {
		result = *res
	} else {
		result = "NULL"
		return result, nil
	}
	Logging(LOG_DEBUG, "getpackageinfo", fmt.Sprintf("sql query result: %s", result))

	if err != nil {
		Logging(LOG_ERROR, "getpackageinfo", fmt.Sprintf("Error: %v", err))
		return result, fmt.Errorf("%s", err)
	}

	return result, nil
}

// 454 "pkg-updated.nw"
func AddPackage(db *sql.DB, name string, origin string, version string, status string) int {
	tx, err := db.Begin()
	if err != nil {
		Logging(LOG_FATAL, "addpackage", fmt.Sprintf("Error: %v", err))
	}
	stmt, err := tx.Prepare("insert into packages(name, origin, version, status) values(?, ?, ?, ?)")

	Logging(LOG_DEBUG, "addpackage", fmt.Sprintf("sql query: INSERT INTO packages (name,origin,version,status) value (%s, %s, %s, %s)", name, origin, version, status))

	if err != nil {
		Logging(LOG_ERROR, "addpackage", fmt.Sprintf("Error: %v", err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, origin, version, status)

	if err != nil {
		Logging(LOG_ERROR, "addpackage", fmt.Sprintf("Error: %v", err))
		return -1
	}
	tx.Commit()

	return 0
}

// 483 "pkg-updated.nw"
func UpdatePackage(db *sql.DB, set_field string, set_value string, where_field string, where_value string, opts ...string) int {
	var query string

	tx, err := db.Begin()
	if err != nil {
		Logging(LOG_FATAL, "updatepackage", fmt.Sprint(err))
	}

	if len(opts) > 0 {
		query = "UPDATE " + opts[0] + " SET " + set_field + " = ? WHERE " + where_field + " = ?"
	} else {
		query = "UPDATE packages SET " + set_field + " = ? WHERE " + where_field + " = ?"
	}

	Logging(LOG_DEBUG, "updatepackage", fmt.Sprintf("sql query: %s | Params: %s;%s", query, set_value, where_value))

	stmt, err := tx.Prepare(query)
	if err != nil {
		Logging(LOG_ERROR, "updatepackage", fmt.Sprintf("Error: %v", err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(set_value, where_value)
	if err != nil {
		Logging(LOG_ERROR, "updatepackage", fmt.Sprintf("Error: %v", err))
		return -1
	}
	tx.Commit()
	return 0
}

// 518 "pkg-updated.nw"
func GetAllPackages(db *sql.DB, opts ...string) ([]string, error) {
	var (
		result []string
		name   string
		err    error
		rows   *sql.Rows
	)

	if len(opts) == 1 {
		err = errors.New("To few paramters for GetAllPackages")
		Logging(LOG_ERROR, "getallpackages", fmt.Sprint(err))
		return result, err
	} else if len(opts) >= 2 {
		Logging(LOG_DEBUG, "getallpackages", fmt.Sprintf("sql query: SELECT %s FROM %s", opts[0], opts[1]))
		rows, err = db.Query("SELECT " + opts[0] + " FROM " + opts[1])
	} else {
		Logging(LOG_DEBUG, "getallpackages", fmt.Sprintf("sql query: SELECT name FROM packages"))
		rows, err = db.Query("SELECT name FROM packages")
	}

	if err != nil {
		Logging(LOG_ERROR, "getallpackages", fmt.Sprintf("Error: %v", err))
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

// 567 "pkg-updated.nw"
func AddService(db *sql.DB, name string, svccmd string, enabled int) int {
	tx, err := db.Begin()
	if err != nil {
		Logging(LOG_ERROR, "addservice", fmt.Sprintf("Error: %v", err))
	}
	Logging(LOG_DEBUG, "addservice", fmt.Sprintf("sql query: INSERT INTO services (name,svccmd,enabled) VALUES (%s,%s,%s)", name, svccmd, enabled))
	stmt, err := tx.Prepare("insert into services (name, svccmd, enabled) values (?, ?, ?)")
	if err != nil {
		Logging(LOG_ERROR, "addservice", fmt.Sprintf("Error: %v", err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, svccmd, enabled)

	if err != nil {
		Logging(LOG_ERROR, "addservice", fmt.Sprintf("Error: %v", err))
		return -1
	}
	tx.Commit()
	return 0
}

// 1638 "pkg-updated.nw"
func AddLogToDB(recordtime time.Time, logtype string, facility string, msg string) int {
	reportdb := OpenDB(config.ReportDatabaseFile)
	defer reportdb.Close()

	tx, err := reportdb.Begin()
	if err != nil {
		Logging(LOG_FATAL2, "addlogtodb", fmt.Sprintf("Error to open reportdb: %s", err))
		return -1
	}
	stmt, err := tx.Prepare("insert into report(timestamp, eventtype, facility, message) values(?, ?, ?, ?)")
	if err != nil {
		Logging(LOG_FATAL2, "addlogtodb", fmt.Sprintf("Can not insert into reportdb: %s", err))
		return -1
	}
	defer stmt.Close()
	_, err = stmt.Exec(recordtime.Unix(), logtype, facility, msg)

	if err != nil {
		Logging(LOG_FATAL2, "addlogtodb", fmt.Sprintf("Can not insert into reportdb: %s", err))
		return -1
	}
	tx.Commit()
	return 0
}

// 1591 "pkg-updated.nw"
func Logging(logtype string, facility string, msg string) {
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

// 1855 "pkg-updated.nw"
func RunCmd(cmd string, opts ...string) (string, error) {
	var (
		cmdName    string
		cmdArgs    []string
		cmdArg1    []string
		cmdTimeout time.Duration
		err        error
	)
	cmdTimeout = (60 * time.Second)
	cmdStdOut := &bytes.Buffer{}
	cmdStdErr := &bytes.Buffer{}

	cmdName = "pkg"
	switch cmd {
	case "install":
		cmdArg1 = []string{"install", "-y", "-f"}
		cmdTimeout = (30 * time.Second)
	case "update":
		cmdArg1 = []string{"update", "-f"}
		cmdTimeout = (60 * time.Second)
	case "lock":
		cmdArg1 = []string{"lock", "-y"}
		cmdTimeout = (10 * time.Second)
	case "unlock":
		cmdArg1 = []string{"unlock", "-y"}
		cmdTimeout = (10 * time.Second)
	case "upgrade":
		cmdArg1 = []string{"upgrade", "-y"}
		cmdTimeout = (300 * time.Second)
	case "create":
		cmdArg1 = []string{"create", "-o"}
		cmdTimeout = (120 * time.Second)
	case "version":
		cmdArg1 = []string{"version", "-Rov"}
		cmdTimeout = (60 * time.Second)
	case "which":
		cmdArg1 = []string{"which", "-qo"}
		cmdTimeout = (10 * time.Second)
	case "sleep":
		_, err = exec.Command("sleep", opts[0]).CombinedOutput()
		return "wakeup", err
	case "service":
		cmdName = "service"
		cmdArg1 = []string{"-e"}
		cmdTimeout = (30 * time.Second)
	case "service_restart":
		cmdName = opts[0]
		opts[0] = ""
		cmdArg1 = []string{"restart"}
		cmdTimeout = (30 * time.Second)
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
	Logging(LOG_DEBUG, "runcmd", fmt.Sprintf("Run command: %s", cmdline.String()))

	cmdExec := exec.Command(cmdName, cmdArgs...)
	cmdExec.Stdout = cmdStdOut
	cmdExec.Stderr = cmdStdErr

	if err = cmdExec.Start(); err != nil {
		Logging(LOG_DEBUG, "runcmd", fmt.Sprintf("Error on start: %s", err))
	}

	// cmd timeout function, need to kill after timeout reached
	go func() {
		time.Sleep(cmdTimeout)
		inerr := cmdExec.Process.Kill()
		if inerr != nil {
			Logging(LOG_FATAL2, "runcmd", fmt.Sprintf("Panic: %s", err))
		}
		Logging(LOG_DEBUG, "runcmd", "Killed running command, timeout reached")
	}()
	err = cmdExec.Wait()

	Logging(LOG_STDOUT, "runcmd", fmt.Sprintf("StdOut [%s]: %s", cmdline.String(), cmdStdOut.String()))
	if cmdStdErr.Len() > 0 {
		Logging(LOG_STDERR, "runcmd", fmt.Sprintf("StdErr [%s]: %s", cmdline.String(), cmdStdErr.String()))
	}

	return cmdStdOut.String(), err
}

// 1958 "pkg-updated.nw"
func chop(s string) string {
	return s[0 : len(s)-1]
}

// 715 "pkg-updated.nw"
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
				Logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Current Lockstatus Error: %s", err))
			}
			if lockstatus == "1" {
				Logging(LOG_INFO, "syncpkgdatabase", fmt.Sprintf("Unlock excluded packages before sync: %s", name))
				LockPackage(db, 0, name)
			}
		}
	}

	if config.ClearSyncDatabaseEnabled == true {
		Logging(LOG_INFO, "syncpkgdatabase", "Fresh pkg databases syncronize")
		_, err = db.Exec("DELETE FROM packages")
		if err != nil {
			Logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Clear pkg-updated database: failed | Error: %s", err))
			return err
		}
		Logging(LOG_INFO, "syncpkgdatabase", "Clear pkg-updated database: done")
	}

	rows, err := dbpkg.Query("SELECT name, version, origin, locked FROM packages")
	if err != nil {
		Logging(LOG_ERROR, "syncpkgdatabase", fmt.Sprintf("Sync pkg database: failed | Error: %s", err))
	}
	defer rows.Close()

	var (
		name    string
		version string
		origin  string
		locked  int
	)

	// 776 "pkg-updated.nw"
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
	Logging(LOG_INFO, "syncpkgdatabase", "Sync pkg database: done")
	return nil
}

// 819 "pkg-updated.nw"
func CheckUpdates(db *sql.DB) bool {
	var (
		cmdOut  string
		err     error
		pkgname bytes.Buffer
		ret     bool
	)

	cmdOut, err = RunCmd("update")
	if err != nil {
		Logging(LOG_ERROR, "checkupdates", fmt.Sprintf("There was an error running pkg command: %s", err))
		return false
	}

	cmdOut, err = RunCmd("version")
	if err != nil {
		Logging(LOG_ERROR, "checkupdates", fmt.Sprintf("There was an error running pkg command: %s", err))
		return false
	}

	// 60 == '<'
	// 10 == SPACE
	// 32 == C.R.
	for n := 0; n < len(cmdOut); n++ {
		if cmdOut[n] == 60 {
			UpdatePackage(db, "status", "update-available", "origin", pkgname.String())
			Logging(LOG_DEBUG, "checkupdates", fmt.Sprintf("Update available for pkg: %s", pkgname.String()))
			pkgname.Reset()
			ret = true
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
	return ret
}

// 1122 "pkg-updated.nw"
func LockPackage(db *sql.DB, lock int, name string) string {
	var (
		lockstatus string
		err        error
	)

	if lockstatus, err = GetPackageInfo(db, "lockstatus", name); err != nil {
		Logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Error: %v", err))
	}
	if lockstatus == "ENOEXIST" {
		lockstatus = "Package not exists"
		return lockstatus
	}

	if lockstatus == "2" {
		lockstatus = "systemlocked"
		return lockstatus
	}

	// 1147 "pkg-updated.nw"
	switch lock {
	case 0:
		if _, err = RunCmd("unlock", name); err != nil {
			Logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Unlock pkg error: %v", err))
		}

	case 1:
		if _, err = RunCmd("lock", name); err != nil {
			Logging(LOG_ERROR, "lockpackage", fmt.Sprintf("Lock pkg error: %v", err))
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

// 1187 "pkg-updated.nw"
func LockExclude(db *sql.DB, lock int) {
	var mode string
	switch lock {
	case 0:
		mode = "Unlock"
	case 1:
		mode = "Lock"
	default:
		Logging(LOG_ERROR, "lockexclude", fmt.Sprintf("The lock id is not supported"))
		return
	}
	for _, name := range config.ExcludePackages {
		Logging(LOG_INFO, "lockexclude", fmt.Sprintf("%s exclude package %s: %s", mode, name, LockPackage(db, lock, name)))
	}
}

// 1054 "pkg-updated.nw"
func RollbackPackage(db *sql.DB, name string) {
	var (
		path   string
		cmdOut string
		err    error
	)

	if path, err = GetPackageInfo(db, "archivepath", name); err != nil {
		Logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Error to get archivepath: %v", err))
	}

	if path == "NULL" {
		Logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Rollback Error: No rollback pkg available for package %s", name))
		return
	}
	Logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Rollback package: %s , pkg: %s", name, path))

	if cmdOut, err = RunCmd("install", path); err != nil {
		Logging(LOG_ERROR, "rollbackpackage", fmt.Sprintf("Install error: %s", err))
	}

	/* Set status for success rollback
	 else {
		UpdatePackage(db, "status", "update-available", "name", name);
	}
	*/
	Logging(LOG_DEBUG, "rollbackpackage", fmt.Sprintf("Output: %s", string(cmdOut)))
	return
}

// 1398 "pkg-updated.nw"
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
			Logging(LOG_ERROR, "scanenabledservices", fmt.Sprintf("Truncated service db: failed | Error: %s", err))
			return
		}
		Logging(LOG_INFO, "scanenabledservices", "Truncated service db: done")

		absolute := "/usr/local/etc/rc.d/"
		files, _ := ioutil.ReadDir(absolute)

		for _, f := range files {
			if f.IsDir() == false {
				svc := absolute + f.Name()
				pkgorigin, err = RunCmd("which", svc)
				if err == nil {

					pkgname, err = GetPackageInfo(db, "name", chop(pkgorigin), "packages", "origin")

					Logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Pkgname for origin %s: %s", chop(pkgorigin), pkgname))
					Logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Add %s: %s", pkgname, svc))

					ret, cmdout, err = ScanScript(svc, cmdout)
					if ret != 0 {
						Logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Service: %s is enabled", svc))
						enabled = 1
					} else {
						Logging(LOG_DEBUG, "scanenabledservices", fmt.Sprintf("Service: %s is not enabled", svc))
						enabled = 0
					}
					AddService(db, pkgname, svc, enabled)
				}
			}
		}
	}
}

// 1454 "pkg-updated.nw"
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

// 1487 "pkg-updated.nw"
func RestartService(svc string) int {
	var (
		ret int
	)
	cmdOut, err := RunCmd("service_restart", svc)
	if err != nil {
		Logging(LOG_ERROR, "restartservice", fmt.Sprintf("Service: Could not restart: %s | cmdOut: %v", svc, cmdOut))
		ret = 1
	}
	return ret
}

// 1505 "pkg-updated.nw"
func RestartEnabledServices(db *sql.DB) []string {
	var (
		ret       int
		err       error
		pkglist   []string
		tmplist   []string
		svc       string
		enabled   string
		pkgname   string
		pkgstatus string
	)

	svccmd, err := GetAllPackages(db, "svccmd", "services")
	if err != nil {
		Logging(LOG_ERROR, "restartenableservices", fmt.Sprintf("Unable to get service list: %s", err))
	}

	for _, value := range svccmd {
		enabled, err = GetPackageInfo(db, "enabled", value, "services", "svccmd")
		if err != nil {
			Logging(LOG_ERROR, "restartenableservices", fmt.Sprintf("Unable to get service enable status: %s", value))
		}
		if enabled == "1" {
			Logging(LOG_DEBUG, "restartenableservices", fmt.Sprintf("Service is enabled: %s", value))

			pkgname, err = GetPackageInfo(db, "name", value, "services", "svccmd")
			if err != nil {
				Logging(LOG_ERROR, "restartenableservices", fmt.Sprintf("Unable to get pkg name from service: %s", value))
				continue
			}

			pkgstatus, err = GetPackageInfo(db, "status", pkgname, "packages", "name")
			if err != nil {
				Logging(LOG_ERROR, "restartenableservices", fmt.Sprintf("Unable to get update status from service: %s", value))
				continue
			}

			if pkgstatus == "update-available" {
				Logging(LOG_DEBUG, "restartenableservices", fmt.Sprintf("Try to restart service: %s", value))
				ret = RestartService(value)
				if ret != 0 {
					Logging(LOG_DEBUG, "restartenableservices", fmt.Sprintf("Put service on rollback list: %s", value))
					svc, err = GetPackageInfo(db, "name", value, "services", "svccmd")
					tmplist = make([]string, len(pkglist)+1)
					copy(tmplist, pkglist)
					pkglist = tmplist
					pkglist[len(pkglist)-1] = svc
				}
			}
		}
	}
	return pkglist
}

// 939 "pkg-updated.nw"
func GetUpdateList(db *sql.DB) ([]string, error) {
	var list []string
	var name string
	list = make([]string, 0)
	var nlist []string

	rows, err := db.Query("SELECT name FROM packages WHERE status = $1", "update-available")
	if err != nil {
		Logging(LOG_ERROR, "getupdatelist", fmt.Sprint(err))
		return list, fmt.Errorf("%s", err)
	}
	defer rows.Close()

	for rows.Next() {
		nlist = make([]string, len(list)+1)
		copy(nlist, list)
		list = nlist
		err = rows.Scan(&name)
		if err != nil {
			Logging(LOG_ERROR, "getupdatelist", fmt.Sprint(err))
			return list, fmt.Errorf("%s", err)
		}
		list[len(list)-1] = name
	}
	return list, nil
}

// 972 "pkg-updated.nw"
func SavePackages(db *sql.DB) {
	var (
		version string
		origin  string
		path    string
		index   int
	)

	updatelist, err := GetUpdateList(db)
	if err != nil {
		Logging(LOG_ERROR, "savepackages", fmt.Sprintf("GetUpdateList(): %v", err))
	}

	// 991 "pkg-updated.nw"
	for _, pkg := range updatelist {
		version, err = GetPackageInfo(db, "version", pkg)
		path = config.ArchivePath + "/" + pkg + "-" + version + ".txz"
		if _, err := os.Stat(path); err != nil {
			RunCmd("create", config.ArchivePath, pkg)
			index++
		}
	}

	for _, pkg := range updatelist {
		version, err = GetPackageInfo(db, "version", pkg)
		path = config.ArchivePath + "/" + pkg + "-" + version + ".txz"
		if _, err := os.Stat(path); err == nil {
			origin, err = GetPackageInfo(db, "origin", pkg)
			// 1010 "pkg-updated.nw"
			//			if err != nil {
			Logging(LOG_DEBUG, "savepackages", fmt.Sprintf("Archive pkg: %s", pkg))
			UpdatePackage(db, "archivepath", path, "origin", origin)
			//				UpdatePackage(db, origin, "archive", path)
			//			}
		} else {
			Logging(LOG_ERROR, "savepackages", fmt.Sprintf("Could not found rollback package file: %s", path))
		}
	}
}

// 1027 "pkg-updated.nw"
func Upgrade() int {
	var (
		cmdOut string
		err    error
		ret    int
	)
	ret = 0
	cmdOut, err = RunCmd("upgrade")
	Logging(LOG_STDOUT, "upgrade", fmt.Sprintf("Output Upgrade(): %s", cmdOut))
	if err != nil {
		Logging(LOG_ERROR, "upgrade", fmt.Sprintf("Error: ", string(cmdOut), err))
		ret = -1
	}
	return ret
}

// 873 "pkg-updated.nw"
func UpdateRoutine(db *sql.DB) bool {
	var (
		ret               bool
		updates_available bool
	)

	Logging(LOG_EVENT, "updateroutine", "Start update routine")
	SyncPkgDatabases(db)

	updates_available = CheckUpdates(db)

	if updates_available == true {
		Logging(LOG_EVENT, "updateroutine", "Updates available")

		if config.ArchiveEnable {
			Logging(LOG_EVENT, "updateroutine", "Save package")
			SavePackages(db)
		}

		// 1204 "pkg-updated.nw"
		LockExclude(db, 1)
		// 892 "pkg-updated.nw"
		ScanEnabledServices(db)

		if Upgrade() != 0 {
			Logging(LOG_EVENT, "updateroutine", "Pkg Upgrade failed")
			return ret
		}

		if config.RestartDaemons == true {
			svc := RestartEnabledServices(db)
			if len(svc) > 0 {
				Logging(LOG_EVENT, "updateroutine", "Service restart failed")
				if config.DowngradePackageOnFailedRestart == true {
					Logging(LOG_EVENT, "updateroutine", "Starting rollback")

					//iterate through the list of failed restarts
					for _, pkgname := range svc {
						RollbackPackage(db, pkgname)
					}
				}
			}
		}

		// 1207 "pkg-updated.nw"
		LockExclude(db, 0)
		// 914 "pkg-updated.nw"
	} else {
		Logging(LOG_EVENT, "updateroutine", "No updates available")
	}
	/* Second RUN
	Need to verify how it behaves

	if config.ClearSyncDatabaseEnabled == true {
		config.ClearSyncDatabaseEnabled = false
	}
	SyncPkgDatabases(db)
	*/

	ret = true
	return ret
}

// 1252 "pkg-updated.nw"
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
		checktime      bool
		checkday       bool
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
		Logging(LOG_ERROR, "scheduler", fmt.Sprintf("Cannot parse schedule value: %s", err))
	}

	Logging(LOG_DEBUG, "scheduler", fmt.Sprintf("Recurring time: %v:%v", recur.Hour(), recur.Minute()))
	Logging(LOG_DEBUG, "scheduler", fmt.Sprintf("Recurring days: %v", config.RecurDays))

	LastCheck = time.Now()
	RecurUnixStamp = (recur.Hour() * 3600) + (recur.Minute() * 60)

	for {
		TimeNow = time.Now()
		diff = LastCheck.Unix() - TimeNow.Unix()
		Ticker = (TimeNow.Hour() * 3600) + (TimeNow.Minute() * 60) + (TimeNow.Second())
		LastCheck = TimeNow

		// Check Recurring Day !
		checkday = false
		if len(config.RecurDays) > 0 {
			for _, day := range config.RecurDays {
				if day == int(TimeNow.Weekday()) {
					checkday = true
				}
			}
		} else {
			checkday = true
		}

		// Detect timejump
		if (diff <= (0 - int64(sleeptimer) - int64(tolerance))) || (diff > (int64(sleeptimer) + int64(tolerance))) {
			Logging(LOG_INFO, "scheduler", "Time jump detected")
			Logging(LOG_DEBUG, "scheduler", "Set AlreadyRun to false")
			Logging(LOG_DEBUG, "scheduler", fmt.Sprintf("Diff (<=-70,>70 ?): %d", diff))
			AlreadyRun = false
		}

		// Only run if time has been arrived

		if (AlreadyRun == false) && ((TimeNow.Hour() > recur.Hour()) || ((TimeNow.Hour() == recur.Hour()) && (TimeNow.Minute() >= recur.Minute()))) {

			// If strict time is enabled, check the 5 minute tolerance
			checktime = false
			if config.StrictRecurTime == true {
				diff2 = Ticker - RecurUnixStamp
				if (diff2 > 0) && (diff2 < 300) {
					checktime = true
				}
			} else {
				checktime = true
			}

			if (checktime == true) && (checkday == true) {
				Logging(LOG_EVENT, "scheduler", fmt.Sprintf("Scheduled Time reached, start job"))
				if UpdateRoutine(db) == true {
					AlreadyRun = true
				}
			}
		}

		if (Ticker >= (86400 - sleeptimer - tolerance)) && (AlreadyRun == true) {
			Logging(LOG_DEBUG, "scheduler", fmt.Sprintf("New day detect, set AlreadyRunn to false"))
			AlreadyRun = false
		}

		Logging(LOG_DEBUG, "scheduler", fmt.Sprintf("RecurTime: %v:%v | Now: %v:%v | Weekday: %v | Weekday Match: %v | AlreadyRun: %v", recur.Hour(), recur.Minute(), TimeNow.Hour(), TimeNow.Minute(), int(TimeNow.Weekday()), checkday, AlreadyRun))

		RunCmd("sleep", strconv.Itoa(sleeptimer))
	}
}

// 636 "pkg-updated.nw"
func main() {

	// 1808 "pkg-updated.nw"
	config.Param_Help = flag.Bool("help", false, "Show help page and exit")
	config.Param_Version = flag.Bool("version", false, "Show version and exit")
	config.Param_ConfigFile = flag.String("config", "", "Path to alternativ config file")
	config.Param_DebugMode = flag.Bool("debug", false, "Run in debug mode")
	config.Param_RunOnce = flag.Bool("runonce", false, "Run only once, disable scheduler")

	/* need to find a way to overwrite config file settings with parameters, but only if parameter is set !
	   config.Param_CreateReport = flag.Bool("reporting", true, "Create and use a report db for all events");
	   config.Param_ClearSyncDatabaseEnabled = flag.Bool("cleardbonstart", false, "If pkg-updated db is already exists, truncate all informations before sync");
	   config.Param_DoFreebsdUpdate = flag.Bool("enableosupdate", false, "Enable update of OS too");
	   config.Param_RestartDaemons = flag.Bool("restartdaemons", true, "Restart enable services if an update affecte");
	   config.Param_DowngradePackageOnFailedRestart = flag.Bool("enablerollback", true, "Rollback package if service restart failed, require -enablearchive true");
	   config.Param_UseSudo = flag.Bool("sudo", false, "Use sudo for all commands");
	   config.Param_ArchiveEnable = flag.Bool("enablearchive", true, "Create a backup package before upgrade, required for rollback");
	   config.Param_ArchivePath = flag.String("archivepath", "", "In which directory should the pkg backups stored");
	   config.Param_PkgDatabaseFile = flag.String("pkgdbfile", "", "The local pkg database file");
	   config.Param_DatabaseFile = flag.String("dbfile", "", "The pkg-updated database file");
	   config.Param_ReportDatabaseFile = flag.String("reportdbfile", "", "The report database file");
	*/

	flag.Parse()

	// 653 "pkg-updated.nw"
	Check()

	Logging(LOG_EVENT, "main", fmt.Sprintf("Started pkg-updated %d.%d.%d", MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION))

	var ret int
	db := OpenDB(config.DatabaseFile)
	defer db.Close()
	ret = FileExists(config.DatabaseFile)
	Logging(LOG_INFO, "main", fmt.Sprintf("Check packages Database %s: %d", config.DatabaseFile, ret))

	if ret != 0 {
		ret = CreateDatabase(db, 0)
		if ret != 0 {
			Logging(LOG_ERROR, "main", "Create packages database: failed")
			os.Exit(2)
		} else {
			Logging(LOG_INFO, "main", "Create packages database: done")
		}
	}

	ret = FileExists(config.ReportDatabaseFile)
	Logging(LOG_INFO, "main", fmt.Sprintf("Check report Database %s: %d", config.ReportDatabaseFile, ret))

	if ret != 0 {
		reportdb := OpenDB(config.ReportDatabaseFile)
		ret = CreateDatabase(reportdb, 1)
		if ret != 0 {
			Logging(LOG_ERROR, "main", "Create report database: failed")
			reportdb.Close()
			os.Exit(2)
		} else {
			Logging(LOG_INFO, "main", "Create report database: done")
		}
		reportdb.Close()
	}

	// 695 "pkg-updated.nw"
	if *config.Param_RunOnce != true {
		go Scheduler(db)
		for {
			RunCmd("sleep", "300")
		}
	} else {
		UpdateRoutine(db)
	}
	// 1212 "pkg-updated.nw"
}
