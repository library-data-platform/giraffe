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
	datetime := fields[0]
	level := fields[1]
	id := fields[3]
	return RecordHeader{
		LineNo:   lineno,
		DateTime: datetime,
		Level:    level,
		Id:       id,
	}
}

func makeRecord(lineno int, fields []string) (Record, error) {
	pdutype := fields[4]
	switch pdutype {
	case "REQ":
		return &Request{
			RecordHeader: makeHeader(lineno, fields),
			Addr:         fields[5],
			Tenant:       fields[6],
			Method:       fields[7],
			Resource:     fields[8],
			Params:       fields[9:],
		}, nil
	case "RES":
		rsTimeStr := strings.TrimSuffix(fields[6], "us")
		rsTime, err := strconv.Atoi(rsTimeStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid response time '%s'",
				rsTimeStr)
		}
		return &Response{
			RecordHeader: makeHeader(lineno, fields),
			StatusCode:   fields[5],
			RsTime:       rsTime,
			Params:       fields[7:],
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
		if len(fields) > 2 && fields[2] == "ProxyContext" {
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
