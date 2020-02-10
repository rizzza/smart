// Copyright 2017-18 Daniel Swarbrick. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drivedb

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// SMART attribute conversion rule
type AttrConv struct {
	Conv string `yaml:"conv"`
	Name string `yaml:"name"`
}

type DriveModel struct {
	Family         string              `yaml:"family"`
	ModelRegex     string              `yaml:"model_regex"`
	FirmwareRegex  string              `yaml:"firmware_regex"`
	WarningMsg     string              `yaml:"warning"`
	Presets        map[string]AttrConv `yaml:"presets"`
	CompiledRegexp *regexp.Regexp
}

type DriveDb struct {
	Drives []DriveModel `yaml:"drives"`
}

var DB DriveDb

func init() {
	DB = DriveDb{
		Drives: []DriveModel{
			{
				Family: "DEFAULT",
				Presets: map[string]AttrConv{
					"1":   {Conv: "Raw48", Name: "Raw_Read_Error_Rate"},
					"2":   {Conv: "Raw48", Name: "Throughput_Performance"},
					"3":   {Conv: "raw16(avg16)", Name: "Spin_Up_Time"},
					"4":   {Conv: "raw48", Name: "Start_Stop_Count"},
					"5":   {Conv: "raw16(raw16)", Name: "Reallocated_Sector_Ct"},
					"6":   {Conv: "raw48", Name: "Read_Channel_Margin"},
					"7":   {Conv: "raw48", Name: "Seek_Error_Rate"},
					"8":   {Conv: "raw48", Name: "Seek_Time_Performance"},
					"9":   {Conv: "raw24(raw8)", Name: "Power_On_Hours"},
					"10":  {Conv: "raw48", Name: "Spin_Retry_Count"},
					"11":  {Conv: "raw48", Name: "Calibration_Retry_Count"},
					"12":  {Conv: "raw48", Name: "Power_Cycle_Count"},
					"13":  {Conv: "raw48", Name: "Read_Soft_Error_Rate"},
					"175": {Conv: "raw48", Name: "Program_Fail_Count_Chip"},
					"176": {Conv: "raw48", Name: "Erase_Fail_Count_Chip"},
					"177": {Conv: "raw48", Name: "Wear_Leveling_Count"},
					"178": {Conv: "raw48", Name: "Used_Rsvd_Blk_Cnt_Chip"},
					"179": {Conv: "raw48", Name: "Used_Rsvd_Blk_Cnt_Tot"},
					"180": {Conv: "raw48", Name: "Unused_Rsvd_Blk_Cnt_Tot"},
					"181": {Conv: "raw48", Name: "Program_Fail_Cnt_Total"},
					"182": {Conv: "raw48", Name: "Erase_Fail_Count_Total"},
					"183": {Conv: "raw48", Name: "Runtime_Bad_Block"},
					"184": {Conv: "raw48", Name: "End-to-End_Error"},
					"187": {Conv: "raw48", Name: "Reported_Uncorrect"},
					"188": {Conv: "raw48", Name: "Command_Timeout"},
					"189": {Conv: "raw48", Name: "High_Fly_Writes"},
					"190": {Conv: "tempminmax", Name: "Airflow_Temperature_Cel"},
					"191": {Conv: "raw48", Name: "G-Sense_Error_Rate"},
					"192": {Conv: "raw48", Name: "Power-Off_Retract_Count"},
					"193": {Conv: "raw48", Name: "Load_Cycle_Count"},
					"194": {Conv: "tempminmax", Name: "Temperature_Celsius"},
					"195": {Conv: "raw48", Name: "Hardware_ECC_Recovered"},
					"196": {Conv: "raw16(raw16)", Name: "Reallocated_Event_Count"},
					"197": {Conv: "raw48", Name: "Current_Pending_Sector"},
					"198": {Conv: "raw48", Name: "Offline_Uncorrectable"},
					"199": {Conv: "raw48", Name: "UDMA_CRC_Error_Count"},
					"200": {Conv: "raw48", Name: "Multi_Zone_Error_Rate"},
					"201": {Conv: "raw48", Name: "Soft_Read_Error_Rate"},
					"202": {Conv: "raw48", Name: "Data_Address_Mark_Errs"},
					"203": {Conv: "raw48", Name: "Run_Out_Cancel"},
					"204": {Conv: "raw48", Name: "Soft_ECC_Correction"},
					"205": {Conv: "raw48", Name: "Thermal_Asperity_Rate"},
					"206": {Conv: "raw48", Name: "Flying_Height"},
					"207": {Conv: "raw48", Name: "Spin_High_Current"},
					"208": {Conv: "raw48", Name: "Spin_Buzz"},
					"209": {Conv: "raw48", Name: "Offline_Seek_Performnce"},
					"220": {Conv: "raw48", Name: "Disk_Shift"},
					"221": {Conv: "raw48", Name: "G-Sense_Error_Rate"},
					"222": {Conv: "raw48", Name: "Loaded_Hours"},
					"223": {Conv: "raw48", Name: "Load_Retry_Count"},
					"224": {Conv: "raw48", Name: "Load_Friction"},
					"225": {Conv: "raw48", Name: "Load_Cycle_Count"},
					"226": {Conv: "raw48", Name: "Load-in_Time"},
					"227": {Conv: "raw48", Name: "Torq-amp_Count"},
					"228": {Conv: "raw48", Name: "Power-off_Retract_Count"},
					"230": {Conv: "raw48", Name: "Head_Amplitude"},
					"231": {Conv: "raw48", Name: "Temperature_Celsius"},
					"232": {Conv: "raw48", Name: "Available_Reservd_Space"},
					"233": {Conv: "raw48", Name: "Media_Wearout_Indicator"},
					"240": {Conv: "raw24(raw8)", Name: "Head_Flying_Hours"},
					"241": {Conv: "raw48", Name: "Total_LBAs_Written"},
					"242": {Conv: "raw48", Name: "Total_LBAs_Read"},
					"250": {Conv: "raw48", Name: "Read_Error_Retry_Rate"},
					"254": {Conv: "raw48", Name: "Free_Fall_Sensor"},
				},
			},
		},
	}
}

// LookupDrive returns the most appropriate DriveModel for a given ATA IDENTIFY value.
func (db *DriveDb) LookupDrive(ident []byte) DriveModel {
	var model DriveModel

	for _, d := range db.Drives {
		// Skip placeholder entry
		if strings.HasPrefix(d.Family, "$Id") {
			continue
		}

		if d.Family == "DEFAULT" {
			model = d
			continue
		}

		if d.CompiledRegexp.Match(ident) {
			model.Family = d.Family
			model.ModelRegex = d.ModelRegex
			model.FirmwareRegex = d.FirmwareRegex
			model.WarningMsg = d.WarningMsg
			model.CompiledRegexp = d.CompiledRegexp

			for id, p := range d.Presets {
				if _, exists := model.Presets[id]; exists {
					// Some drives override the conv but don't specify a name, so copy it from default
					if p.Name == "" {
						p.Name = model.Presets[id].Name
					}
				}
				model.Presets[id] = AttrConv{Name: p.Name, Conv: p.Conv}
			}

			break
		}
	}

	return model
}

// OpenDriveDb opens a YAML-formatted drive database, unmarshalls it, and returns a DriveDb.
func OpenDriveDb(dbfile string) (DriveDb, error) {
	var db DriveDb

	f, err := os.Open(dbfile)
	if err != nil {
		return db, nil
	}

	defer f.Close()
	dec := yaml.NewDecoder(f)

	if err := dec.Decode(&db); err != nil {
		return db, err
	}

	for i, d := range db.Drives {
		db.Drives[i].CompiledRegexp, _ = regexp.Compile(d.ModelRegex)
	}

	return db, nil
}
