package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sudankdk/ceev2/internal/model"
)


func GetExt(lang string) string{
	switch lang{
	case "python" :
		return "py"
	case "go":
		return "go"
	default:
		return "txt"
	}
}

var BASE_CODE_PATH="./storage/code/"

func getHostFilePath(dirPath ,fileName string) string {
	return filepath.Join(dirPath,fileName)
}

func getContainerFilePath(mountPath,fileName string) string{
	return  filepath.Join(mountPath,fileName)
}

func FileAndCodepath(code model.Code) (string,string){
	dirPath :=BASE_CODE_PATH+GetExt(code.Language)  // yeslai ma paxi env bata config to herer tanxu
	fileName := uuid.New().String()
	codeFileName := fileName+"."+GetExt(code.Language)
	InputFileName:=fileName+".txt"
	return getHostFilePath(dirPath,codeFileName),getHostFilePath(dirPath,InputFileName)
}

func CreateFile(filePath ,base64contetn string) (string,error){
	file,err:=os.Create(filePath)
	if err != nil {
		return "",errors.New("error in creating file")
	}
	data,err:=base64.StdEncoding.DecodeString(base64contetn)
	if err != nil {
				return "", fmt.Errorf("failed to decode the file content: %w", err)
	}
	_,err=file.Write(data)
	if err != nil {
		return "",fmt.Errorf("failed to write the content to the file: %w", err)
	}
	return filepath.Base(filePath),nil

}


func Create(code model.Code) (string,string,error){
	codeFilePath,InputFilePath:=FileAndCodepath(code)
	CodefileName,err:=CreateFile(codeFilePath,code.SourceCode)
	if err != nil {
		return "", "", fmt.Errorf("failed to create the code file: %w", err)
	}
	inputFileName,err:= CreateFile(InputFilePath,code.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("failed to create the input file: %w", err)
	}

	return CodefileName, inputFileName, nil

}