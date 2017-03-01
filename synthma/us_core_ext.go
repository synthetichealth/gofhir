package synthma

import (
	"errors"
	"fmt"

	"github.com/intervention-engine/fhir/search"
	"gopkg.in/mgo.v2/bson"
)

// Custom search parameters for us-core FHIR extensions (race, ethnicity).
func init() {
	registry := search.GlobalRegistry()

	// Register the us-core-race parameter.
	fmt.Println("Registered 'race' search parameter on 'Patient'.")
	registry.RegisterParameterInfo(USCoreRaceParamInfo)
	registry.RegisterParameterParser(USCoreRaceParamInfo.Type, USCoreRaceParamParser)
	search.GlobalMongoRegistry().RegisterBSONBuilder(USCoreRaceParamInfo.Type, USCoreRaceBSONBuilder)

	// Register the us-core-ethnicity parameter.
	fmt.Println("Registered 'ethnicity' search parameter on 'Patient'.")
	registry.RegisterParameterInfo(USCoreEthnicityParamInfo)
	registry.RegisterParameterParser(USCoreEthnicityParamInfo.Type, USCoreEthnicityParamParser)
	search.GlobalMongoRegistry().RegisterBSONBuilder(USCoreEthnicityParamInfo.Type, USCoreEthnicityBSONBuilder)
}

// USCoreRaceParam represents the "us-core-race" search parameter. Patient's may be searched
// by their race using a combination of system and code. This behaves exactly the same
// as a standard TokenParam. See: http://hl7.org/fhir/2017jan/search.html#token
type USCoreRaceParam struct {
	search.TokenParam
}

var USCoreRaceParamInfo = search.SearchParamInfo{
	Resource: "Patient",
	Name:     "race",
	Type:     "synthma.race",
}

// USCoreRaceParamParser parses the parameter and returns a USCoreRaceParam.
var USCoreRaceParamParser = func(info search.SearchParamInfo, data search.SearchParamData) (search.SearchParam, error) {
	return &USCoreRaceParam{
		TokenParam: *search.ParseTokenParam(data.Value, info),
	}, nil
}

// USCoreRaceBSONBuilder builds the Mongo BSON object corresponding to the query by Patient race.
var USCoreRaceBSONBuilder = func(param search.SearchParam, searcher *search.MongoSearcher) (object bson.M, err error) {
	p, ok := param.(*USCoreRaceParam)
	if !ok {
		return nil, errors.New("Expected a USCoreRaceParam")
	}

	system := "http://hl7.org/fhir/v3/Race"
	if p.System != "" {
		system = p.System
	}

	return bson.M{
		"extension.us-core-race.coding.system": system,
		"extension.us-core-race.coding.code":   p.Code,
	}, nil
}

// USCoreEthnicityParam represents the "us-core-ethnicity" search parameter. Patient's may be searched
// by their ethnicity using a combination of system and code. This behaves exactly the same
// as a standard TokenParam. See: http://hl7.org/fhir/2017jan/search.html#token
type USCoreEthnicityParam struct {
	search.TokenParam
}

var USCoreEthnicityParamInfo = search.SearchParamInfo{
	Resource: "Patient",
	Name:     "ethnicity",
	Type:     "synthma.ethnicity",
}

// USCoreEthnicityParamParser parses the parameter and returns a USCoreEthnicityParam.
var USCoreEthnicityParamParser = func(info search.SearchParamInfo, data search.SearchParamData) (search.SearchParam, error) {
	return &USCoreEthnicityParam{
		TokenParam: *search.ParseTokenParam(data.Value, info),
	}, nil
}

// USCoreEthnicityBSONBuilder builds the Mongo BSON object corresponding to the query by Patient ethnicity.
var USCoreEthnicityBSONBuilder = func(param search.SearchParam, searcher *search.MongoSearcher) (object bson.M, err error) {
	p, ok := param.(*USCoreEthnicityParam)
	if !ok {
		return nil, errors.New("Expected a USCoreEthnicityParam")
	}

	system := "http://hl7.org/fhir/v3/Ethnicity"
	if p.System != "" {
		system = p.System
	}

	return bson.M{
		"extension.us-core-ethnicity.coding.system": system,
		"extension.us-core-ethnicity.coding.code":   p.Code,
	}, nil
}
