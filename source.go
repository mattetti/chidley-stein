package chidleystein

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
)

type Source interface {
	io.Closer
	NewSource(name string) error
	GetName() string
	GetReader() io.Reader
	copySource() (Source, error)
}

type GenericSource struct {
	name   string
	reader io.Reader
}

type FileSource struct {
	GenericSource
	file *os.File
}

type UrlSource struct {
	GenericSource
}

type StdinSource struct {
	GenericSource
}

//StdInSource impl
func (us *StdinSource) copySource() (Source, error) {
	err := new(InternalError)
	//error.ErrorString = "copySource not supported"
	return nil, err
}

func (us *StdinSource) GetName() string {
	return ""
}

func (us *StdinSource) NewSource(name string) error {
	us.reader = bufio.NewReader(os.Stdin)
	return nil
}

func (us *StdinSource) Close() error {
	return nil
}

func (us *StdinSource) GetReader() io.Reader {
	return us.reader
}

//UrlSource impl
func (us *UrlSource) copySource() (Source, error) {
	copy := new(UrlSource)
	err := copy.NewSource(us.name)
	return copy, err
}

func (us *UrlSource) GetName() string {
	return us.name
}

func (us *UrlSource) NewSource(name string) error {
	us.name = name
	var err error

	res, err := http.Get(name)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatal("ERROR: bad http status code != 200: ", res.StatusCode, "   ", name)
		return nil

	}
	us.reader = res.Body

	return err
}

func (us UrlSource) Close() error {
	closer, ok := us.reader.(io.Closer)
	if ok {
		return closer.Close()
	}

	return nil
}

func (us *UrlSource) GetReader() io.Reader {
	return us.reader
}

// FileSource impl
func (fs *FileSource) copySource() (Source, error) {
	copy := new(FileSource)
	err := copy.NewSource(fs.name)
	return copy, err
}

func (fs *FileSource) GetName() string {
	return fs.name
}

func (fs *FileSource) NewSource(name string) error {
	fs.name = name
	var err error
	fs.reader, fs.file, err = genericReader(name)

	return err
}

func (fs *FileSource) Close() error {
	return fs.file.Close()
}

func (fs FileSource) GetReader() io.Reader {
	return fs.reader
}
