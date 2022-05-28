package okapilog

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Record interface {
	Header() *RecordHeader
}

type RecordHeader struct {
	LineNo   int
	DateTime string
	Level    string
	Id       string
}

type Request struct {
	RecordHeader
	Addr     string
	Tenant   string
	Method   string
	Resource string
	Params   []string
}

func (req *Request) Header() *RecordHeader {
	return &RecordHeader{
		LineNo:   req.LineNo,
		DateTime: req.DateTime,
		Level:    req.Level,
		Id:       req.Id,
	}
}

func (req *Request) String() string {
	return fmt.Sprintf(
		"[ %d ]\\n%s %s\\n%s\\nREQ %s %s\\n%s %s\\n%s",
		req.LineNo,
		req.DateTime, req.Level,
		req.Id,
		req.Addr, req.Tenant,
		req.Method, wrapString(req.Resource),
		strings.Join(wrapStrings(req.Params), "\\n"))
}

type Response struct {
	RecordHeader
	StatusCode string
	RsTime     int
	Params     []string
}

func (res *Response) Header() *RecordHeader {
	return &RecordHeader{
		LineNo:   res.LineNo,
		DateTime: res.DateTime,
		Level:    res.Level,
		Id:       res.Id,
	}
}

func (res *Response) String() string {
	rsTime := int(math.RoundToEven(float64(res.RsTime) / 1000))
	rsTimeStr := "< 1"
	if rsTime > 0 {
		rsTimeStr = fmt.Sprintf("%d", rsTime)
	}
	return fmt.Sprintf(
		"[ %d ]\\n%s %s\\n%s\\nRES %s ( %s ms )\\n%s",
		res.LineNo,
		res.DateTime, res.Level,
		res.Id,
		res.StatusCode, rsTimeStr,
		strings.Join(wrapStrings(res.Params), "\\n"))
}

func wrapStrings(strs []string) []string {
	var newstrs = make([]string, 0)
	var str string
	for _, str = range strs {
		newstrs = append(newstrs, wrapString(str))
	}
	return newstrs
}

func wrapString(str string) string {
	const maxlen = 60
	var b strings.Builder
	var s = str
	for len(s) > maxlen {
		b.WriteString(s[:maxlen])
		b.WriteRune('\n')
		s = s[maxlen:]
	}
	b.WriteString(s)
	return b.String()
}

type Log struct {
	Records []Record
}

func makeHeader(lineno int, fields []string) RecordHeader {
	datetime := fields[0]
	level := fields[5]
	id := fields[7]
	return RecordHeader{
		LineNo:   lineno,
		DateTime: datetime,
		Level:    level,
		Id:       id,
	}
}

func makeRecord(lineno int, fields []string) (Record, error) {
	pdutype := fields[8]
	switch pdutype {
	case "REQ":
		return &Request{
			RecordHeader: makeHeader(lineno, fields),
			Addr:         fields[9],
			Tenant:       fields[10],
			Method:       fields[11],
			Resource:     fields[12],
			Params:       fields[13:],
		}, nil
	case "RES":
		var f = fields[10]
		var t string
		if f == "-" {
			t = "0"
		} else {
			t = strings.TrimSuffix(f, "us")
		}
		var err error
		var rsTime int
		if rsTime, err = strconv.Atoi(t); err != nil {
			return nil, fmt.Errorf("Invalid response time '%s'", f)
		}
		return &Response{
			RecordHeader: makeHeader(lineno, fields),
			StatusCode:   fields[9],
			RsTime:       rsTime,
			Params:       fields[11:],
		}, nil
	default:
		return nil, fmt.Errorf("Unknown record type '%s'", pdutype)
	}
}

func NewLog(file *os.File) (*Log, error) {
	scanner := bufio.NewScanner(file)
	records := []Record{}
	lineno := 0
	for scanner.Scan() {
		lineno++
		s := scanner.Text()
		if strings.TrimSpace(s) == "" {
			continue
		}
		fields := strings.Fields(s)
		if len(fields) >= 12 && fields[5] == "INFO" && (fields[8] == "REQ" || fields[8] == "RES") {
			rec, err := makeRecord(lineno, fields)
			if err != nil {
				return nil, err
			}
			records = append(records, rec)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	olog := &Log{
		Records: records,
	}
	return olog, nil
}
