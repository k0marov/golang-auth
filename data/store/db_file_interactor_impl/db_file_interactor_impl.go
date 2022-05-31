package db_file_interactor_impl

import (
	"auth/domain/entities"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type sizeGetter interface {
	Size() int64
}

// os.File is ok at implementing this
type DBFile interface {
	io.ReadWriteSeeker
	Stat() (os.FileInfo, error)
	Truncate(size int64) error
}

type DBFileInteractorImpl struct {
	dbFile DBFile
	mu     sync.Mutex
}

func NewDBFileInteractor(dBFile DBFile) *DBFileInteractorImpl {
	return &DBFileInteractorImpl{
		dbFile: dBFile,
	}
}

func (d *DBFileInteractorImpl) ReadUsers() ([]entities.StoredUser, error) {
	_, err := d.dbFile.Seek(0, 0)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while seeking in ReadUsers(): %v", err))
	}

	var users []entities.StoredUser
	err = json.NewDecoder(d.dbFile).Decode(&users)
	if err != nil && err != io.EOF { // EOF is ok, just means an empty user list
		return nil, errors.New(fmt.Sprintf("Error while decoding in ReadUsers(): %v", err))
	}
	return users, nil
}

func (d *DBFileInteractorImpl) WriteUser(newUser entities.StoredUser) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// a hack to improve performance:
	// instead of rewriting the whole file, append newUser to the end of the JSON list
	prepareForAppending(d.dbFile)
	err := json.NewEncoder(d.dbFile).Encode(newUser)
	if err != nil {
		return errors.New(fmt.Sprintf("error while encoding in WriteUser(): %v", err))
	}
	addClosingBracket(d.dbFile)

	return nil
}

func prepareForAppending(dbFile DBFile) {
	fi, _ := dbFile.Stat()
	size := fi.Size()
	if size > 0 {
		dbFile.Truncate(size - 1) // remove the "]"
		dbFile.Seek(0, io.SeekEnd)
		dbFile.Write([]byte(", "))
	} else {
		dbFile.Write([]byte("[")) // add the "["
	}
}

func addClosingBracket(dbFile DBFile) {
	dbFile.Write([]byte("]"))
}
