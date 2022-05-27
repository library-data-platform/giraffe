package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/library-data-platform/giraffe/okapilog"
)

type callEdge struct {
	rec1 okapilog.Record
	rec2 okapilog.Record
}

type callOutput struct {
	graph []callEdge
}

func sortByLineno(graph []callEdge) {
	sort.Slice(graph, func(i, j int) bool {
		if graph[i].rec2 == nil {
			if graph[j].rec2 == nil {
				return graph[i].rec1.Header().LineNo >
					graph[j].rec1.Header().LineNo
			} else {
				return true
			}
		} else {
			if graph[j].rec2 == nil {
				return false
			} else {
				if graph[i].rec1.Header().LineNo ==
					graph[j].rec1.Header().LineNo {
					return graph[i].rec2.Header().LineNo <
						graph[j].rec2.Header().LineNo
				} else {
					return graph[i].rec1.Header().LineNo >
						graph[j].rec1.Header().LineNo
				}
			}
		}
	})
}

func (cg *callGraph) prepareOutput(out *callOutput) {
	for _, rec := range cg.olog.Records {
		rec1 := rec
		out.graph = append(out.graph,
			callEdge{
				rec1: rec1,
				rec2: nil,
			})
	}
	for _, records := range cg.calls {
		for _, rec := range records {
			n := cg.calls[rec.Header().Id]
			if len(n) == 0 {
				continue
			}
			switch recv := rec.(type) {
			case *okapilog.Request:
				for _, s := range n {
					switch sv := s.(type) {
					case *okapilog.Request:
						out.graph = append(out.graph,
							callEdge{
								rec1: recv,
								rec2: sv,
							})
					}
				}
			}
		}
	}
	for _, rq := range cg.requests {
		rss, match := cg.responses[rq.Id]
		if match {
			for _, rs := range rss {
				rec1 := rq
				rec2 := rs
				out.graph = append(out.graph,
					callEdge{
						rec1: &rec1,
						rec2: &rec2,
					})
			}
			//} else {
			//        rec1 := rq
			//        out.graph = append(out.graph,
			//                callEdge{
			//                        rec1: &rec1,
			//                        rec2: nil,
			//                })
		}
	}
	// Output any leftover responses
	//for _, rss := range cg.responses {
	//        for _, rs := range rss {
	//                rec1 := rs
	//                out.graph = append(out.graph,
	//                        callEdge{
	//                                rec1: &rec1,
	//                                rec2: nil,
	//                        })
	//        }
	//}
}

//func write(out *callOutput, file *os.File) {
func write(out *callOutput, file io.WriteCloser, rsTimeFlag *int) {
	fmt.Fprintf(file, "digraph G {\n")
	fmt.Fprintf(file, "    node [shape=record,fontname=\"Helvetica-Bold\",fontcolor=white];\n")
	fmt.Fprintf(file, "    rankdir=LR;\n")
	fmt.Fprintf(file, "    ordering=out;\n")
	fmt.Fprintf(file, "\n")
	for _, edge := range out.graph {
		var color1 string
		var color2 string
		var arrowhead string
		switch v := edge.rec1.(type) {
		case *okapilog.Request:
			if edge.rec2 == nil {
				color1 = "forestgreen"
			} else {
				switch edge.rec2.(type) {
				case *okapilog.Request:
					color1 = "forestgreen"
					color2 = "forestgreen"
					arrowhead = "normal"
				case *okapilog.Response:
					color1 = "forestgreen"
					color2 = "cornflowerblue"
					arrowhead = "odot"
				}
			}
		case *okapilog.Response:
			if edge.rec2 == nil {
				color1 = "cornflowerblue"
				if *rsTimeFlag > 0 &&
					v.RsTime >= (*rsTimeFlag*1000) {
					color1 = "maroon"
				}
			}
		}
		if edge.rec2 == nil {
			fmt.Fprintf(file, "    \"%s\" "+
				"[color=%s,fontcolor=white,style=filled];\n", edge.rec1,
				color1)
		}
		if edge.rec2 != nil {
			//fmt.Fprintf(file, "    \"%s\" "+
			//        "[color=%s,fontcolor=black,style=bold];\n", edge.rec2,
			//        color2)
			fmt.Fprintf(file, "    edge [color=%s,style=bold];\n",
				color2)
			fmt.Fprintf(file,
				"    \"%s\" -> \"%s\" "+
					"[arrowhead=%s];\n", edge.rec1, edge.rec2,
				arrowhead)
		}
	}
	fmt.Fprintf(file, "}\n")
}

type callGraph struct {
	olog      *okapilog.Log
	calls     map[string][]okapilog.Record
	requests  map[string]okapilog.Request
	responses map[string][]okapilog.Response
}

func newCallGraph(olog *okapilog.Log) (*callGraph, error) {
	calls := make(map[string][]okapilog.Record)
	requests := make(map[string]okapilog.Request)
	responses := make(map[string][]okapilog.Response)
	for _, rec := range olog.Records {
		storeCall(calls, rec)
		storeRecord(requests, responses, rec)
	}
	cg := &callGraph{
		olog:      olog,
		calls:     calls,
		requests:  requests,
		responses: responses,
	}
	return cg, nil
}

func storeCall(calls map[string][]okapilog.Record, rec okapilog.Record) {
	ids := strings.Split(rec.Header().Id, ";")
	parent := strings.Join(ids[0:len(ids)-1], ";")
	n := calls[parent]
	if len(n) == 0 {
		nn := make([]okapilog.Record, 1)
		nn[0] = rec
		calls[parent] = nn
	} else {
		n = append(n, rec)
		calls[parent] = n
	}
}

func storeRecord(requests map[string]okapilog.Request,
	responses map[string][]okapilog.Response, rec okapilog.Record) {
	switch recv := rec.(type) {
	case *okapilog.Request:
		requests[recv.Id] = *recv
	case *okapilog.Response:
		r, ok := responses[recv.Id]
		if ok {
			r = append(r, *recv)
			responses[recv.Id] = r
		} else {
			n := []okapilog.Response{}
			n = append(n, *recv)
			responses[recv.Id] = n
		}
	}
}
