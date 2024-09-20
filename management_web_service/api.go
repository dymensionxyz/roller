package management_web_service

import (
	"bufio"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"cosmossdk.io/errors"

	"github.com/gin-gonic/gin"
)

type RequestEIbcClientStart struct {
	Denom         string `json:"denom"`
	MinFeePercent string `json:"min_fee_percent"`
}

func HandleApiEIbcClientStart(c *gin.Context) {
	w := wrapGin(c)
	// cfg := w.Config()

	// read request

	var req RequestEIbcClientStart

	if err := c.ShouldBindJSON(&req); err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("failed to parse request: %v", err)).
			SendResponse()
		return
	}

	// validate request, to prevent XSS & remote execution

	req.Denom = strings.TrimSpace(req.Denom)
	req.MinFeePercent = strings.TrimSpace(req.MinFeePercent)
	if req.Denom == "" || req.MinFeePercent == "" {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("all fields are required").
			SendResponse()
		return
	}

	if regexp.MustCompile(`^ibc/[A-Z\d]{64}$`).MatchString(req.Denom) {
		// IBC denom, allow
	} else if regexp.MustCompile(`^[a-zA-Z\d]+$`).MatchString(req.Denom) {
		// maybe native denom, still allow
	} else {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("invalid denom format").
			SendResponse()
		return
	}
	denom := req.Denom

	if strings.Contains(req.MinFeePercent, ",") {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("comma ',' is not allowed, only '.' is allowed as decimal point").
			SendResponse()
		return
	}
	if regexp.MustCompile(`^\d+(\.\d*)?$`).MatchString(req.MinFeePercent) {
		// allow
	} else {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("invalid min fee percent format").
			SendResponse()
		return
	}
	minFeePercent, err := strconv.ParseFloat(req.MinFeePercent, 64)
	if err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("invalid min fee percent value").
			SendResponse()
		return
	}
	if minFeePercent > 100 || minFeePercent <= 0 {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("min fee percent must be greater than 0 or lower/equals to 100").
			SendResponse()
		return
	}
	minFeePercent = float64(int(minFeePercent*100)) / 100 // rounding

	// validate process

	if err := cache.CanStartEIbcClientProcessID(); err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("can not start eIBC client: %v", err)).
			SendResponse()
	}

	if pid := cache.GetEIbcClientProcessID(); pid > 0 {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("eIBC client is already running at PID: %d", pid)).
			SendResponse()
		return
	}

	anyEIbcClient, err := AnyEIbcClient() // extra check
	if err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("failed to check running eIBC client: %v", err)).
			SendResponse()
		return
	}
	if anyEIbcClient {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("eIBC client is already running").
			SendResponse()
		return
	}

	// start process

	var cmd *exec.Cmd
	if UseSimulation {
		cmd = exec.Command(EIbcClientBinaryName, SimulationStartCommand)
	} else {
		// TODO: setup config based on the request denom and min fee percent
		panic("not implemented")
	}
	if err := startEIbcClientAndCacheOutput(cmd); err != nil {
		msg := fmt.Sprintf("failed to start eIBC client: %v", err)
		cache.AppendEIbcClientLog(msg)
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(msg).
			SendResponse()
		return
	}

	cache.SetEIbcClientProcess(cmd.Process)
	cache.SetEIbcClientArgs(denom, minFeePercent)
	w.PrepareDefaultSuccessResponse(nil).SendResponse()
}

func HandleApiEIbcClientStop(c *gin.Context) {
	w := wrapGin(c)
	// cfg := w.Config()

	if pid := cache.GetEIbcClientProcessID(); pid == 0 {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("eIBC client is not running").
			SendResponse()
		return
	}

	anyEIbcClient, err := AnyEIbcClient() // extra check
	if err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("failed to check running eIBC client: %v", err)).
			SendResponse()
		return
	}

	if !anyEIbcClient {
		cache.SetEIbcClientProcess(nil)
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult("eIBC client is not running").
			SendResponse()
		return
	}

	if err := cache.GetEIbcClientProcess().Kill(); err != nil {
		w.PrepareDefaultErrorResponse().
			WithHttpStatusCode(http.StatusOK).
			WithResult(fmt.Sprintf("failed to kill eIBC client: %v", err)).
			SendResponse()
		return
	}
	cache.SetEIbcClientProcess(nil)

	w.PrepareDefaultSuccessResponse(nil).SendResponse()
}

func HandleApiEIbcClientStatus(c *gin.Context) {
	w := wrapGin(c)
	// cfg := w.Config()

	type Status struct {
		Running       bool    `json:"running"`
		PID           int     `json:"pid"`
		Denom         string  `json:"denom"`
		MinFeePercent float64 `json:"min_fee_percent"`
	}

	proc := cache.GetEIbcClientProcess()
	var pid int
	if proc != nil {
		pid = proc.Pid
	}

	denom, minFeePercent := cache.GetEIbcClientArgs()

	w.PrepareDefaultSuccessResponse(Status{
		Running:       proc != nil,
		PID:           pid,
		Denom:         denom,
		MinFeePercent: minFeePercent,
	}).SendResponse()
}

func startEIbcClientAndCacheOutput(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to pipe stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "failed to pipe stderr")
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				cache.AppendEIbcClientLog(fmt.Sprintf("panic stdout reader: %v", r))
			}
			_ = stdout.Close()
		}()
		for stdoutScanner.Scan() {
			cache.AppendEIbcClientLog("O: " + stdoutScanner.Text())
		}
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				cache.AppendEIbcClientLog(fmt.Sprintf("panic stdout reader: %v", r))
			}
			_ = stderr.Close()
		}()
		for stderrScanner.Scan() {
			cache.AppendEIbcClientLog("E: " + stderrScanner.Text())
		}
	}()

	return nil
}
