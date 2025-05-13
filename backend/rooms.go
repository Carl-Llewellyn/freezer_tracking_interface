package freezerinv

import (
	"gitlab.com/mgl-database/mgl-go"
	"net/http"
	"strconv"
	"gitlab.com/UrsusArcTech/logger"
	"context"
	"encoding/json"
)

type FreezerRoom struct{
	Lab string `json:"lab"`
	Floor string `json:"floor"`
	Id int `json:"id"`
}

func GetFreezerRooms(w http.ResponseWriter, r *http.Request){
	mglgo.Help()
	query := "SELECT lab, floor, id FROM mgl_freezer_inventory.freezer_locations"
	args := []interface{}{}
	where := ""
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 200
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}


	searchQuery := r.URL.Query().Get("search")
	if searchQuery != ""{
		where = " WHERE lab ILIKE $1 OR floor ILIKE $1"
		args = append(args, "%"+searchQuery+"%")
	}

	query += where + " ORDER BY floor LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer rows.Close()

	var results []FreezerRoom

	for rows.Next(){
		var freezerRoom FreezerRoom
		lab := ""
		floor := ""
		id := 0

		err := rows.Scan(
			&lab, &floor, &id,
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		
		freezerRoom.Lab = lab
		freezerRoom.Floor = floor
		freezerRoom.Id = id
		results = append(results, freezerRoom)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
