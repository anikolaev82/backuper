package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileAttr struct {
	Path         string
	FileName     string
	Ext          string
	Size         int64
	LastModified time.Time
	IsDir        bool
	Hash         []byte
	CrDate       time.Time
	Data         string
}

func toJson(listFile *[]FileAttr, target string) bool {
	if len(*listFile) == 0 {
		log.Fatalf("List files are empty")
		return false
	}
	var f *os.File
	var err error

	if _, err = os.Stat(target); errors.Is(err, os.ErrNotExist) {
		fmt.Println("file not exists")
		f, err = os.Create(target)
	} else if err == nil {
		fmt.Println("file exists")
		in := bufio.NewReader(os.Stdin)
		f, err = os.Open(target)
		exists := true
		for exists == true {
			fmt.Printf("File %s is exists. Rewrite? (Y or N) ", f.Name())
			response, _, _ := in.ReadLine()
			cmd := string(response)

			switch strings.Trim(strings.ToUpper(cmd), " ") {
			case "N":
				log.Fatalf("File is exists. Processing stopped")
				return false
			case "Y":
				f, err = os.Create(target)
				exists = false
			default:
				continue
			}
		}
	}

	defer f.Close()
	if err != nil {
		panic(fmt.Sprintf("Error. File %s not created", *out))
	}

	ret, err := json.Marshal(listFile)
	if err != nil {
		panic("Error. Serializing is not success")
	}

	f.Write(ret)
	fa, err := f.Stat()
	fmt.Printf("Created file: %s, size: %v mb \n", f.Name(), fa.Size()/1024/1024)

	return true
}

func extract(src string, dest string) {
	file, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	data, _ := ioutil.ReadAll(file)

	var listFileAttr []FileAttr
	_ = json.Unmarshal(data, &listFileAttr)

	for _, item := range listFileAttr {
		f, err := os.Create(filepath.Join(dest, item.FileName))

		if err != nil {
			f.Close()
			panic(err)
		}
		bytes, err := base64.StdEncoding.DecodeString(item.Data)
		if err != nil {
			f.Close()
			panic(err)
		}
		hash := sha256.New()
		hash.Write(bytes)

		if string(hash.Sum(nil)) == string(item.Hash) {
			var _, _ = f.Write(bytes)
		} else {
			fmt.Printf("Hash file %s not compare\n", item.FileName)
		}
		f.Close()
	}
}

var source = flag.String("s", "description", ".")
var out = flag.String("out", "description", "list_files.json")

func main() {
	var listFile []FileAttr
	flag.Parse()

	fs, err := os.Stat(*source)
	if err != nil {
		panic(fmt.Sprintf("Source dont`t valid: %s", err))
	}

	if fs.IsDir() {
		fmt.Println("Reading files from: ", *source)
		err := filepath.Walk(*source, func(src string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			f, err := os.Open(src)
			if err != nil {
				f.Close()
				panic(err)
			}
			reader := bufio.NewReader(f)
			data, _ := ioutil.ReadAll(reader)
			f.Close()

			hash := sha256.New()
			hash.Write(data)

			fa := FileAttr{
				filepath.Dir(src),
				fi.Name(),
				filepath.Ext(fi.Name()),
				int64(fi.Size()),
				fi.ModTime(),
				fi.IsDir(),
				hash.Sum(nil),
				time.Now(),
				base64.StdEncoding.EncodeToString(data)}

			listFile = append(listFile, fa)
			return nil
		})
		if err != nil {
			fmt.Println("Error for traversal: ", err)
		}
		fmt.Printf("Founded %v files \n", len(listFile))
	}

	if out != nil {

		if _, err := os.Stat(*out); err == nil {
			fmt.Println("toJson")
			toJson(&listFile, *out)
		} else {
			fmt.Println("extract")
			extract(*source, *out)
		}
	}
}
