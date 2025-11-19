package executer

import (
	"github.com/sudankdk/ceev2/internal/model"
)

type CodeController struct {
	cc model.CodeContainer
}

func NewCodeController(cc model.CodeContainer) *CodeController {
	return &CodeController{cc: cc}
}

func (c *CodeController) Execute(code *model.Code) (*model.Result, error) {
	
	return c.cc.Execute(code)
}


