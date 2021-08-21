package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/tsawler/vigilate/internal/helpers"
	"github.com/tsawler/vigilate/internal/models"
)

type ByHost []models.Schedule

// Len is used to sort by Host
func (a ByHost) Len() int { return len(a) }

// Less is used to sort by Host
func (a ByHost) Less(i, j int) bool { return a[i].Host < a[j].Host }

// Swap is used to sort by Host
func (a ByHost) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	// look at schedule running in background from application wide config
	for k, v := range repo.App.MonitorMap {
		var item models.Schedule

		// get key and value from map
		item.ID = k
		item.EntryID = v
		item.Entry = repo.App.Scheduler.Entry(v) // get the entry
		hs, err := repo.DB.GetHostServiceByID(k) // get host service
		if err != nil {
			log.Println(err)
			return
		}
		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
		item.LastRunFromHS = hs.LastCheck
		item.Host = hs.HostName
		item.Service = hs.Service.ServiceName

		items = append(items, item)
	}

	// sort the slice by host
	sort.Sort(ByHost(items))

	// pass data to template
	data := make(jet.VarMap)

	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
