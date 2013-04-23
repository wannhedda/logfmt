package logfmt

import (
	"fmt"
)

func gotoScanner(data []byte, h Handler) (err error) {
	saveError := func(e error) {
		if err == nil {
			err = e
		}
	}

	var c byte
	var i int
	var m int
	var key []byte
	var val []byte
	var ok bool
	var esc bool

garbage:
	if i == len(data) {
		goto eof
	}

	c = data[i]
	switch {
	case c > ' ' && c != '"' && c != '=':
		m = -1
		key, val = nil, nil
		goto key
	default:
		i++
		goto garbage
	}

key:
	if i >= len(data) {
		goto eof
	}

	c = data[i]
	switch {
	case c > ' ' && c != '"' && c != '=':
		if m < 0 {
			m = i
		}
		i++
		goto key
	case c == '=':
		key = data[m:i]
		i++
		goto value
	default:
		if m >= 0 {
			key = data[m:i]
			saveError(h.HandleLogfmt(key, nil))
		}
		i++
		goto garbage
	}

value:
	if i >= len(data) {
		goto eof
	}

	c = data[i]
	switch {
	case c > ' ' && c != '"' && c != '=':
		m = i
		i++
		goto ivalue
	case c == '"':
		m = i
		i++
		esc = false
		goto qvalue
	default:
		if key != nil {
			saveError(h.HandleLogfmt(key, val))
		}
		i++
		goto garbage
	}

ivalue:
	if i >= len(data) {
		goto eof
	}

	c = data[i]
	switch {
	case c > ' ' && c != '"' && c != '=':
		i++
		goto ivalue
	default:
		val = data[m:i]
		saveError(h.HandleLogfmt(key, val))
		i++
		goto garbage
	}

qvalue:
	if i >= len(data) {
		goto eof
	}

	c = data[i]
	switch c {
	case '\\':
		i += 2
		esc = true
		goto qvalue
	case '"':
		i++
		val = data[m:i]
		if esc {
			val, ok = unquoteBytes(val)
			if !ok {
				saveError(fmt.Errorf("logfmt: error unquoting bytes %q", string(val)))
				goto garbage
			}
		} else {
			val = val[1 : len(val)-1]
		}
		saveError(h.HandleLogfmt(key, val))
		goto garbage
	default:
		i++
		goto qvalue
	}

eof:
	if key != nil {
		saveError(h.HandleLogfmt(key, val))
	}

	return
}
