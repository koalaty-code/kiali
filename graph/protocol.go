package graph

import (
	"strings"

	"github.com/kiali/kiali/log"
)

type Rate struct {
	Name         string
	IsErr        bool
	IsIn         bool
	IsOut        bool
	IsPercentErr bool
	IsPercentReq bool
	IsTotal      bool
	Precision    int
}

type Protocol struct {
	Name      string
	EdgeRates []Rate
	NodeRates []Rate
	Unit      string
	UnitShort string
}

/* TODO: For now, treat gRPC like HTTP, uncomment this when we want to treat gRPC explicitly
var GRPC Protocol = Protocol{
	Name: "grpc",
	EdgeRates: []Rate{
		Rate{Name: "grpc", IsTotal: true, Precision: 2},
		Rate{Name: "grpcErr", IsErr: true, Precision: 2},
		Rate{Name: "grpcPercentErr", IsPercentErr: true, Precision: 1},
		Rate{Name: "grpcPercentReq", IsPercentReq: true, Precision: 1},
	},
	NodeRates: []Rate{
		Rate{Name: "grpcIn", IsIn: true, Precision: 2},
		Rate{Name: "grpcInErr", IsErr: true, Precision: 2},
		Rate{Name: "grpcOut", IsOut: true, Precision: 2},
	},
	Unit:      "requests per second",
	UnitShort: "rps",
}
*/
var HTTP Protocol = Protocol{
	Name: "http",
	EdgeRates: []Rate{
		Rate{Name: "http", IsTotal: true, Precision: 2},
		Rate{Name: "http3xx", Precision: 2},
		Rate{Name: "http4xx", IsErr: true, Precision: 2},
		Rate{Name: "http5xx", IsErr: true, Precision: 2},
		Rate{Name: "httpPercentErr", IsPercentErr: true, Precision: 1},
		Rate{Name: "httpPercentReq", IsPercentReq: true, Precision: 1},
	},
	NodeRates: []Rate{
		Rate{Name: "httpIn", IsIn: true, Precision: 2},
		Rate{Name: "httpIn3xx", Precision: 2},
		Rate{Name: "httpIn4xx", IsErr: true, Precision: 2},
		Rate{Name: "httpIn5xx", IsErr: true, Precision: 2},
		Rate{Name: "httpOut", IsOut: true, Precision: 2},
	},
	Unit:      "requests per second",
	UnitShort: "rps",
}
var TCP Protocol = Protocol{
	Name: "tcp",
	EdgeRates: []Rate{
		Rate{Name: "tcp", IsTotal: true, Precision: 2},
	},
	NodeRates: []Rate{
		Rate{Name: "tcpIn", IsIn: true, Precision: 2},
		Rate{Name: "tcpOut", IsOut: true, Precision: 2},
	},
	Unit:      "bytes per second",
	UnitShort: "bps",
}

// TODO: For now, treat gRPC like HTTP, uncomment this when we want to treat gRPC explicitly
// var Protocols []Protocol = []Protocol{GRPC, HTTP, TCP}
var Protocols []Protocol = []Protocol{HTTP, TCP}

func AddToMetadata(protocol string, val float64, code string, sourceMetadata, destMetadata, edgeMetadata map[string]interface{}) {
	switch protocol {
	case "grpc":
		// TODO: For now, treat gRPC like HTTP, uncomment this (and remove the http line below it) when we want to treat gRPC explicitly
		// addToMetadataGrpc(val, code, sourceMetadata, destMetadata, edgeMetadata)
		addToMetadataHttp(val, code, sourceMetadata, destMetadata, edgeMetadata)
	case "http":
		addToMetadataHttp(val, code, sourceMetadata, destMetadata, edgeMetadata)
	case "tcp":
		addToMetadataTcp(val, code, sourceMetadata, destMetadata, edgeMetadata)
	default:
		log.Tracef("Ignore unhandled metadata protocol [%s]", protocol)
	}
}

/* TODO: For now, treat gRPC like HTTP, uncomment this when we want to treat gRPC explicitly
func addToMetadataGrpc(val float64, code string, sourceMetadata, destMetadata, edgeMetadata map[string]interface{}) {
	addToMetadataValue(sourceMetadata, "grpcOut", val)
	addToMetadataValue(destMetadata, "grpcIn", val)
	addToMetadataValue(edgeMetadata, "grpc", val)

	// Istio telemetry may use HTTP codes for gRPC, so if it quacks like a duck...
	isHttpCode := len(code) == 3
	isErr := false
	if isHttpCode {
		isErr = strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5")
	} else {
		isErr = code != "0"
	}
	if isErr {
		addToMetadataValue(destMetadata, "grpcInErr", val)
		addToMetadataValue(edgeMetadata, "grpcErr", val)
	}
}
*/

func addToMetadataHttp(val float64, code string, sourceMetadata, destMetadata, edgeMetadata map[string]interface{}) {
	addToMetadataValue(sourceMetadata, "httpOut", val)
	addToMetadataValue(destMetadata, "httpIn", val)
	addToMetadataValue(edgeMetadata, "http", val)

	// note, we don't track 2xx because it's not used downstream and can be easily
	// calculated: 2xx = (rate - 3xx - 4xx - 5xx)
	switch {
	case strings.HasPrefix(code, "3"):
		addToMetadataValue(destMetadata, "httpIn3xx", val)
		addToMetadataValue(edgeMetadata, "http3xx", val)
	case strings.HasPrefix(code, "4"):
		addToMetadataValue(destMetadata, "httpIn4xx", val)
		addToMetadataValue(edgeMetadata, "http4xx", val)
	case strings.HasPrefix(code, "5"):
		addToMetadataValue(destMetadata, "httpIn5xx", val)
		addToMetadataValue(edgeMetadata, "http5xx", val)
	}
}

func addToMetadataTcp(val float64, code string, sourceMetadata, destMetadata, edgeMetadata map[string]interface{}) {
	addToMetadataValue(sourceMetadata, "tcpOut", val)
	addToMetadataValue(destMetadata, "tcpIn", val)
	addToMetadataValue(edgeMetadata, "tcp", val)
}

func AddOutgoingEdgeToMetadata(sourceMetadata, edgeMetadata map[string]interface{}) {
	/* TODO: For now, treat gRPC like HTTP, uncomment this when we want to treat gRPC explicitly
	if val, valOk := edgeMetadata["grpc"]; valOk {
		addToMetadataValue(sourceMetadata, "grpcOut", val.(float64))
	}
	*/
	if val, valOk := edgeMetadata["http"]; valOk {
		addToMetadataValue(sourceMetadata, "httpOut", val.(float64))
	}
	if val, valOk := edgeMetadata["tcp"]; valOk {
		addToMetadataValue(sourceMetadata, "tcpOut", val.(float64))
	}
}

func addToMetadataValue(md map[string]interface{}, k string, v float64) {
	if curr, ok := md[k]; ok {
		md[k] = curr.(float64) + v
	} else {
		md[k] = v
	}
}

func averageMetadataValue(md map[string]interface{}, k string, v float64) {
	total := v
	count := 1.0
	kTotal := k + "_total"
	kCount := k + "_count"
	if prevTotal, ok := md[kTotal]; ok {
		total += prevTotal.(float64)
	}
	if prevCount, ok := md[kCount]; ok {
		count += prevCount.(float64)
	}
	md[kTotal] = total
	md[kCount] = count
	md[k] = total / count
}