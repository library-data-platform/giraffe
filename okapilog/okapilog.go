package okapilog

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type recordHeader struct {
	lineno   int
	datetime string
	level    string
	id       string
}

type requestRecord struct {
	recordHeader
	addr     string
	tenant   string
	method   string
	resource string
	params   []string
}

type responseRecord struct {
	recordHeader
	statusCode string
	rsTime     int
	params     []string
}

type record interface {
	header() *recordHeader
}

func (req *requestRecord) header() *recordHeader {
	return &recordHeader{
		lineno:   req.lineno,
		datetime: req.datetime,
		level:    req.level,
		id:       req.id,
	}
}

func (req *requestRecord) String() string {
	return fmt.Sprintf(
		"[ %d ]\\n%s %s\\n%s\\nREQ %s %s\\n%s %s\\n%s",
		req.lineno,
		req.datetime, req.level,
		req.id,
		req.addr, req.tenant,
		req.method, req.resource,
		strings.Join(req.params, "\\n"))
}

func (res *responseRecord) header() *recordHeader {
	return &recordHeader{
		lineno:   res.lineno,
		datetime: res.datetime,
		level:    res.level,
		id:       res.id,
	}
}

func (res *responseRecord) String() string {
	rsTime := int(math.RoundToEven(float64(res.rsTime) / 1000))
	rsTimeStr := "< 1"
	if rsTime > 0 {
		rsTimeStr = fmt.Sprintf("%d", rsTime)
	}
	return fmt.Sprintf(
		"[ %d ]\\n%s %s\\n%s\\nRES %s ( %s ms )\\n%s",
		res.lineno,
		res.datetime, res.level,
		res.id,
		res.statusCode, rsTimeStr,
		strings.Join(res.params, "\\n"))
}

type okapiLog struct {
	records []record
}

func makeHeader(lineno int, fields []string) recordHeader {
	datetime := fields[0] + " " + fields[1]
	level := fields[2]
	id := fields[4]
	return recordHeader{
		lineno:   lineno,
		datetime: datetime,
		level:    level,
		id:       id,
	}
}

func makeOkapiRecord(lineno int, fields []string) (record, error) {
	pdutype := fields[5]
	switch pdutype {
	case "REQ":
		return &requestRecord{
			recordHeader: makeHeader(lineno, fields),
			addr:         fields[6],
			tenant:       fields[7],
			method:       fields[8],
			resource:     fields[9],
			params:       fields[10:],
		}, nil
	case "RES":
		rsTimeStr := strings.TrimSuffix(fields[7], "us")
		rsTime, err := strconv.Atoi(rsTimeStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid response time '%s'",
				rsTimeStr)
		}
		return &responseRecord{
			recordHeader: makeHeader(lineno, fields),
			statusCode:   fields[6],
			rsTime:       rsTime,
			params:       fields[8:],
		}, nil
	default:
		return nil, fmt.Errorf("Unknown record type '%s'", pdutype)
	}
}

func newOkapiLog(file *os.File) (*okapiLog, error) {
	scanner := bufio.NewScanner(file)
	records := []record{}
	lineno := 0
	for scanner.Scan() {
		lineno++
		s := scanner.Text()
		if strings.TrimSpace(s) == "" {
			continue
		}
		fields := strings.Fields(s)
		if len(fields) > 3 && fields[3] == "ProxyContext" {
			rec, err := makeOkapiRecord(lineno, fields)
			if err != nil {
				return nil, err
			}
			records = append(records, rec)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	olog := &okapiLog{
		records: records,
	}
	return olog, nil
}
