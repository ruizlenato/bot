package bot

import (
	"regexp"
	"strings"

	"github.com/go-telegram/bot/models"
)

type HandlerType int

const (
	HandlerTypeMessageText HandlerType = iota
	HandlerTypeCommand
	HandlerTypeCallbackQueryData
	HandlerTypeCallbackQueryGameShortName
	HandlerTypePhotoCaption
)

type handler struct {
	id          string
	handlerType HandlerType
	handler     HandlerFunc

	pattern   any
	re        *regexp.Regexp
	matchFunc MatchFunc
}

func (h handler) match(update *models.Update) bool {
	if h.matchFunc != nil {
		return h.matchFunc(update)
	}

	var data string
	var entities []models.MessageEntity

	switch h.handlerType {
	case HandlerTypeMessageText, HandlerTypeCommand:
		if update.Message == nil {
			return false
		}
		data = update.Message.Text
		entities = update.Message.Entities
	case HandlerTypeCallbackQueryData:
		if update.CallbackQuery == nil {
			return false
		}
		data = update.CallbackQuery.Data
	case HandlerTypeCallbackQueryGameShortName:
		if update.CallbackQuery == nil {
			return false
		}
		data = update.CallbackQuery.GameShortName
	case HandlerTypePhotoCaption:
		if update.Message == nil {
			return false
		}
		data = update.Message.Caption
		entities = update.Message.CaptionEntities
	}

	switch pattern := h.pattern.(type) {
	case string:
		if h.handlerType == HandlerTypeCommand {
			for _, e := range entities {
				if e.Type == models.MessageEntityTypeBotCommand {
					if e.Offset == 0 && data[e.Offset+1:e.Offset+e.Length] == pattern {
						return true
					}
				}
			}
			return false
		}
		p := regexp.MustCompile(pattern)
		return p.MatchString(data) || strings.Contains(data, pattern)
	case *regexp.Regexp:
		return pattern.MatchString(data)
	default:
		return false
	}
}

func (b *Bot) RegisterHandler(handlerType HandlerType, pattern any, f HandlerFunc, m ...Middleware) string {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		id:          id,
		handlerType: handlerType,
		pattern:     pattern,
		handler:     applyMiddlewares(f, m...),
	}

	b.handlers = append(b.handlers, h)

	return id
}

func (b *Bot) RegisterHandlerMatchFunc(matchFunc MatchFunc, f HandlerFunc, m ...Middleware) string {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	id := RandomString(16)

	h := handler{
		id:        id,
		matchFunc: matchFunc,
		handler:   applyMiddlewares(f, m...),
	}

	b.handlers = append(b.handlers, h)

	return id
}

func (b *Bot) UnregisterHandler(id string) {
	b.handlersMx.Lock()
	defer b.handlersMx.Unlock()

	for i, h := range b.handlers {
		if h.id == id {
			b.handlers = append(b.handlers[:i], b.handlers[i+1:]...)
			return
		}
	}
}
