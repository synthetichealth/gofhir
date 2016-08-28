package stats

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/intervention-engine/fhir/models"
	_ "github.com/lib/pq"
)

// Condition coding systems
const SnomedCodeSystem = "http://snomed.info/sct"

var InvalidGenderError = errors.New("Invalid gender")
var InvalidUpdateOperationError = errors.New("Invalid update operation")

// StatsDataAccess is the top-level interface for interacting with the Postgres database.
type StatsDataAccess struct {
	DB *sql.DB
}

func (da *StatsDataAccess) AddPatient(patient *models.Patient) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateStats(countyfp, cousubfp, patient.Gender, "increment")

	} else {
		err = InvalidGenderError
	}
	return
}

func (da *StatsDataAccess) RemovePatient(patient *models.Patient) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		err = da.updateStats(countyfp, cousubfp, patient.Gender, "decrement")

	} else {
		err = InvalidGenderError
	}
	return
}

func (da *StatsDataAccess) AddPatientWithCondition(patient *models.Patient, condition *models.Condition) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	diseasefp, err := da.identifyDiseaseForCondition(condition)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		da.updateFacts(countyfp, cousubfp, diseasefp, patient.Gender, "increment")

	} else {
		err = InvalidGenderError
	}
	return
}

func (da *StatsDataAccess) RemovePatientWithCondition(patient *models.Patient, condition *models.Condition) (err error) {

	countyfp, cousubfp, err := da.identifyCountyAndSubdivisionForPatient(patient)
	diseasefp, err := da.identifyDiseaseForCondition(condition)

	if err != nil {
		return
	}

	if patientGenderIsValid(patient) {
		da.updateFacts(countyfp, cousubfp, diseasefp, patient.Gender, "decrement")

	} else {
		err = InvalidGenderError
	}
	return
}

func (da *StatsDataAccess) identifyCountyAndSubdivisionForPatient(patient *models.Patient) (countyfp, cousubfp string, err error) {
	err = da.DB.QueryRow("SELECT countyfp, cousubfp FROM tiger.cousub WHERE name = $1", patient.Address[0].City).Scan(&countyfp, &cousubfp)
	return
}

func (da *StatsDataAccess) identifyDiseaseForCondition(condition *models.Condition) (diseasefp string, err error) {
	err = da.DB.QueryRow("SELECT diseasefp FROM synth_ma.synth_disease WHERE code_snomed = $1", getSnomedCode(condition)).Scan(&diseasefp)
	return
}

func (da *StatsDataAccess) updateStats(countyfp, cousubfp, gender, op string) (err error) {

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return InvalidUpdateOperationError
	}

	stmt, err := da.DB.Prepare(fmt.Sprintf("UPDATE synth_ma.synth_county_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %s 1) / sq_mi) WHERE ct_fips = (?)", symbol, gender, gender, symbol, symbol))
	_, err = stmt.Exec(countyfp)
	err = stmt.Close()
	stmt, err = da.DB.Prepare(fmt.Sprintf("UPDATE synth_ma.synth_cousub_stats SET pop = pop %s 1, pop_%s = pop_%s %s 1, pop_sm = ((pop %S 1) / sq_mi) WHERE cs_fips = (?)", symbol, gender, gender, symbol, symbol))
	_, err = stmt.Exec(cousubfp)
	err = stmt.Close()
	return
}

func (da *StatsDataAccess) updateFacts(countyfp, cousubfp, diseasefp, gender, op string) (err error) {

	var symbol string
	switch op {
	case "increment":
		symbol = "+"
	case "decrement":
		symbol = "-"
	default:
		return InvalidUpdateOperationError
	}

	stmt, err := da.DB.Prepare(fmt.Sprintf("UPDATE synth_ma.synth_county_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE countyfp = (?) AND diseasefp = (?)", symbol, gender, gender, symbol))
	_, err = stmt.Exec(countyfp, diseasefp)
	err = stmt.Close()
	stmt, err = da.DB.Prepare(fmt.Sprintf("UPDATE synth_ma.synth_cousub_facts SET pop = pop %s 1, pop_%s = pop_%s %s 1 WHERE cousubfp = (?) AND diseasefp = (?)", symbol, gender, gender, symbol))
	_, err = stmt.Exec(cousubfp, diseasefp)
	err = stmt.Close()
	return
}

func patientGenderIsValid(patient *models.Patient) bool {
	return (patient.Gender == "male" || patient.Gender == "female")
}

// getSnomedCode returns the condition's SNOMED code, if it exists
func getSnomedCode(condition *models.Condition) string {
	codings := condition.Code.Coding
	for _, coding := range codings {
		if coding.System == SnomedCodeSystem {
			return coding.Code
		}
	}
	return ""
}
