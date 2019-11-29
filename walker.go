package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/spikebike/Backups-Done-Right/bdrsql"
	"github.com/spikebike/Backups-Done-Right/bdrupload"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	configFile = flag.String("config", "etc/config.cfg", "Defines where to load configuration from")
	newDB      = flag.Bool("new-db", false, "true = creates a new database | false = use existing database")
	debug_flag = flag.Bool("debug", false, "activates debug mode")

	upchan   = make(chan *bdrupload.Upchan_t, 100)
	downchan = make(chan *bdrupload.Downchan_t, 100)
	done     = make(chan int64)

	debug bool
)

type clientInfo struct {
   PrivateKey   string
   PublicKey    string
   BackupDirs   []string
   ExcludeDirs  []string
   Threads      int
   SqlFile      string
   QueueBlobDir string
}


type ByteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

func checkPath(dirArray []string, excludeArray []string, dir string) bool {
	for _, j := range excludeArray {
		if strings.Contains(dir, j) {
			return true
		}
	}
	for _, i := range dirArray {
		if i == dir {
			return true
		}
	}
	return false
}

func backupDir(db *sql.DB, dirList []string, excludeList []string, dataBaseName string) error {
	var i int
	var fileC int64
	var backupFileC int64
	var dirC int64
	var dFile int64
	var dDir int64
	fileC = 0
	dirC = 0
	backupFileC = 0
	dFile = 0
	dDir = 0
	start := time.Now().Unix()
	i = 0
	for _, dirname := range dirList {
		// get dirID of dirname, even if it needs inserted.
		log.Printf("working on %s\n",dirname)
		dirID, err := bdrsql.GetSQLID(db, "dirs", "path", dirname)
		// get a map for filename -> modified time
		SQLmap := bdrsql.GetSQLFiles(db, dirID)
		if debug == true {
			fmt.Printf("scanning dir %s ", dirname)
		}
		d, err := os.Open(dirname)
		if err != nil {
			log.Printf("failed to open %s error : %s", dirname, err)
			os.Exit(1)
		}
		fi, err := d.Readdir(-1)
		if err != nil {
			log.Printf("directory %s failed with error %s", dirname, err)
		}
		Fmap := map[string]int64{}
		// Iterate over the entire directory
		dFile = 0
		dDir = 0
		for _, f := range fi {
			if !f.IsDir() {
				fileC++ //track files per backup
				dFile++ //trace files per directory
				// and it's been modified since last backup
				if f.ModTime().Unix() <= SQLmap[f.Name()] {
					// log.Printf("NO backup needed for %s \n",f.Name())
					Fmap[f.Name()] = f.ModTime().Unix()
				} else {
					log.Printf("backup needed for %s \n",f.Name())
					backupFileC++
					bdrsql.InsertSQLFile(db, f, dirID)
				}
			} else { // is directory
				dirC++ //track directories per backup
				dDir++ //track subdirs per directory
				fullpath := filepath.Join(dirname, f.Name())

				if !checkPath(dirList, excludeList, fullpath) {
					dirList = append(dirList, fullpath)
				}
			}
		}
		// All files that we've seen, set last_seen
		t1 := time.Now().UnixNano()
		bdrsql.SetSQLSeen(db, Fmap, dirID)
		if debug == true {
			t2 := time.Now().UnixNano()
			fmt.Printf("files=%d dirs=%d duration=%dms\n", dFile, dDir, (t2-t1)/1000000)
		}
		i++
	}
	// if we have not seen the files since start it must have been deleted.
	bdrsql.SetSQLDeleted(db, start)

	log.Printf("scanned %d files and %d directories\n", fileC, dirC)
	log.Printf("%d files scheduled for backup\n", backupFileC)

	return nil
}

func main() {
	var bytes int64
	var bytesDone int64

	flag.Parse()
	debug = *debug_flag

	log.Printf("loading config file from %s\n", *configFile)

	viper.SetConfigFile("etc/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	var C clientInfo
	prod := viper.Sub("client")
	err := prod.Unmarshal(&C)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	runtime.GOMAXPROCS(C.Threads)

	os.Mkdir(C.QueueBlobDir+"/tmp", 0700)
	os.Mkdir(C.QueueBlobDir+"/blob", 0700)

	db, err := bdrsql.Init_db(C.SqlFile, *newDB, debug)
	if err != nil {
		log.Printf("could not open %s, error: %s", C.SqlFile, err)
	} else {
		log.Printf("opened database %v\n", C.SqlFile)
	}

	err = bdrsql.CreateClientTables(db)
	if err != nil && debug == true {
		log.Printf("couldn't create tables: %s", err)
	} else {
		log.Printf("created tables\n")
	}
   fmt.Println(C.BackupDirs)
	log.Printf("backing up these directories: %s\n", C.BackupDirs)
	log.Printf("start walking...")
	t0 := time.Now()
	err = backupDir(db, C.BackupDirs, C.ExcludeDirs, C.SqlFile)
	t1 := time.Now()
	duration := t1.Sub(t0)
	if err != nil {
		log.Printf("walking didn't finished successfully. Error: %s", err)
	} else {
		log.Printf("walking successfully finished")
	}
	log.Printf("walking took: %v\n", duration)

	// shutdown database, make a copy, open it, backup copy of db
	// db, _ = bdrsql.BackupDB(db,dataBaseName)
	// launch server to receive uploads
	tn0 := time.Now().UnixNano()
	for i := 0; i < C.Threads; i++ {
		go bdrupload.Uploader(upchan, done, debug, C.QueueBlobDir)
	}
	log.Printf("started %d uploaders\n", C.Threads)
	// send all files to be uploaded to server.

	log.Printf("started sending files to uploaders...\n")
	bdrsql.SQLUpload(db, upchan)
	bytesDone = 0
	bytes = 0
	for i := 0; i < C.Threads; i++ {
		bytes = <-done
		bytesDone += bytes
	}
	tn1 := time.Now().UnixNano()
	if debug == true {
		seconds := float64(tn1-tn0) / 1000000000
		log.Printf("%d threads %s %s/sec in %4.2f seconds\n", C.Threads, ByteSize(float64(bytesDone)), ByteSize(float64(bytesDone)/seconds), seconds)
	}
	log.Printf("uploading successfully finished\n")
}
