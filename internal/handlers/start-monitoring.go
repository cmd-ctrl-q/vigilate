package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type job struct {
	HostServiceID int
}

// Run runs a schedule check
func (j job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

// StartMonitoring starts the monitoring process
func (repo *DBRepo) StartMonitoring() {
	// monitor jobs if set to 1
	if app.PreferenceMap["monitoring_live"] == "1" {
		// trigger a message to broadcast to all clients that the app is starting to monitor.
		// sends a message to every client thats connected to the service/website.
		// payload to be sent to all connected clients via websockets (using pusher / ipe)
		data := make(map[string]string)
		data["message"] = "Monitoring is starting..."
		err := app.WsClient.Trigger("public-channel", "app-starting", data)
		if err != nil {
			log.Println(err)
		}

		// get all of the services that need to be monitored
		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		// range through the services
		log.Println("Length of services to monitor is", len(servicesToMonitor))

		// range through services
		for _, x := range servicesToMonitor {
			log.Println("Services to monitor on", x.HostName, "is", x.Service.ServiceName)

			// get the schedule unit and number
			var sch string
			if x.ScheduleUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)
			}

			// create a job
			var j job
			j.HostServiceID = x.ID

			// schedule a job
			scheduleID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}

			// save id of job to start/stop it.
			app.MonitorMap[x.ID] = scheduleID

			// broadcast over websockets that the service is scheduled
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(x.ID)
			yearOne := time.Date(0001, 11, 17, 20, 34, 58, 65138737, time.UTC)

			// if monitoring is stopped after running,
			// if the scheduler contains a job after year one.
			if app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.After(yearOne) {
				data["next_run"] = app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				// on app start, scheduler will always be pending
				data["next_run"] = "Pending..."
			}
			payload["host"] = x.HostName
			payload["service"] = x.Service.ServiceName
			// if date is after year one, then the scheduler has ran a scheduled check in the past
			if x.LastCheck.After(yearOne) {
				payload["last_run"] = x.LastCheck.Format("2006-01-02 3:04:05 PM")
			} else {
				// its never run a scheduled check
				payload["last_run"] = "Pending..."
			}

			payload["schedule"] = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)

			// trigger web client again
			err = app.WsClient.Trigger("public-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
