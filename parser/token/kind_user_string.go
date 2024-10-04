// Code generated by "stringer -type=Kind -output=kind_user_string.go -linecomment"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Illegal-0]
	_ = x[EOF-1]
	_ = x[Comment-2]
	_ = x[Type-3]
	_ = x[TypeVar-4]
	_ = x[Name-5]
	_ = x[Underscore-6]
	_ = x[Int-7]
	_ = x[Float-8]
	_ = x[String-9]
	_ = x[LParen-10]
	_ = x[RParen-11]
	_ = x[LCurly-12]
	_ = x[RCurly-13]
	_ = x[LBracket-14]
	_ = x[RBracket-15]
	_ = x[Colon-16]
	_ = x[Comma-17]
	_ = x[Eq-18]
	_ = x[EqOp-19]
	_ = x[Bang-20]
	_ = x[NeOp-21]
	_ = x[LtOp-22]
	_ = x[LeOp-23]
	_ = x[GtOp-24]
	_ = x[GeOp-25]
	_ = x[Shl-26]
	_ = x[ShlEq-27]
	_ = x[Shr-28]
	_ = x[ShrEq-29]
	_ = x[Plus-30]
	_ = x[PlusEq-31]
	_ = x[Minus-32]
	_ = x[MinusEq-33]
	_ = x[Asterisk-34]
	_ = x[AsteriskEq-35]
	_ = x[Slash-36]
	_ = x[SlashEq-37]
	_ = x[Percent-38]
	_ = x[PercentEq-39]
	_ = x[Caret-40]
	_ = x[CaretEq-41]
	_ = x[Amp-42]
	_ = x[AmpEq-43]
	_ = x[And-44]
	_ = x[AndEq-45]
	_ = x[Pipe-46]
	_ = x[PipeEq-47]
	_ = x[Or-48]
	_ = x[OrEq-49]
	_ = x[Arrow-50]
	_ = x[FatArrow-51]
	_ = x[Dot-52]
	_ = x[Dot2-53]
	_ = x[KwLet-54]
	_ = x[KwType-55]
	_ = x[KwExtern-56]
	_ = x[KwWith-57]
	_ = x[KwWhen-58]
	_ = x[KwIf-59]
	_ = x[KwElse-60]
	_ = x[KwIn-61]
	_ = x[KwAs-62]
	_ = x[KwDefer-63]
	_ = x[KwBreak-64]
	_ = x[KwReturn-65]
	_ = x[KwContinue-66]
}

const _Kind_user_name = "illegal characterend of filecommenttypetype varnameunderscoreuntyped intuntyped floatuntyped string'('')''{''}''['']'':'','operator '='operator '=='operator '!'operator '!='operator '<'operator '<='operator '>'operator '>='operator '<<'operator '<<='operator '>>'operator '>>='operator '+'operator '+='operator '-'operator '-='operator '*'operator '*='operator '/'operator '/='operator '%'operator '%='operator '^'operator '^='operator '&'operator '&='operator '&&'operator '&&='operator '|'operator '|='operator '||'operator '||='operator '->'operator '=>'operator '.'operator '..'keyword 'let'keyword 'type'keyword 'extern'keyword 'with'keyword 'when'keyword 'if'keyword 'else'keyword 'in'keyword 'as'keyword 'defer'keyword 'break'keyword 'return'keyword 'continue'"

var _Kind_user_index = [...]uint16{0, 17, 28, 35, 39, 47, 51, 61, 72, 85, 99, 102, 105, 108, 111, 114, 117, 120, 123, 135, 148, 160, 173, 185, 198, 210, 223, 236, 250, 263, 277, 289, 302, 314, 327, 339, 352, 364, 377, 389, 402, 414, 427, 439, 452, 465, 479, 491, 504, 517, 531, 544, 557, 569, 582, 595, 609, 625, 639, 653, 665, 679, 691, 703, 718, 733, 749, 767}

func (i Kind) UserString() string {
	if i >= Kind(len(_Kind_user_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_user_name[_Kind_user_index[i]:_Kind_user_index[i+1]]
}