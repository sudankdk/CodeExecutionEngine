package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)


type CodeFiles struct {
	Dir       string
	CodePath  string
	StdinPath string
	Ext       string
}


func Save(code ,stdin,ext string)(*CodeFiles,error){
	base:= filepath.Join(os.TempDir(),"cee")
	dir := filepath.Join(base, uuid.New().String())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	codePath := filepath.Join(dir, "main"+ext)
	stdinPath := filepath.Join(dir, "stdin.txt")
	fmt.Println(codePath,stdinPath,code,stdin)
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	if err := os.WriteFile(stdinPath, []byte(stdin), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	return &CodeFiles{
		Dir:       dir,
		CodePath:  codePath,
		StdinPath: stdinPath,
		Ext:       ext,
	}, nil
}


func GetFileExt(lang string) string {
	switch lang {
	case "python":
		return "py"
	case "go":
		return "go"
	case "cpp":
		return "cpp"
	default:
		return "txt"
	}
}