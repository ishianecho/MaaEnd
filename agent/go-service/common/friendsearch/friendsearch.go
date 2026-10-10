package friendsearch

import (
	"encoding/json"
	"regexp"
	"strings"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

const componentName = "FriendSearch"

var nameSplitRegex = regexp.MustCompile(`[,，;；、\s]+`)

var state friendSearchState

type friendSearchState struct {
	queue   []string
	total   int
	current string
}

// InitAction 解析搜索名单并重建队列，由调用方在任务启动时执行一次。
type InitAction struct{}

// InputAction 弹出队列中下一个关键词并输入当前聚焦的搜索框。
type InputAction struct{}

// ExhaustedRecognition 在搜索队列耗尽时命中。
type ExhaustedRecognition struct{}

var (
	_ maa.CustomActionRunner      = &InitAction{}
	_ maa.CustomActionRunner      = &InputAction{}
	_ maa.CustomRecognitionRunner = &ExhaustedRecognition{}
)

type initParam struct {
	Names string `json:"names"`
}

func (a *InitAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	var params initParam
	if err := json.Unmarshal([]byte(arg.CustomActionParam), &params); err != nil {
		log.Error().
			Err(err).
			Str("component", componentName+".InitAction").
			Msg("failed to parse params")
		return false
	}

	parts := nameSplitRegex.Split(params.Names, -1)
	seen := make(map[string]struct{}, len(parts))
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	state = friendSearchState{
		queue: names,
		total: len(names),
	}
	if len(names) == 0 {
		log.Warn().
			Str("component", componentName+".InitAction").
			Msg("name list is empty")
		return true
	}
	log.Info().
		Int("total", state.total).
		Str("component", componentName+".InitAction").
		Msg("search queue initialized")
	return true
}

func (a *InputAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	if len(state.queue) == 0 {
		log.Error().
			Str("component", componentName+".InputAction").
			Msg("name queue is empty")
		return false
	}
	tasker := ctx.GetTasker()
	if tasker == nil {
		return false
	}
	controller := tasker.GetController()
	if controller == nil {
		return false
	}

	name := state.queue[0]
	state.queue = state.queue[1:]
	state.current = name
	controller.PostInputText(name).Wait()
	log.Debug().
		Str("name", name).
		Int("remaining", len(state.queue)).
		Int("total", state.total).
		Str("component", componentName+".InputAction").
		Msg("input search name")
	return true
}

func (r *ExhaustedRecognition) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	if len(state.queue) > 0 {
		return nil, false
	}
	return &maa.CustomRecognitionResult{
		Box:    arg.Roi,
		Detail: "queue empty",
	}, true
}
