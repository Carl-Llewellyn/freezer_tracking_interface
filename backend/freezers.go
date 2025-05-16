package freezerinv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gitlab.com/UrsusArcTech/logger"
)

type FreezerDB struct {
	Id                      int        `json:"id"`
	FreezerLocationId       int        `json:"freezer_location_id"`
	LastCalibrated          *time.Time `json:"last_calibrated"`
	Name                    string     `json:"name"`
	Model                   string     `json:"model"`
	Comments                *string    `json:"comments"`
	CurrentHoldingTempC     *int       `json:"current_holding_temp_c"`
	ManualProjectsContained string     `json:"manual_projects_contained"`
}

type FreezerExtr struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func GetAllFreezers(w http.ResponseWriter, r *http.Request) {

	query := "select id, name from mgl_freezer_inventory.freezer"

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var results []FreezerExtr

	for rows.Next() {
		var freezer FreezerExtr

		err := rows.Scan(
			&freezer.Id,
			&freezer.Name,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		results = append(results, freezer)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

func GetFreezersInRoom(w http.ResponseWriter, r *http.Request) {
	//there shouldn't be a join here but being lazy
	query := "SELECT f.id, freezer_location_id, last_calibrated, name, model, comments, current_holding_temp_c, manual_projects_contained from mgl_freezer_inventory.freezer f join mgl_freezer_inventory.freezer_locations fl on fl.id = f.freezer_location_id WHERE fl.id = $1"

	args := []interface{}{}

	roomId := r.URL.Query().Get("roomid")

	errRooms := errors.New("No room ID specified for freezer")
	if roomId == "" {
		logger.LogError("No room ID specified for freezer")
		http.Error(w, errRooms.Error(), 500)
		return
	}

	args = append(args, roomId)

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var results []FreezerDB

	for rows.Next() {
		var freezer FreezerDB

		err := rows.Scan(
			&freezer.Id,
			&freezer.FreezerLocationId,
			&freezer.LastCalibrated,
			&freezer.Name,
			&freezer.Model,
			&freezer.Comments,
			&freezer.CurrentHoldingTempC,
			&freezer.ManualProjectsContained,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		results = append(results, freezer)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}
