package bdrsql

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"path/filepath"
	"github.com/spikebike/Backups-Done-Right/bdrupload"
	"syscall"
	"time"
)

var (
	clientSQLs= []string{
		"create table dirs (id INTEGER PRIMARY KEY, mode INT, ino BIGINT, uid INT, gid INT, path varchar(2048), last_seen BIGINT, deleted INT)",
		"create table files (id INTEGER PRIMARY KEY, mode INT, ino BIGINT, dev BIGINT, uid INT, gid INT, size BIGINT, mtime BIGINT, ctime BIGINT, name varchar(254), dirID BIGINT, last_seen BIGINT, deleted INT, do_upload INT, FOREIGN KEY(dirID) REFERENCES dirs(id))",
		"create index ctimeindex on files (ctime)",
		"create index pathindex on dirs (path)",
	}
	serverSQLs = []string{
			"create table peers (id INTEGER PRIMARY KEY, publicKey varchar(1024), last_ip varchar(64)",
	}
	debug bool
)

func Init_db(dataBaseName string, newDB bool, dbg bool) (db *sql.DB, err error) {
	debug = dbg
	if newDB == true {
		// rm dataBaseName*  (all backups)
		fps, _ := filepath.Glob(dataBaseName + "*")
		for _, fp := range fps {
			os.Remove(fp)
		}
	}

	db, err = sql.Open("sqlite3", dataBaseName)
	if err != nil {
		log.Printf("couldn't open database: %s", err)
		os.Exit(1)
	}
	// Allow commits to be buffered, MUCH faster.  
	// debug = true makes database writes synchronous and much slower,
	if debug == false {
		_, err = db.Exec("PRAGMA synchronous=OFF")
		if err != nil {
			log.Printf("%s", err)
		}
	}
	return db, err
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	//log.Printf("dest=%#v\n", dstName)

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

func BackupDB(db *sql.DB, dbname string) (*sql.DB, error) {
	time := time.Now().Unix()
	tmpDBName := fmt.Sprintf("%s.%d", dbname, time)
	log.Printf("backing up %s to %s\n", dbname, tmpDBName)
	db.Close()                  // Insure that ALL writes are completed.
	CopyFile(tmpDBName, dbname) // make a snapshot
	fi, err := os.Stat(tmpDBName)
	directory, _ := filepath.Split(tmpDBName)
	db2, err := Init_db(dbname, false, false)
	if err != nil {
		log.Fatalf("Init_DB failed with: %s\n", err)
	}
	dirID, err := GetSQLID(db2, "dirs", "path", directory)
	InsertSQLFile(db2, fi, dirID)
	return db2, err
}

func CreateClientTables(db *sql.DB) error {
	var err error
	for _, sql := range clientSQLs {
		_, err = db.Exec(sql)
		if err != nil && debug == true {
			log.Printf("%s", err)
		}
	}
	return err
}

func CreateServerTables(db *sql.DB) error {
	var err error
	for _, sql := range serverSQLs {
		_, err = db.Exec(sql)
		if err != nil && debug == true {
			log.Printf("%s", err)
		}
	}
	return err
}

func SQLUpload(db *sql.DB, UpChan chan *bdrupload.Upchan_t) error {
	var rowID int64
	var dirID int64
	var olddirID int64
	var name string
	var dir string

	rows, err := db.Query("select name, id, dirID from files where do_upload = 1")
	if err != nil {
		fmt.Printf("GetSQLToUpload query failed: %s\n", err)
	}
	defer rows.Close()
	olddirID=-1
	for rows.Next() {
		rows.Scan(&name, &rowID, &dirID)
		if dirID != olddirID {
			dir, err = GetDir(db, dirID)
			olddirID = dirID
		}
		fullpath := filepath.Join(dir,name)
		log.Printf("fullpath=%s\n",fullpath)

		// send fullpath and rowID to channel
		recptr:=&bdrupload.Upchan_t{}
		recptr.Rowid=rowID
		recptr.Path=fullpath
		if debug == true {
			log.Printf("sending %s",fullpath)
		}
		UpChan <- recptr
	}
	close(UpChan)
	return nil
}

func GetSQLFiles(db *sql.DB, dirID int64) map[string]int64 {
	var fileMap = map[string]int64{}
	var name string
	var mtime int64

	stmt, err := db.Prepare("select name,mtime from files where dirID=? and deleted=0")
	if err != nil {
		fmt.Printf("GetSQLFiles prepare of select failed: %s\n", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(dirID)
	if err != nil {
		fmt.Printf("GetSQLFiles query failed: %s\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&name, &mtime)
		fileMap[name] = mtime
	}
	return fileMap
}

func SetSQLSeen(db *sql.DB, fmap map[string]int64, dirID int64) {
	now := time.Now().Unix()
	//	tx,_ := db.Begin()
	update := fmt.Sprintf("update files set last_seen=%d where name=? and dirID=%d and deleted=0 and ctime=?", now, dirID)
	stmt, _ := db.Prepare(update)
	defer stmt.Close()
	for i, _ := range fmap {
		//		log.Printf("file = %v dirID=%d\n",i,dirID)
		stmt.Exec(i, fmap[i])
	}
	//	tx.Commit()
}

func SetSQLDeleted(db *sql.DB, now int64) {
	stmt, err := db.Prepare("update files set deleted=1 where last_seen<?")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()

	stmt.Exec(now)
}

func GetDir(db *sql.DB, dirID int64) (string, error) {
		var dirname string
		stmt,_ := db.Prepare("select path from dirs where id=?")
		err := stmt.QueryRow(dirID).Scan(&dirname)
		return dirname, err
}


func GetSQLID(db *sql.DB, tablename string, field string, value string) (int64, error) {
	var dirID int64

	dirID = -1

	query := "select id from " + tablename + " where " + field + "= (?)"
	//	fmt.Printf("query=%s\n", query)
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("GetSQLID: prepare select failed: %s\n", err)
	}
	err = stmt.QueryRow(value).Scan(&dirID)
	if err != nil {
		if debug == true {
			log.Printf("GetSQLID: missing %s, error %s\n", value, err)
		}


		insert := "insert into " + tablename + "(" + field + ") values(?);"
		stmt, err := db.Prepare(insert)
		if err != nil {
			log.Println(err)
			return -1, err
		}
		result, err := stmt.Exec(value)
		dirID, err = result.LastInsertId()
	}
	defer stmt.Close()
	return dirID, err
}

func InsertSQLFile(db *sql.DB, fi os.FileInfo, dirID int64) error {
	now := time.Now().Unix()
	e, _ := fi.Sys().(*syscall.Stat_t)

	stmt, err := db.Prepare("insert into files(name,size,mode,gid,uid,ino,dev,mtime,ctime,last_seen,dirID,deleted,do_upload) values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Printf("InsertSQL prepare: %s\n", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(fi.Name(), e.Size, e.Mode, e.Gid, e.Uid, e.Ino, e.Dev, e.Mtim.Sec, e.Ctim.Sec, now, dirID, 0, 1)
	if err != nil {
		log.Printf("InsertSQL Exec: %s\n", err)
		return err
	}
	return err
}

func GetDBSize(DataBaseName string) int64 {
	fi, err := os.Stat(DataBaseName)
	if err != nil {
		log.Printf("couldn't open file %s with ERROR: %s\n", DataBaseName, err)
	}
	return fi.Size()
}

func sqlServer (db *sql.DB) {
/* this will run as a separate process.  This select statement will ensure
that only a single SQL query happens at once.  This allow multiple processes to safely use the sqlite database */
	for {
		select {
		case sqlMapQ := <- sqlMAPQC // return timestamps for entire dir
		case InsertSQLFile // insert metadata about new file
		case // update all file list for last_seen
      case // handle all files not found, to deleted
		}
	}
}


func main_test() {
	db, _ := Init_db("fsmeta.sql", true, false)
	id, _ := GetSQLID(db, "dirs", "path", "/home/bill/bdr/src/bdrsql")
	fmt.Printf("I found id=%d\n", id)
	d, _ := os.Open(".")
	fi, _ := d.Readdir(-1)
	for _, fi := range fi {
		//		unixStat, _ := fi.Sys().(*syscall.Stat_t)
		//		fmt.Printf("%T %#v\n",&fi,&fi)
		//		fmt.Printf("%T %#v\n\n", unixStat, unixStat)
		InsertSQLFile(db, fi, -1)
		//		fmt.Printf("%T %#v\n\n", unixStat.Ino, unixStat.Ino)
	}
}

