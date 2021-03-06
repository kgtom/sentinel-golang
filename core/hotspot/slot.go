package hotspot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
)

const (
	RuleCheckSlotName = "sentinel-core-hotspot-rule-check-slot"
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

// matchArg matches the arg from args based on TrafficShapingController
// return nil if match failed.
func matchArg(tc TrafficShapingController, args []interface{}) interface{} {
	if tc == nil {
		return nil
	}
	idx := tc.BoundParamIndex()
	if idx < 0 {
		idx = len(args) + idx
	}
	if idx < 0 {
		if logging.DebugEnabled() {
			logging.Debug("The param index of hotspot traffic shaping controller is invalid", "args", args, "paramIndex", tc.BoundParamIndex())
		}
		return nil
	}
	if idx >= len(args) {
		if logging.DebugEnabled() {
			logging.Debug("The argument in index doesn't exist", "args", args, "paramIndex", tc.BoundParamIndex())
		}
		return nil
	}
	arg := args[idx]
	if arg == nil {
		return nil
	}
	switch arg.(type) {
	case bool:
	case float32:
		n32 := arg.(float32)
		n64, err := strconv.ParseFloat(fmt.Sprintf("%.5f", n32), 64)
		if err != nil {
			return nil
		}
		arg = n64
	case float64:
		n64 := arg.(float64)
		n64, err := strconv.ParseFloat(fmt.Sprintf("%.5f", n64), 64)
		if err != nil {
			return nil
		}
		arg = n64
	case int:
		arg = arg.(int)
	case int8:
		n := arg.(int8)
		arg = int(n)
	case int16:
		n := arg.(int16)
		arg = int(n)
	case int32:
		n := arg.(int32)
		arg = int(n)
	case int64:
		n := arg.(int64)
		arg = int(n)
	case uint:
		n := arg.(uint)
		arg = int(n)
	case uint8:
		n := arg.(uint8)
		arg = int(n)
	case uint16:
		n := arg.(uint16)
		arg = int(n)
	case uint32:
		n := arg.(uint32)
		arg = int(n)
	case uint64:
		n := arg.(uint64)
		arg = int(n)
	case string:
	default:
	}
	return arg
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	args := ctx.Input.Args
	batch := int64(ctx.Input.BatchCount)

	result := ctx.RuleCheckResult
	tcs := getTrafficControllersFor(res)
	if len(tcs) == 0 {
		return result
	}

	for _, tc := range tcs {
		arg := matchArg(tc, args)
		if arg == nil {
			continue
		}
		r := canPassCheck(tc, arg, batch)
		if r == nil {
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if waitMs := r.WaitMs(); waitMs > 0 {
				// Handle waiting action.
				time.Sleep(time.Duration(waitMs) * time.Millisecond)
			}
			continue
		}
	}
	return result
}

func canPassCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return canPassLocalCheck(tc, arg, batch)
}

func canPassLocalCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return tc.PerformChecking(arg, batch)
}
