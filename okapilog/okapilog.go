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
		req.Method, req.Resource,
		strings.Join(req.Params, "\\n"))
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
		strings.Join(res.Params, "\\n"))
}

type Log struct {
	Records []Record
}

func makeHeader(lineno int, fields []string) RecordHeader {
	datetime := fields[0] + " " + fields[1]
	level := fields[2]
	id := fields[4]
	return RecordHeader{
		LineNo:   lineno,
		DateTime: datetime,
		Level:    level,
		Id:       id,
	}
}

func makeRecord(lineno int, fields []string) (Record, error) {
	pdutype := fields[5]
	switch pdutype {
	case "REQ":
		return &Request{
			RecordHeader: makeHeader(lineno, fields),
			Addr:         fields[6],
			Tenant:       fields[7],
			Method:       fields[8],
			Resource:     fields[9],
			Params:       fields[10:],
		}, nil
	case "RES":
		rsTimeStr := strings.TrimSuffix(fields[7], "us")
		rsTime, err := strconv.Atoi(rsTimeStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid response time '%s'",
				rsTimeStr)
		}
		return &Response{
			RecordHeader: makeHeader(lineno, fields),
			StatusCode:   fields[6],
			RsTime:       rsTime,
			Params:       fields[8:],
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
		if len(fields) > 3 && fields[3] == "ProxyContext" {
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
