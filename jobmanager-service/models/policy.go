package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"etsn/server/jobmanager-service/utils/logs"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"moul.io/http2curl"
)

var (
	policyManagerBaseURL = os.Getenv("POLICYMANAGER_URL")
)

type Notification struct {
	ID           uuid.UUID `json:"-"`
	AppInstance  uuid.UUID `json:"app_instance"`
	CommonAction Action    `json:"common_action"`
	Service      string    `json:"service"`
	Manifest     string    `json:"app_descriptor"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Action struct {
	URI             string            `json:"uri"`
	HTTPMethod      string            `json:"http_method"`
	ExtraParameters map[string]string `json:"updated_at"`
}

type ExtraParameters struct {
	JobGoupId uuid.UUID `json:"updated_at"`
}
type Incompliance struct {
	// gorm.Model
	ID           uuid.UUID `gorm:"type:char(36);primary_key"` // lets abstract this id from the shell user -> TODO: should be uuid
	ResourceID   uuid.UUID `json:"resource_id"`
	CurrentValue string    `gorm:"type:text" json:"current_value,omitempty"`
	Threshold    string    `gorm:"type:text" json:"threshold,omitempty"`
	PolicyName   string    `gorm:"type:text" json:"policy_name"`
	PolicyID     uuid.UUID `gorm:"type:text" json:"policy_id"`
	ExtraLabels  []string  `gorm:"type:text" json:"extra_labels,omitempty"`
	Subject      Subject   `gorm:"type:text" json:"subject,omitempty"`
	Remediation  JobType   `gorm:"type:text" json:"remediation"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subject struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key"`
	Type         string    `gorm:"type:text" json:"type,omitempty"`
	AppName      string    `gorm:"type:text" json:"app_name,omitempty"`
	AppComponent string    `gorm:"type:text" json:"app_component,omitempty"`
	AppInstance  string    `gorm:"type:text" json:"app_instance,omitempty"`
}

func NotifyPolicyManager(db *gorm.DB, manifest string, jobGroup JobGroup, token string) (err error) {
	// create notification body first
	notification := Notification{
		AppInstance: jobGroup.ID,
		Service:     "job-manager",
		CommonAction: Action{
			URI:        "/jobmanager/policies/incompliance/create",
			HTTPMethod: "POST",
		},
		Manifest:  manifest,
		UpdatedAt: time.Now(),
	}
	bodyBytes, err := json.Marshal(notification)

	// create PM request
	req, err := http.NewRequest("POST", policyManagerBaseURL+"/polman/registry/api/v1/", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return
	}

	// add content type
	req.Header.Set("Content-Type", "application/json")
	// forward the authorization token
	req.Header.Add("Authorization", token)

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}
	defer resp.Body.Close()

	command, _ := http2curl.GetCurlCommand(req)
	logs.Logger.Println("Request sent to Policy Manager Service: ")
	logs.Logger.Println(command)
	logs.Logger.Println("End Policy Manager Request.")

	// if response is OK
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return err

	} else {
		err := errors.New("Bad response from Policy Manager: status code - " + string(rune(resp.StatusCode)))
		return err
	}
}

func (v *Incompliance) SaveIncompliance(db *gorm.DB) (*Incompliance, error) {
	err := db.Debug().Create(&v).Error
	if err != nil {
		return &Incompliance{}, err
	}
	return v, nil
}
