package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/exp/slices"
)

type PackageLockFile struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	LockfileVersion  int                    `json:"lockfileVersion"`
	Requires         bool                   `json:"requires"`
	Packages         map[string]*Package    `json:"packages"`
	RootDependencies map[string]*Dependency `json:"dependencies"`
}

type Package struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Requires        map[string]string `json:"requires"`
}

type Dependency struct {
	Version         string                       `json:"version"`
	Dependencies    map[string]*NestedDependency `json:"dependencies"`
	DevDependencies map[string]string            `json:"devDependencies"`
	Requires        map[string]string            `json:"requires"`
}

type NestedDependency struct {
	Version      string                       `json:"version"`
	Dependencies map[string]*NestedDependency `json:"dependencies"`
	Requires     map[string]string            `json:"requires"`
}

type Result struct {
	Path    string
	Version string
}

func getRequiresNpmPackage(Requires map[string]string, packageLockFileStruct PackageLockFile, checked []string, result []Result) ([]string, []Result) {
	for requireKey := range Requires {
		if !slices.Contains(checked, requireKey) {
			checked, result = getNpmPackage(requireKey, packageLockFileStruct, checked, result)
		}
	}
	return checked, result
}

func getNestedDependencyNpmPackage(Dependency map[string]*NestedDependency, packageLockFileStruct PackageLockFile, checked []string, result []Result) ([]string, []Result) {
	flag := false
	for dependencyKey, dependencyValue := range Dependency {
		if !slices.Contains(checked, dependencyKey) {
			checked = append(checked, dependencyKey)
			result = append(result, Result{dependencyKey, dependencyValue.Version})
		} else {
			// key is present in checked, check key, path in result
			for _, j := range result {
				if j.Path == dependencyKey && j.Version == dependencyValue.Version {
					flag = true
					break
				}
			}
			if !flag {
				result = append(result, Result{dependencyKey, dependencyValue.Version})
			}
			flag = false
		}

		// check if requires exist in nested Dependency
		if Dependency[dependencyKey].Requires != nil {
			checked, result = getRequiresNpmPackage(Dependency[dependencyKey].Requires, packageLockFileStruct, checked, result)
		}
		// check if second nested Dependency exist in nested Dependency
		if Dependency[dependencyKey].Dependencies != nil {
			checked, result = getNestedDependencyNpmPackage(Dependency[dependencyKey].Dependencies, packageLockFileStruct, checked, result)
		}
	}
	return checked, result
}

func getNpmPackage(key string, packageLockFileStruct PackageLockFile, checked []string, result []Result) ([]string, []Result) {
	if !slices.Contains(checked, key) {
		_, present := packageLockFileStruct.RootDependencies[key]
		if present {
			ver := packageLockFileStruct.RootDependencies[key].Version
			checked = append(checked, key)
			result = append(result, Result{key, ver})

			// check if requires exist
			if packageLockFileStruct.RootDependencies[key].Requires != nil {
				checked, result = getRequiresNpmPackage(packageLockFileStruct.RootDependencies[key].Requires, packageLockFileStruct, checked, result)
			}

			// check if nested dependency exist
			if packageLockFileStruct.RootDependencies[key].Dependencies != nil {
				checked, result = getNestedDependencyNpmPackage(packageLockFileStruct.RootDependencies[key].Dependencies, packageLockFileStruct, checked, result)
			}
		}
	}
	return checked, result
}

func ParsePackageLockFile(Data []byte) (*PackageLockFile, error) {
	var packageLockFileStruct PackageLockFile

	err := json.Unmarshal(Data, &packageLockFileStruct)
	if err != nil {
		fmt.Println("unmarshal error", err)
		return nil, err
	}

	var checked []string
	var result []Result
	for key := range packageLockFileStruct.Packages[""].Dependencies {
		//Add any specific conditions for Dependencies
		checked, result = getNpmPackage(key, packageLockFileStruct, checked, result)
	}
	for key := range packageLockFileStruct.Packages[""].DevDependencies {
		//Add any specific conditions for Dev Dependencies
		checked, result = getNpmPackage(key, packageLockFileStruct, checked, result)
	}
	return &packageLockFileStruct, nil
}

/*
ReadFileBytes - read the file from the given path.  If you cannot read print an error and exit.
*/
func ReadFileBytes(path string) ([]byte, error) {
	if path == "" {
		return []byte{}, fmt.Errorf("path cannot be an empty value")
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to resolve path '%s' : %s", path, err.Error())
	}
	if !fileInfo.Mode().IsRegular() {
		return []byte{}, fmt.Errorf("ERROR file '%s' is not a regular file", fileInfo.Name())
	}
	byteContent, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading file '%s' : %s", fileInfo.Name(), err.Error())
	}
	return byteContent, nil
}

func main() {
	packageLockFile := "test-app/package-lock.json"
	bytes, err := ReadFileBytes(packageLockFile)
	if err != nil {
		fmt.Print(err)
	}
	parsedLockFile, _ := ParsePackageLockFile([]byte(bytes))
	fmt.Println(parsedLockFile)
}
