package eternityFS

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type HashTable struct{}

type FileIndexEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type efsOpts struct {
	Dir     string   `json:"path"`
	FileDir string   `json:"filepath"`
	Peers   []string `json:"peers"`
}

type EternityFS struct {
	Opts    efsOpts           `json:"opts"`
	FileMap map[string]string `json:"filemap"`
}

func makeConfig(dir string) EternityFS {
	fmt.Println(dir)
	defaultOpts := &efsOpts{
		Dir:     dir,
		FileDir: dir + "/files",
		Peers:   make([]string, 0),
	}
	defaultConfig := &EternityFS{
		Opts:    *defaultOpts,
		FileMap: make(map[string]string),
	}
	file, err := json.Marshal(defaultConfig)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(dir+"/config.json", file, os.ModePerm)
	return *defaultConfig
}
func InitEFS(dir string) (EternityFS, error) {
	dirExists := false
	fileDir := dir + "/files"
	// check if the directory exists
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		dirExists = true
	}

	// if no directory exists, create it and populate the config and filepath
	if !dirExists {
		os.Mkdir(dir, 0777)
		os.Mkdir(dir+"/files", os.ModePerm)

		defaultConfig := makeConfig(dir)
		return defaultConfig, nil
	} else {
		// load config if it does exist
		efs := EternityFS{}
		configFile := make([]byte, 0)
		if _, err := os.Stat(dir + "/config.json"); errors.Is(err, os.ErrNotExist) {
			efs = makeConfig(dir)
		} else {
			configFile, err = ioutil.ReadFile(dir + "/config.json")
			if err != nil {
				return EternityFS{}, err
			}

		}

		json.Unmarshal(configFile, &efs)
		configFilePath := efs.Opts.FileDir
		if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
			os.Mkdir(fileDir, 0664)
		}
		efs.IndexFiles(fileDir)
		return efs, nil
	}
}

func (efs EternityFS) SaveConfig() error {
	file, err := json.Marshal(efs)
	if err != nil {
		return err
	}
	os.Remove(efs.Opts.Dir + "/config.json")
	ioutil.WriteFile(efs.Opts.Dir+"/config.json", file, os.ModePerm)
	return nil
}

type FileNotFoundError struct{}

func (e *FileNotFoundError) Error() string {
	return "file with that hash not found"
}

func (efs EternityFS) GetFile(hash string) ([]byte, error) {
	filepath, ok := efs.FileMap[hash]
	if !ok {
		// file does not exist
		return make([]byte, 0), &FileNotFoundError{}
	}
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return make([]byte, 0), err
	}
	return file, nil
}

func (efs EternityFS) Store(file []byte) (string, error) {
	r := bytes.NewReader(file)
	h := sha256.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	fileHash := hex.EncodeToString(h.Sum(nil))
	filepath := efs.Opts.Dir + "/" + string(fileHash)
	println("storing file with hash: ", string(fileHash), " at ", efs.Opts.FileDir+"/"+string(fileHash))
	err := ioutil.WriteFile(efs.Opts.FileDir+"/"+string(fileHash), file, os.ModePerm)
	if err != nil {
		return "", err
	}
	efs.FileMap[string(fileHash)] = filepath

	efs.SaveConfig()

	return fileHash, nil
}

func checkFileHash(hash string, path string) (bool, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	r := bytes.NewReader(file)
	h := sha256.New()

	if _, err := io.Copy(h, r); err != nil {
		return false, err
	}
	tempHash := hex.EncodeToString(h.Sum(nil))
	if hash != tempHash {
		return false, nil
	}

	return true, nil
}

func (efs EternityFS) IndexFiles(dir string) error {
	// look for a hashfile

	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for hash, path := range efs.FileMap {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			delete(efs.FileMap, hash)
		} else {
			val, err := checkFileHash(hash, path)
			if err != nil {
				return err
			}

			if !val {
				delete(efs.FileMap, hash)
				os.Remove(path)
				efs.SaveConfig()
			}

		}
		if err != nil {
			return err
		}
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		} else {
			path := dir + "/" + item.Name()
			println("file path: ", path)

			val, err := checkFileHash(item.Name(), path)

			if err != nil {
				return err
			}

			if !val {
				// hashes do not match
				os.Remove(path)

				// remove from file hash
				delete(efs.FileMap, item.Name())
				efs.SaveConfig()
			} else {
				println("hash matches file")
				// add file to hash map if its name and hash match
				if _, ok := efs.FileMap[item.Name()]; !ok {
					efs.FileMap[item.Name()] = path
					efs.SaveConfig()
				}
			}
		}
	}

	return nil
}
