package synthma

import (
	"errors"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	// Register the condition-code parameter
	search.GlobalRegistry().RegisterParameterInfo(ConditionCodeParamInfo)
	search.GlobalRegistry().RegisterParameterParser(ConditionCodeParamInfo.Type, ConditionCodeParamParser)
	search.GlobalMongoRegistry().RegisterBSONBuilder(ConditionCodeParamInfo.Type, ConditionCodeBSONBuilder)
}

// ConditionCodeParam represents the condition-code search parameter. Patient's may be searched
// by conditions they have using a combination of system and code. This behaves exactly the same
// as a standard TokenParam. See: http://hl7.org/fhir/2016Sep/search.html#token
type ConditionCodeParam struct {
	search.TokenParam
}

// ConditionCodeParamInfo represents the condition-code for Patients' conditions. This allows
// patients to be searched by conditions they have.
var ConditionCodeParamInfo = search.SearchParamInfo{
	Resource: "Patient",
	Name:     "condition-code",
	Type:     "synthma.patient_condition_code",
}

// ConditionCodeParamParser parses the parameter and returns a ConditionCodeParam.
var ConditionCodeParamParser = func(info search.SearchParamInfo, data search.SearchParamData) (search.SearchParam, error) {
	return &ConditionCodeParam{
		TokenParam: *search.ParseTokenParam(data.Value, info),
	}, nil
}

// ConditionCodeBSONBuilder builds the Mongo BSON object corresponding to the query by Patient condition.
var ConditionCodeBSONBuilder = func(param search.SearchParam, searcher *search.MongoSearcher) (object bson.M, err error) {

	cc, ok := param.(*ConditionCodeParam)
	if !ok {
		return nil, errors.New("Expected a ConditionCodeParam")
	}

	// First get a list of patient IDs for this condition
	var subjectWrappers []struct {
		Subject *models.Reference `bson:"subject,omitempty"`
	}
	if err := searcher.GetDB().C("conditions").Find(bson.M{"code.coding.code": cc.Code}).Select(bson.M{"_id": 0, "subject": 1}).All(&subjectWrappers); err != nil {
		return nil, err
	}

	patientIDMap := make(map[string]bool)
	for _, wrapper := range subjectWrappers {
		if wrapper.Subject.Type == "Patient" {
			patientIDMap[wrapper.Subject.ReferencedID] = true
		}
	}

	// convert map of patient IDs to a list
	patientIDList := make([]string, len(patientIDMap))
	i := 0
	for k := range patientIDMap {
		patientIDList[i] = k
		i++
	}

	// Return a BSON object (for Patient) indicating the set of Patient IDs that should be used
	return bson.M{
		"_id": bson.M{
			"$in": patientIDList,
		},
	}, nil
}
