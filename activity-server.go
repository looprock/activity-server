package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	"github.com/rs/zerolog"
)

// TODO: N (and probably minutes) should be read from env vars or configs (whatever works best for systemd)
// number of results to hold in the queue / number of minutes to keep results for
var Port = "18800"
var N = 120
var a = make([]int, N)
var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func sum(array []int) int {
	result := 0
	for _, v := range array {
		result += v
	}
	return result
}

// remove the oldest result from the stack and append a new one
func logSessions() {
	a = append(a[:0], a[1:]...)
	a = append(a, findActiveSshSessions())
	logger.Debug().Int("Result:", sum(a))
}

// run a background gorouting every N seconds
func executeCronJob() {
	gocron.Every(60).Second().Do(logSessions)
	<-gocron.Start()
}

func main() {
	// fill the stack with 1s so instance isn't terminated instantly
	for i := 0; i < N; i++ {
		a[i] = 1
	}
	// update the stack with the newest result
	logSessions()
	// start the background thread
	go executeCronJob()
	// create an endpoint
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/sessions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"current_active_sessions":   findActiveSshSessions(),
			"total_active_sessions_2hr": sum(a),
		})
	})
	r.Run(":" + Port)
}

func findActiveSshSessions() int {
	// find all established ssh sessions. Treating this as 'activity'
	cmd := exec.Command("ss", "-a", "state", "established", "(", "sport", "=", ":ssh", ")")
	stdout, err := cmd.Output()
	if err != nil {
		logger.Error().Msg(err.Error())
	}
	s := string(stdout)
	// count the sessions (plus the header and extra newline)
	c := 0
	for _, line := range strings.Split(s, "\n") {
		c = c + 1
		// idk why the compiler is complaining this isn't used...
		logger.Debug().Str("Found line: %s", line)
	}
	// remove the column headers and an extra newline from count
	c = c - 2
	return c
}
