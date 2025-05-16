package freezerinv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/UrsusArcTech/logger"
	"gitlab.com/mgl-database/mgl-go/object_processing_x/object_processing_edna"
)

type EdnaLink struct {
	EdnaId      *int   `json:"edna_id"`
	EnteredName string `json:"entered_name"`
	BoxId       int    `json:"box_id"`
}

type EdnaToLocation struct {
	Shelf        int    `json:"shelf"`
	EdnaName     string `json:"edna_name"`
	BoxName      string `json:"box_name"`
	FreezerName  string `json:"freezer_name"`
	FreezerModel string `json:"freezer_model"`
	Lab          string `json:"lab"`
	Floor        string `json:"floor"`
}

func EdnaLinkByBox(w http.ResponseWriter, r *http.Request) {
	query := "select edna_id, entered_name, box_id from mgl_freezer_inventory.mgl_edna_box_link where box_id = $1"
	args := []interface{}{}

	boxId := r.URL.Query().Get("boxid")

	errEdna := errors.New("No box ID specified for eDNA")
	if boxId == "" {
		logger.LogError("No box ID specified for eDNA")
		http.Error(w, errEdna.Error(), 500)
		return
	}

	args = append(args, boxId)

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var results []EdnaLink

	for rows.Next() {
		var ednaLink EdnaLink

		err := rows.Scan(
			&ednaLink.EdnaId,
			&ednaLink.EnteredName,
			&ednaLink.BoxId,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		results = append(results, ednaLink)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

// check if this eDNA exists. This will search basic and unique IDs, etc.
// really this needs to propagate the not unique error from mgl-pg - this will come later as it needs an edit for the mgl-pg repo
func CheckEdnaExists(ednaName string) (ednaId int, ednaNameFound string) {

	//this will search basic IDs, unique IDs, etc.
	ednaFoundID, ednaFoundName := object_processing_edna.GetEDNAID(ednaName)

	if ednaFoundID == -1 && ednaFoundName == "" {
		logger.LogError(ednaName, " - eDNA not linked. May be due to being not a unique ID. Try using unique ID.")
		return ednaFoundID, ednaFoundName
	}

	logger.LogMessage("Using: ", ednaFoundName, " from ", ednaName, " found eDNA sample: ", ednaFoundName, " with ID: ", ednaFoundID)
	return ednaFoundID, ednaFoundName
}

func CheckEdnaAlreadyInABox(w http.ResponseWriter, r *http.Request) {

	ednaName := r.URL.Query().Get("ednaid")

	if ednaName == "" {
		logger.LogError("Empty eDNA name for link check")
		http.Error(w, "Empty eDNA name for link check", http.StatusInternalServerError)
		return
	}

	query := "SELECT shelf, ebl.entered_name as edna_name, b.name as box_name, f.name as freezer_name, f.model as freezer_model, fl.lab, fl.floor FROM mgl_freezer_inventory.mgl_edna_box_link ebl join mgl_freezer_inventory.boxes b on ebl.box_id = b.id join mgl_freezer_inventory.freezer f on b.freezer_id = f.id join mgl_freezer_inventory.freezer_locations fl on fl.id = f.freezer_location_id WHERE entered_name = $1"
	args := []interface{}{ednaName}

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		logger.LogError("eDNA box check err: ", err.Error())
		http.Error(w, "eDNA box check err: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []EdnaToLocation

	for rows.Next() {
		var ednaLink EdnaToLocation

		err := rows.Scan(
			&ednaLink.Shelf,
			&ednaLink.EdnaName,
			&ednaLink.BoxName,
			&ednaLink.FreezerName,
			&ednaLink.FreezerModel,
			&ednaLink.Lab,
			&ednaLink.Floor,
		)

		if err != nil {
			logger.LogError("eDNA box check err1: ", err.Error())
			http.Error(w, "eDNA box check err:1 "+err.Error(), http.StatusInternalServerError)
			return
		}

		results = append(results, ednaLink)

	}

	w.WriteHeader(http.StatusOK)
	if len(results) > 0 {
		locStr := fmt.Sprintf("Box: %s, Freezer: %s (%s), Shelf: %d, Floor: %s, Lab: %s", results[0].BoxName, results[0].FreezerName, results[0].FreezerModel, results[0].Shelf, results[0].Floor, results[0].Lab)
		w.Write([]byte("eDNA already exists in: " + locStr))
		return
	}

	w.Write([]byte("0"))
}

// InsertBox handles HTTP POST requests to create a new box
func InsertEdnaLink(w http.ResponseWriter, r *http.Request) {
	enteredName := r.URL.Query().Get("enteredname")
	boxId := r.URL.Query().Get("boxid")

	if enteredName == "" || boxId == "" {
		logger.LogError("Missing required fields: enteredname, and boxid")
		http.Error(w, "Missing required fields: enteredname, and boxid", http.StatusBadRequest)
		return
	}

	ednaDbId, ednaDbName := CheckEdnaExists(enteredName)

	query := ""
	var args []interface{}
	if ednaDbId != -1 {
		query = "INSERT INTO mgl_freezer_inventory.mgl_edna_box_link(edna_id, box_id, entered_name) VALUES ($1, $2, $3)"
		args = []interface{}{ednaDbId, boxId, enteredName}
	} else {
		query = "INSERT INTO mgl_freezer_inventory.mgl_edna_box_link(box_id, entered_name) VALUES ($1, $2)"
		args = []interface{}{boxId, enteredName}
	}

	//logger.LogMessage(query)

	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if ednaDbId == -1 {
		w.Write([]byte("eDNA ID was not found in database or is not unique (unique IDs can be used). The record is recorded bt not linked to the eDNA table."))
	} else if ednaDbName != enteredName {
		w.Write([]byte("eDNA ID matched to: " + ednaDbName + ". This will be used."))
	}
}

// UpdateBox handles HTTP PUT requests to update a box's FreezerID
func UpdateEdnaLink(w http.ResponseWriter, r *http.Request) {
	boxId := r.URL.Query().Get("boxid")
	enteredname := r.URL.Query().Get("enteredname")
	newenteredname := r.URL.Query().Get("newenteredname")

	if boxId == "" || enteredname == "" {
		logger.LogError("Missing required fields: enteredname, and boxid")
		http.Error(w, "Missing required fields: enteredfishname, and boxid", http.StatusBadRequest)
		return
	}

	query := ""
	var args []interface{}

	if newenteredname != "" {
		query = "UPDATE mgl_freezer_inventory.mgl_edna_box_link set entered_name = $1, box_id = $2 WHERE entered_name = $3"
		args = []interface{}{newenteredname, boxId, enteredname}
	} else {
		query = "UPDATE mgl_freezer_inventory.mgl_edna_box_link set box_id = $1 WHERE entered_name = $2"
		args = []interface{}{boxId, enteredname}
	}
	result, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.LogError("No rows affected - fish link not found")
		http.Error(w, "eDNA link not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteEdnaLink(w http.ResponseWriter, r *http.Request) {
	enteredname := r.URL.Query().Get("enteredname")

	if enteredname == "" {
		logger.LogError("Missing required fields: enteredname")
		http.Error(w, "Missing required fields: enteredname", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM mgl_freezer_inventory.mgl_edna_box_link WHERE entered_name = $1"
	args := []interface{}{enteredname}

	//logger.LogMessage(query)

	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
