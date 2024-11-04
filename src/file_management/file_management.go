package file_management

import (
	"archive/zip"
	"cli_window_helper/src/app_log"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var rutaDelHost string = "C:\\Windows\\System32\\drivers\\etc\\hosts"

func copyFile(rutaDelArhivoOriginal string, rutaDelArchivoNuevo string) error {
	original, err := os.Open(rutaDelArhivoOriginal)
	if err != nil {
		return err
	}
	defer original.Close()

	// Create new file
	new, err := os.Create(rutaDelArchivoNuevo)
	if err != nil {
		return err
	}
	defer new.Close()

	//This will copy
	_, err = io.Copy(new, original)
	if err != nil {
		return err
	}
	return nil
}
func CopyDirectory(src string, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursive copy for directories
			err = CopyDirectory(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy files
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ZipSource(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
func GuardarHostFile(rutaDelCopiado string) error {
	rutaDelCopiado = path.Join(rutaDelCopiado, fmt.Sprintf("hosts_%d", time.Now().UnixMilli()))
	return copyFile(rutaDelHost, rutaDelCopiado)
}
func GuardarInformeCLI(rutaCli string, rutaDelCopiado string) error {
	rutaDelCopiado = path.Join(rutaDelCopiado, fmt.Sprintf("informe_%d", time.Now().UnixMilli()))
	return copyFile(rutaCli, rutaDelCopiado)
}
func MoverArchivoLog(nuevaRuta string) error {
	oldLocation := fmt.Sprintf("./%s", app_log.GetNombreDelArchivoLog())
	return os.Rename(oldLocation, nuevaRuta)
}
func CrearDirectorio(ruta string) error {
	err := os.Mkdir(ruta, 0777)
	if err != nil {
		if !os.IsExist(err) {
			return os.MkdirAll(ruta, 0777)
		} else {
			return err
		}
	}
	return nil
}

type Directory struct {
	BaseDirectory        string
	DirectSubDirectories []string
	AllSubDirectories    []string
	Files                []string
}

func DirectoryInfo(path string) *Directory {
	return &Directory{BaseDirectory: path}
}

func (dir *Directory) GetDirectories(topDirsOnly bool) []string {
	var subDirectories []string
	var path = dir.BaseDirectory

	filepath.WalkDir(path, func(subPath string, dir fs.DirEntry, err error) error {
		if dir.Type().IsDir() && (!topDirsOnly || IsChildFolderOf(path, subPath)) {
			subDirectories = append(subDirectories, subPath)
		}
		return nil
	})
	SortByName(subDirectories)
	return subDirectories
}

func (dir *Directory) GetFiles(topDirsOnly bool) []string {
	var files []string
	var rootPath = dir.BaseDirectory

	filepath.WalkDir(rootPath, func(subPath string, dir fs.DirEntry, err error) error {
		if dir.Type().IsRegular() && (!topDirsOnly || IsChildFolderOf(rootPath, subPath)) {
			files = append(files, subPath)
		}
		return nil
	})
	SortByName(files)
	return files
}

func IsChildFolderOf(rootDir string, subDir string) bool {
	return len(strings.Split(subDir, "\\")) == len(strings.Split(rootDir, "\\"))+1 ||
		len(strings.Split(subDir, "\\")) == len(strings.Split(rootDir, "/"))+1
}

func GetFather(path string) string {
	var segmented = strings.Split(path, "\\")
	return segmented[len(segmented)-2]
}

func SortByName(slice []string) {
	sort.Slice(slice, func(i, j int) bool { return strings.ToLower(slice[i]) < strings.ToLower(slice[j]) })
}

var routes = [...]string{"SIMF\\MSWS", "SIMF\\RestAPI_SIMF", "SIMF\\RestAPI", "SGLBTR\\MSWS", "SGLBTR\\RestApi_LBTR", "SIMF\\Dashboard"}
var logsRoutes = [...]string{"LOGS_SIMF\\SIMF\\Microservicios", "LOGS_SIMF\\SIMF\\Restapi", "LOGS_SIMF\\SGLBTR\\Microservicios%SGLBTR\\LOGS/MSWS", "LOGS_SIMF\\SGLBTR\\Restapi%SGLBTR\\LOGS\\RESTAPI"}

type ConfigFileType string

const (
	NlogFile ConfigFileType = "nlog.config"
	AppFile  ConfigFileType = "appsettings.json"
)

var ConfigFiles = make(map[string]string)

var LogsFiles = make(map[string]string)

func GetConfigFiles(driver string, fileType ConfigFileType) map[string]string {
	var prefious = fileType
	for _, route := range routes {

		if route == "SIMF\\Dashboard" && fileType == AppFile {
			fileType = "dashboard.exe.config"
		}

		path := driver + route

		if _, err := os.Stat(path); err != nil {
			continue
		}

		var files = DirectoryInfo(path).GetFiles(false)

		for _, file := range files {
			if strings.Contains(strings.ToLower(file), string(fileType)) {
				if _, err := os.Stat(file); err == nil {

					content, _ := ioutil.ReadFile(file)
					ConfigFiles[GetFather(file)] = string(content)
				}
			}
		}
	}
	fileType = prefious
	return ConfigFiles
}

func GetLastLogs(driver string) map[string]string {
	for _, route := range logsRoutes {

		version, path := FindLogsBaseDir(driver, route)
		if version == "" {
			continue
		}

		var latestDir = DirectoryInfo(path).GetDirectories(true)

		if len(latestDir) == 0 {
			continue
		}

		if version == "V1" && strings.Contains(strings.ToLower(route), "sglbtr") ||
			version == "V2" && strings.Contains(strings.ToLower(route), "api") {
			GetBasicLogs(latestDir[len(latestDir)-1])
		} else {
			GetProductsLogs(latestDir[len(latestDir)-1])
		}
	}
	return LogsFiles
}

func GetBasicLogs(latestDir string) {
	for _, level := range DirectoryInfo(latestDir).GetDirectories(true) {
		var lastFiles = DirectoryInfo(level).GetFiles(true)
		var previous = ""

		for i := 1; i <= len(lastFiles); i++ {
			var segmented = strings.Split(lastFiles[len(lastFiles)-i], "\\")
			var fileName = segmented[len(segmented)-1]
			var dirName = level + strings.Split(fileName, "-")[0]
			if strings.Split(fileName, "-")[0] == "2022" {
				dirName = level
			}

			if previous != fileName {
				previous = fileName

				if len(lastFiles) == 0 {
					continue
				}

				content, _ := ioutil.ReadFile(lastFiles[len(lastFiles)-i])
				LogsFiles[GetFoldersUp(dirName, 4)] = string(content)
			}
		}
	}
}

func GetProductsLogs(latestDir string) {
	for _, productDir := range DirectoryInfo(latestDir).GetDirectories(true) {

		for _, microservice := range DirectoryInfo(productDir).GetDirectories(true) {

			for _, level := range DirectoryInfo(microservice).GetDirectories(true) {

				var lastFiles = DirectoryInfo(level).GetFiles(true)
				var dirName = level

				if len(lastFiles) == 0 {
					continue
				}
				content, _ := ioutil.ReadFile(lastFiles[len(lastFiles)-1])
				LogsFiles[GetFoldersUp(dirName, 5)] = string(content)
			}
		}
	}
}

func FindLogsBaseDir(driver string, path string) (string, string) {
	var paths = strings.Split(path, "%")

	if _, err := os.Stat(driver + paths[0]); err == nil {
		return "V2", driver + paths[0]
	} else if len(paths) > 1 {
		if _, err := os.Stat(driver + paths[1]); err == nil {
			return "V1", driver + paths[1]
		} else {
			return "", ""
		}
	} else {
		return "", ""
	}
}

func GetFoldersUp(path string, foldersNum int) string {
	var sections = strings.Split(path, "\\")
	if foldersNum >= len(sections) {
		return path
	}

	var newPath = ""
	for i := foldersNum; i > 0; i-- {
		newPath += sections[len(sections)-i]
		if i-1 > 0 {
			newPath += "_"
		}
	}
	return newPath
}
