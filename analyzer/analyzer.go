package analyzer

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/aleksaan/diskusage/config"
	"github.com/aleksaan/diskusage/files"
)

//pairs of key and scale power x, when 1024^x is scale of size
var sizeUnits = map[string]float64{
	"b":  1,
	"Kb": 2,
	"Mb": 3,
	"Gb": 4,
	"Tb": 5,
	"Pb": 6,
}

var sortValues = map[string]float64{
	"name_asc":  1,
	"size_desc": 2,
}

var sortedKeysSizeUnits = []string{"b", "Kb", "Mb", "Gb", "Tb", "Pb"}

//Files -
var Files = &files.TFiles{}

var basePath string
var units *string

//Run -
func Run(cfg *config.Config) {
	basePath = addPathSeparator(*cfg.Analyzer.Path)
	units = cfg.Printer.Units
	scanDir(basePath, 1)
}

//-----------------------------------------------------------------------------------------

//ScanDir - scan directory and return its size
func scanDir(path string, depth int) int64 {
	//read content of folder
	osfiles, _ := ioutil.ReadDir(path)

	var dirsize int64

	//calc total size throught folder content
	for _, osfile := range osfiles {

		file := scanFile(path, osfile.Name(), depth)
		if file.IsDir {
			newpath := addPathSeparator(path + osfile.Name())
			file.Size = scanDir(newpath, depth+1)
		}

		setAdaptedFileSize(file, units)
		dirsize += file.Size
		*Files = append(*Files, *file)
	}

	return dirsize
}

func setAdaptedFileSize(file *files.TFile, units *string) {
	file.AdaptedSize, file.AdaptedUnit = GetAdaptedSize(file.Size, units)
}

//-----------------------------------------------------------------------------------------

//ScanFile - scan dir/file parameters
func scanFile(path string, name string, depth int) *files.TFile {
	f := &files.TFile{}
	f.Name = name
	f.RelativePath = path[len(basePath):]
	f.Depth = depth

	//if file or folder is not accessible then return nil
	pathName := cleanPath(&path, false) + name

	//dirstat, _ := os.Stat(pathName)
	dir, err := os.Lstat(pathName)

	if err != nil {
		f.IsNotAccessible = true
		f.IsNotAccessibleMessage = err.Error()
		return f
	}

	linkdir, linkerr := filepath.EvalSymlinks(pathName)
	if linkerr != nil {
		f.IsNotAccessible = true
		f.IsNotAccessibleMessage = linkerr.Error()
		return f
	}

	if strings.ToLower(linkdir) != strings.ToLower(pathName) {
		f.IsLink = true
		f.LinkedDirPath = linkdir
	}

	f.IsDir = dir.IsDir()

	if !dir.IsDir() {
		f.Size = dir.Size()
	}

	return f
}

//GetAdaptedSize - get file size adapted to InputArgs.FixUnit units or to a flexible useful units
func GetAdaptedSize(sizeB int64, units *string) (float64, string) {

	var size = float64(sizeB)
	var unit string
	var power float64

	if len(*units) > 0 {
		unit = *units
		power = sizeUnits[*units]
	} else {
		for _, unit = range sortedKeysSizeUnits {
			power = sizeUnits[unit]
			if size < math.Pow(1024, power) {
				break
			}
		}
	}
	return (size / math.Pow(1024, power-1)), unit
}

//CleanPath - get absolute path like C:\temp\
func cleanPath(path *string, isrelativeclean bool) string {
	if isrelativeclean {
		abspath, _ := filepath.Abs(*path)
		return addPathSeparator(filepath.Clean(abspath))
	}

	return addPathSeparator(filepath.Clean(*path))
}

//AddPathSeparator - add os path separator to string
func addPathSeparator(path string) string {
	cleanPath := filepath.Clean(path)
	lastSymbol := cleanPath[(len(cleanPath) - 1):]
	if lastSymbol == string(os.PathSeparator) {
		return filepath.Clean(path)
	}
	return filepath.Clean(path) + string(os.PathSeparator)
}
