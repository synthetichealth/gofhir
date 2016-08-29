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
// and updates the Synthetic Mass population statistics based on that patient's address.
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

// Before caches a patient resource before it's updated, for comparison after the update.
func (s *PatientStatsUpdateInterceptor) Before(resource interface{}) {
	patient, ok := resource.(*models.Patient)

	if ok {
		patientLock.Lock()
		s.patientsBefore[patient.Id] = patient
		patientLock.Unlock()
	} else {
		log.Println("PatientStatsUpdateInterceptor: Before: Failed to cache patient before update")
	}
}

// After compares the updated patient resource to the cached patient resource (from before the update), then
// updates population statistics based on that patient's address.
func (s *PatientStatsUpdateInterceptor) After(resource interface{}) {
	newPatient, ok := resource.(*models.Patient)

	if ok {
		patientLock.RLock()
		oldPatient, found := s.patientsBefore[newPatient.Id]
		patientLock.RUnlock()

		if !found {
			log.Println("PatientStatsUpdateInterceptor: After: Could not find cached patient")
			return
		}

		// delete it from the map to prevent unnecessary use of memory
		patientLock.Lock()
		delete(s.patientsBefore, oldPatient.Id)
		patientLock.Unlock()

		// see if the patient's address (or at least, his/her city) changed, and update stats
		if oldPatient.Address[0].City != "" && oldPatient.Address[0].City != newPatient.Address[0].City {

			var err error

			err = s.DataAccess.RemovePatientStat(oldPatient)
			if err != nil {
				log.Printf("PatientStatsUpdateInterceptor: After: %s\n", err.Error())
				return
			}

			err = s.DataAccess.AddPatientStat(newPatient)
			if err != nil {
				log.Printf("PatientStatsUpdateInterceptor: After: %s\n", err.Error())
			}
		}
	}
}

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
