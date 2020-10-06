package fiks

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	filePDF  = ".pdf"
	fileJSON = ".json"
)

// FileHandler - handles storage of files
type FileHandler struct {

	// root path of files
	RootPath string `yaml:"rootPath"`

	// folder for pdfs
	PDFFolder string `yaml:"pdfFolder"`

	// (hidden) backup folder for pdfs (in case pdfs deleted accidentally)
	PDFBackupFolder string `yaml:"pdfBackupFolder"`

	// (hidden) folder for json version of data
	JSONFolder string `yaml:"jsonFolder"`
}

// Init - initializes the file handler. Create destination folders.
func (f *FileHandler) Init() error {

	// consider performing as singleton at need.
	return createFoldersIfNotExists(
		path.Join(f.RootPath, f.JSONFolder),
		path.Join(f.RootPath, f.PDFFolder),
		path.Join(f.RootPath, f.PDFBackupFolder),
	)
}

// asReadCloser - helper method to get readcloser
func asReadCloser(data []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(data))
}

// writeContentToFile - writes message contents to file.
// NOTE: Currently only supports data in body, not reference to data in 'dokumentlager'
// TODO: Support data reference.
func (f *FileHandler) writeContentToFile(data []byte, messageID string) error {
	zipData, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	//loop through files
	for _, zipFile := range zipData.File {
		rc, err := zipFile.Open()
		if err != nil {
			return err
		}

		//save as respective file type
		if strings.HasSuffix(zipFile.Name, filePDF) {
			err = f.savePDF(rc, messageID)
		} else if strings.HasSuffix(zipFile.Name, fileJSON) {
			err = f.saveJSON(rc, messageID)
		}
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileHandler) savePDF(rc io.ReadCloser, fileName string) error {

	// store in tmp variable for saving to two files.
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}

	file := path.Join(f.RootPath, f.PDFFolder, fileName+filePDF)
	err = writeToFile(asReadCloser(data), file)
	if err != nil {
		return err
	}

	fileBackup := path.Join(f.RootPath, f.PDFBackupFolder, fileName+filePDF)
	return writeToFile(asReadCloser(data), fileBackup)
}

func (f *FileHandler) saveJSON(rc io.ReadCloser, fileName string) error {
	file := path.Join(f.RootPath, f.JSONFolder, fileName+fileJSON)
	return writeToFile(rc, file)
}

// createFoldersIfNotExists - creates folders on file system if they dont exist
func createFoldersIfNotExists(folderPaths ...string) (err error) {
	for _, folderPath := range folderPaths {
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			return
		}
	}
	return
}

func writeToFile(rc io.ReadCloser, fileName string) error {
	// TODO: Consider auto-deleting in case of err.
	fi, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fi.Close()

	_, err = io.Copy(fi, rc)
	return err
}
