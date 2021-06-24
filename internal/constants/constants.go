package constants

import (
	"9fans.net/go/plan9"
	"unicode/utf8"
)

const (
	NRange = 10 // TODO(flux): No reason for this static limit anymore; should we remove?
	//	Infinity  = 0x7FFFFFFF
	MaxBlock  = 8 * 1024
	Blockincr = 256

	EVENTSIZE = 256
	BUFSIZE   = MaxBlock + plan9.IOHDRSZ
	RBUFSIZE  = BUFSIZE / utf8.UTFMax

	Empty    = 0
	Null     = '-'
	Delete   = 'd'
	Insert   = 'i'
	Replace  = 'r'
	Filename = 'f'

	Inactive   = 0
	Inserting  = 1
	Collecting = 2

	// Always apply display scalesize to these.
	Border       = 2
	ButtonBorder = 2
	Scrollwid    = 12
	Scrollgap    = 8

	KF             = 0xF000 // Start of private unicode space
	Kscrolloneup   = KF | 0x20
	Kscrollonedown = KF | 0x21
)
