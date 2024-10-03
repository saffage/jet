package ast

import "github.com/saffage/jet/config"

type File struct {
	Ast *Stmts
	ID  config.FileID
}
