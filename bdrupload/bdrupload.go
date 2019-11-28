package bdrupload

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Upchan_t struct {
	Rowid int64
	Path  string
}

type Downchan_t struct {
	Rowid int
	Err   error
}

var (
	commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

	debug bool
)

const bufferSize = 524288

func Uploader(upchan chan *Upchan_t, done chan int64, dbg bool, UpDir string) {
	debug = dbg
	var rCount int
	var size int64
	var totalSize int64

	totalSize = 0
	readBuffer := make([]byte, bufferSize)
	//	writeBuffer := make([]byte, bufferSize)
	ciphertext := make([]byte, bufferSize)
	//	key_text := "32o4908go293hohg98fh40ghaidlkjk1"
	key_text := []byte{0x12, 0x9e, 0x12, 0xfd, 0x1f, 0x9e, 0x2b, 0x31, 0x12, 0x9e, 0x12, 0xfd, 0x1f, 0x9e, 0x2b, 0x31, 0x12, 0x9e, 0x12, 0xfd, 0x1f, 0x9e, 0x2b, 0x31, 0x12, 0x9e, 0x12, 0xfd, 0x1f, 0x9e, 0x2b, 0x31}
	c, err := aes.NewCipher([]byte(key_text))
	if err != nil {
		fmt.Printf("Error: NewCipher(%d bytes) = %s", len(key_text), err)
		os.Exit(-1)
	}
	for f := range upchan {
		if debug == true {
			fmt.Printf("Server: received rowID=%d path=%s\n", f.Rowid, f.Path)
		}
		size = 0

		// open file and create a reader
		file, err := os.Open(f.Path)
		if err != nil {
			log.Fatal(err)
		}
		outF, err := ioutil.TempFile(UpDir+"/tmp", "bdr")
		log.Printf("received file=%s blob=%s\n",file.Name(),outF.Name())

		if dbg {
			fmt.Printf("Opening tmp file %s\n", outF.Name())
		}

		reader := bufio.NewReader(file)
		writer := bufio.NewWriter(outF)

		// for this file create a cipher and new sha256 state
		cfb := cipher.NewCFBEncrypter(c, commonIV)
		h := sha256.New() // h is a hash.Hash
		// time how long to read, encrypt, and checksum a file
		t0 := time.Now().UnixNano()
		for {
			if rCount, err = reader.Read(readBuffer); err != nil {
				break
			}
			size = size + int64(rCount)
			cfb.XORKeyStream(ciphertext[:rCount], readBuffer[:rCount])
			h.Write(ciphertext[:rCount])
			if _, err = writer.Write(ciphertext[:rCount]); err != nil {
				log.Printf("Write failed for %s with %s\n", outF.Name(), err)
			}
		}
		t1 := time.Now().UnixNano()
		file.Close()
		writer.Flush()
		outF.Close()
		blobName := fmt.Sprintf("%s/blob/%x", UpDir, h.Sum(nil))
		if dbg {
			fmt.Printf("newblob = %s\n", blobName)
		}
		os.Rename(outF.Name(), blobName)
		seconds := float64(t1-t0) / 1000000000
		if debug == true {
			fmt.Printf("%x %s %4.2f MB/sec\n", h.Sum(nil), f.Path, float64(size)/(1024*1024*seconds))
		}
		totalSize = totalSize + size
	}
	if debug == true {
		fmt.Print("Server: Channel closed\n")
	}
	done <- totalSize 
}
