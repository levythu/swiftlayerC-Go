package exception

import "errors"

var EX_WRONG_FILEFORMAT=errors.New("exception.io.wrong_format")
var EX_IMPROPER_DATA=errors.New("exception.io.improper_data")
var EX_IO_ERROR=errors.New("exception.io.error")
