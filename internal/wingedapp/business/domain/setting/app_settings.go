package setting

import (
	"context"
	"fmt"
	"strconv"
	"wingedapp/pgtester/internal/wingedapp/sysparam"
)

type dataProcessor struct {
	dataType     string
	intSetter    func(a *sysparam.Settings, v int)
	stringSetter func(a *sysparam.Settings, v string)
	floatSetter  func(a *sysparam.Settings, v float64)
}

func intProc(setter func(a *sysparam.Settings, v int)) *dataProcessor {
	return &dataProcessor{
		dataType:  "int",
		intSetter: setter,
	}
}

func floatProc(setter func(a *sysparam.Settings, v float64)) *dataProcessor {
	return &dataProcessor{
		dataType:    "float64",
		floatSetter: setter,
	}
}

// dataProcessors maps sysparam keys to their respective data processors.
func dataProcessors() map[string]*dataProcessor {
	return map[string]*dataProcessor{
		// for transcript files
		sysparam.SettingUpAgentTranscriptLoopIntervalSecs: intProc(func(a *sysparam.Settings, v int) {
			a.SettingUpAgentTranscriptLoopIntervalSecs = v
		}),
		sysparam.SettingUpAgentTranscriptMaxLoops: intProc(func(a *sysparam.Settings, v int) {
			a.SettingUpAgentTranscriptMaxLoops = v
		}),

		// for audio files
		sysparam.SettingUpAgentAudioFilesLoopIntervalSecs: intProc(func(a *sysparam.Settings, v int) {
			a.SettingUpAgentAudioFilesLoopIntervalSecs = v
		}),
		sysparam.SettingUpAgentAudioFilesMaxLoops: intProc(func(a *sysparam.Settings, v int) {
			a.SettingUpAgentAudioFilesMaxLoops = v
		}),

		// for user invite codes
		sysparam.UserInviteCodeMaxUsage: intProc(func(a *sysparam.Settings, v int) {
			a.UserInviteCodeMaxUsage = v
		}),

		// for broken audio duration threshold
		sysparam.BrokenAudioDurationThresholdSecs: intProc(func(a *sysparam.Settings, v int) {
			a.BrokenAudioDurationThresholdSecs = v
		}),

		// setting up agent idle waiting time
		sysparam.SettingUpAgentIdleWaitingTimeSecs: intProc(func(a *sysparam.Settings, v int) {
			a.SettingUpAgentIdleWaitingTime = v
		}),

		// invite code expiry days
		sysparam.InviteExpiryDays: intProc(func(a *sysparam.Settings, v int) {
			a.InviteExpiryDays = v
		}),

		// wings economy
		sysparam.WingsEconIncrementThreshSendMessage: intProc(func(a *sysparam.Settings, v int) {
			a.WingsEconIncrementThreshSendMessage = v
		}),
		sysparam.WingsEconIncrementRewardSendMessage: intProc(func(a *sysparam.Settings, v int) {
			a.WingsEconIncrementRewardSendMessage = v
		}),

		// wings economy - winged plus weekly payments
		sysparam.WingsEconWingedPlusWeeklyPaymentAmount: floatProc(func(a *sysparam.Settings, v float64) {
			a.WingsEconWingedPlusWeeklyPaymentAmount = v
		}),
		sysparam.WingsEconWingedPlusWeeklyWingsAmount: intProc(func(a *sysparam.Settings, v int) {
			a.WingsEconWingedPlusWeeklyWingsAmount = v
		}),
	}
}

// Settings retrieves application settings from system parameters.
// sysparam.Settings are live configuration settings that can be changed in real-time.
func (p *Business) Settings(ctx context.Context) (*sysparam.Settings, error) {
	sysParams, err := p.storer.SysParams(ctx, p.trans.DB())
	if err != nil {
		return nil, fmt.Errorf("storer personalInfo: %w", err)
	}

	processors := dataProcessors()

	// loop and populate Settings
	settings := &sysparam.Settings{}
	var processor_ *dataProcessor
	for _, param := range sysParams {
		processor_ = processors[param.Key]
		if processor_ == nil {
			return nil, fmt.Errorf("unknown sys param: %s", param.Key)
		}

		// validate setters - only be 1 setter type
		setters := []bool{
			processor_.intSetter != nil,
			processor_.stringSetter != nil,
			processor_.floatSetter != nil,
		}
		setterCount := 0
		for _, setter := range setters {
			if setter {
				setterCount++
			}
		}
		if setterCount != 1 {
			return nil, fmt.Errorf("only one setter must be set for param: %s, with count: %v", param.Key, setterCount)
		}

		// set the value of Settings based on data type
		switch processor_.dataType {
		case "int":
			var intVal int
			intVal, err = strconv.Atoi(param.Val)
			if err != nil {
				return nil, fmt.Errorf("convert int sys param %s: %w", param.Key, err)
			}
			processor_.intSetter(settings, intVal)
		case "string":
			processor_.stringSetter(settings, param.Val)
		case "float64":
			floatVal, err := strconv.ParseFloat(param.Val, 64)
			if err != nil {
				return nil, fmt.Errorf("convert float64 sys param %s: %w", param.Key, err)
			}
			processor_.floatSetter(settings, floatVal)
		default:
			return nil, fmt.Errorf("unknown data type for param %s: %s", param.Key, processor_.dataType)
		}
	}

	return settings, nil
}
