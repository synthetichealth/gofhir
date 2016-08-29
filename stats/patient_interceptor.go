package stats

import (
	"log"
	"sync"

	"github.com/intervention-engine/fhir/models"
)

var patientLock = sync.RWMutex{}

// PatientStatsCreateInterceptor intercepts any new patient resources added to the database
// and updates the Synthetic Mass population statistics based on that patient's address.
type PatientStatsCreateInterceptor struct {
	DataAccess StatsDataAccess
}

// NewPatientStatsCreateInterceptor returns an initialized instance of PatientStatsCreateInterceptor
func NewPatientStatsCreateInterceptor(pgDataAccess StatsDataAccess) *PatientStatsCreateInterceptor {
	return &PatientStatsCreateInterceptor{DataAccess: pgDataAccess}
}

// Before is unused
func (s *PatientStatsCreateInterceptor) Before(resource interface{}) {}

// After increments population statistics after a patient resource is created.
func (s *PatientStatsCreateInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		err := s.DataAccess.AddPatientStat(patient)
		if err != nil {
			log.Printf("PatientStatsCreateInterceptor: After: %s\n", err.Error())
		}
	}
}

// OnError is unused
func (s *PatientStatsCreateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsUpdateInterceptor intercepts any updated patient resources
// and updates the Synthetic Mass population statistics based on changes to that
// patient resource. Currently we do not support tracking any changes, but this
// interceptor may be implemented at a later date.
type PatientStatsUpdateInterceptor struct {
	DataAccess StatsDataAccess
	// The state of the patient before the database update, for comparison after the database update
	patientsBefore map[string]*models.Patient
}

// NewPatientStatsUpdateInterceptor returns an initialized instance of PatientStatsUpdateInterceptor
func NewPatientStatsUpdateInterceptor(pgDataAccess StatsDataAccess) *PatientStatsUpdateInterceptor {
	interceptor := &PatientStatsUpdateInterceptor{DataAccess: pgDataAccess}
	interceptor.patientsBefore = make(map[string]*models.Patient)
	return interceptor
}

// Before is unused
func (s *PatientStatsUpdateInterceptor) Before(resource interface{}) {}

// After is unused
func (s *PatientStatsUpdateInterceptor) After(resource interface{}) {}

// OnError is unused
func (s *PatientStatsUpdateInterceptor) OnError(err error, resource interface{}) {}

// PatientStatsDeleteInterceptor intercepts any deleted patient resources
// and updates the Synthetic Mass population statistics based on that patient's address.
type PatientStatsDeleteInterceptor struct {
	DataAccess StatsDataAccess
}

// NewPatientStatsDeleteInterceptor returns an initialized instance of PatientStatsDeleteInterceptor
func NewPatientStatsDeleteInterceptor(pgDataAccess StatsDataAccess) *PatientStatsDeleteInterceptor {
	return &PatientStatsDeleteInterceptor{DataAccess: pgDataAccess}
}

// Before is unused
func (s *PatientStatsDeleteInterceptor) Before(resource interface{}) {}

// After decrements population statistics after a patient resource is deleted.
func (s *PatientStatsDeleteInterceptor) After(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		err := s.DataAccess.RemovePatientStat(patient)

		if err != nil {
			log.Printf("PatientStatsDeleteInterceptor: After: %s\n", err.Error())
		}
	}
}

// OnError is unused
func (s *PatientStatsDeleteInterceptor) OnError(err error, resource interface{}) {}
