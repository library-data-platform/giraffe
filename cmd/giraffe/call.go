package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type callEdge struct {
	rec1 record
	rec2 record
}

type callOutput struct {
	graph []callEdge
}

func sortByLineno(graph []callEdge) {
	sort.Slice(graph, func(i, j int) bool {
		if graph[i].rec2 == nil {
			if graph[j].rec2 == nil {
				return graph[i].rec1.header().lineno >
					graph[j].rec1.header().lineno
			} else {
				return true
			}
		} else {
			if graph[j].rec2 == nil {
				return false
			} else {
				if graph[i].rec1.header().lineno ==
					graph[j].rec1.header().lineno {
					return graph[i].rec2.header().lineno <
						graph[j].rec2.header().lineno
				} else {
					return graph[i].rec1.header().lineno >
						graph[j].rec1.header().lineno
				}
			}
		}
	})
}

func (cg *callGraph) prepareOutput(out *callOutput) {
	for _, rec := range cg.olog.records {
		rec1 := rec
		out.graph = append(out.graph,
			callEdge{
				rec1: rec1,
				rec2: nil,
			})
	}
	for _, records := range cg.calls {
		for _, rec := range records {
			n := cg.calls[rec.header().id]
			if len(n) == 0 {
				continue
			}
			switch recv := rec.(type) {
			case *requestRecord:
				for _, s := range n {
					switch sv := s.(type) {
					case *requestRecord:
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
		rss, match := cg.responses[rq.id]
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

func write(out *callOutput, file *os.File) {
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
		case *requestRecord:
			if edge.rec2 == nil {
				color1 = "forestgreen"
			} else {
				switch edge.rec2.(type) {
				case *requestRecord:
					color1 = "forestgreen"
					color2 = "forestgreen"
					arrowhead = "normal"
				case *responseRecord:
					color1 = "forestgreen"
					color2 = "cornflowerblue"
					arrowhead = "odot"
				}
			}
		case *responseRecord:
			if edge.rec2 == nil {
				color1 = "cornflowerblue"
				if v.timing > 250000 {
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
	olog      *okapiLog
	calls     map[string][]record
	requests  map[string]requestRecord
	responses map[string][]responseRecord
}

func newCallGraph(olog *okapiLog) (*callGraph, error) {
	calls := make(map[string][]record)
	requests := make(map[string]requestRecord)
	responses := make(map[string][]responseRecord)
	for _, rec := range olog.records {
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

func storeCall(calls map[string][]record, rec record) {
	ids := strings.Split(rec.header().id, ";")
	parent := strings.Join(ids[0:len(ids)-1], ";")
	n := calls[parent]
	if len(n) == 0 {
		nn := make([]record, 1)
		nn[0] = rec
		calls[parent] = nn
	} else {
		n = append(n, rec)
		calls[parent] = n
	}
}

func storeRecord(requests map[string]requestRecord,
	responses map[string][]responseRecord, rec record) {
	switch recv := rec.(type) {
	case *requestRecord:
		requests[recv.id] = *recv
	case *responseRecord:
		r, ok := responses[recv.id]
		if ok {
			r = append(r, *recv)
			responses[recv.id] = r
		} else {
			n := []responseRecord{}
			n = append(n, *recv)
			responses[recv.id] = n
		}
	}
}
